package nodes

import (
	"errors"
	"strings"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func mapNodeError(err error) error {
	if err == nil {
		return nil
	}

	var invalidConfigErr *intkube.ErrInvalidKubeconfig
	if errors.As(err, &invalidConfigErr) {
		return apperrors.ErrNodeInvalidKubeconfig
	}
	if strings.Contains(strings.ToLower(err.Error()), "kubeconfig") {
		return apperrors.ErrNodeInvalidKubeconfig
	}

	var connectionErr *intkube.ErrConnectionFailed
	if errors.As(err, &connectionErr) {
		return apperrors.ErrNodeClusterUnavailable
	}

	var authErr *intkube.ErrAuthenticationFailed
	if errors.As(err, &authErr) {
		return apperrors.ErrNodeForbidden
	}
	var authorizationErr *intkube.ErrAuthorizationFailed
	if errors.As(err, &authorizationErr) {
		return apperrors.ErrNodeForbidden
	}

	if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
		return apperrors.ErrNodeForbidden
	}
	if apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) {
		return apperrors.ErrNodeTimeout
	}

	return err
}
