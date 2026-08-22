package namespaces

import (
	"errors"
	"strings"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func mapNamespaceError(err error) error {
	if err == nil {
		return nil
	}

	var invalidConfigErr *intkube.ErrInvalidKubeconfig
	if errors.As(err, &invalidConfigErr) {
		return apperrors.ErrNamespaceInvalidKubeconfig
	}
	if strings.Contains(strings.ToLower(err.Error()), "kubeconfig") {
		return apperrors.ErrNamespaceInvalidKubeconfig
	}

	var connectionErr *intkube.ErrConnectionFailed
	if errors.As(err, &connectionErr) {
		return apperrors.ErrNamespaceClusterUnavailable
	}

	var authErr *intkube.ErrAuthenticationFailed
	if errors.As(err, &authErr) {
		return apperrors.ErrNamespaceForbidden
	}
	var authorizationErr *intkube.ErrAuthorizationFailed
	if errors.As(err, &authorizationErr) {
		return apperrors.ErrNamespaceForbidden
	}

	if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
		return apperrors.ErrNamespaceForbidden
	}
	if apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) {
		return apperrors.ErrNamespaceTimeout
	}

	return err
}
