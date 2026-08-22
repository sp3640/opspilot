package logs

import (
	"errors"
	"strings"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func mapLogError(err error) error {
	if err == nil {
		return nil
	}

	var invalidConfigErr *intkube.ErrInvalidKubeconfig
	if errors.As(err, &invalidConfigErr) {
		return apperrors.ErrLogInvalidKubeconfig
	}
	if strings.Contains(strings.ToLower(err.Error()), "kubeconfig") {
		return apperrors.ErrLogInvalidKubeconfig
	}

	var connectionErr *intkube.ErrConnectionFailed
	if errors.As(err, &connectionErr) {
		return apperrors.ErrLogClusterUnavailable
	}

	var authErr *intkube.ErrAuthenticationFailed
	if errors.As(err, &authErr) {
		return apperrors.ErrLogForbidden
	}
	var authorizationErr *intkube.ErrAuthorizationFailed
	if errors.As(err, &authorizationErr) {
		return apperrors.ErrLogForbidden
	}

	if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
		return apperrors.ErrLogForbidden
	}
	if apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) {
		return apperrors.ErrLogTimeout
	}
	if apierrors.IsNotFound(err) {
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "namespaces") {
			return apperrors.ErrLogNamespaceNotFound
		}
		if strings.Contains(message, "pods") {
			return apperrors.ErrLogPodNotFound
		}
	}

	// The kubelet reports a `previous=true` request for a container with no
	// prior terminated instance as a plain (non-structured) error rather than
	// a Kubernetes API NotFound status, so it must be matched on message text.
	if strings.Contains(strings.ToLower(err.Error()), "previous terminated container") {
		return apperrors.ErrLogPreviousNotFound
	}

	return err
}
