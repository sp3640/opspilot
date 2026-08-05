package secrets

import (
	"errors"
	"strings"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func mapSecretError(err error) error {
	if err == nil {
		return nil
	}

	var invalidConfigErr *intkube.ErrInvalidKubeconfig
	if errors.As(err, &invalidConfigErr) {
		return apperrors.ErrSecretInvalidKubeconfig
	}
	if strings.Contains(strings.ToLower(err.Error()), "kubeconfig") {
		return apperrors.ErrSecretInvalidKubeconfig
	}

	var connectionErr *intkube.ErrConnectionFailed
	if errors.As(err, &connectionErr) {
		return apperrors.ErrSecretClusterUnavailable
	}

	var authErr *intkube.ErrAuthenticationFailed
	if errors.As(err, &authErr) {
		return apperrors.ErrSecretForbidden
	}
	var authorizationErr *intkube.ErrAuthorizationFailed
	if errors.As(err, &authorizationErr) {
		return apperrors.ErrSecretForbidden
	}

	if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
		return apperrors.ErrSecretForbidden
	}
	if apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) {
		return apperrors.ErrSecretTimeout
	}
	if apierrors.IsNotFound(err) {
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "namespaces") {
			return apperrors.ErrSecretNamespaceNotFound
		}
		return apperrors.ErrSecretNotFound
	}

	return err
}
