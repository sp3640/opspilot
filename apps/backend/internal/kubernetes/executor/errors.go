package executor

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var errNamespaceMissing = errors.New("deployment namespace missing")

func mapExecutionError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, errNamespaceMissing) {
		return apperrors.ErrDeploymentNamespaceNotFound
	}

	var invalidConfigErr *intkube.ErrInvalidKubeconfig
	if errors.As(err, &invalidConfigErr) {
		return apperrors.ErrDeploymentInvalidKubeconfig
	}
	if strings.Contains(strings.ToLower(err.Error()), "kubeconfig") {
		return apperrors.ErrDeploymentInvalidKubeconfig
	}

	var connectionErr *intkube.ErrConnectionFailed
	if errors.As(err, &connectionErr) {
		if strings.Contains(strings.ToLower(connectionErr.Error()), "timeout") {
			return apperrors.ErrDeploymentExecutionTimeout
		}
		return apperrors.ErrDeploymentClusterUnreachable
	}

	var authErr *intkube.ErrAuthenticationFailed
	if errors.As(err, &authErr) {
		return apperrors.ErrDeploymentExecutionPermissionDenied
	}

	var authorizationErr *intkube.ErrAuthorizationFailed
	if errors.As(err, &authorizationErr) {
		return apperrors.ErrDeploymentExecutionPermissionDenied
	}

	if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
		return apperrors.ErrDeploymentExecutionPermissionDenied
	}

	if apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) {
		return apperrors.ErrDeploymentExecutionTimeout
	}

	message := strings.ToLower(err.Error())
	if strings.Contains(message, "invalidimagename") || strings.Contains(message, "invalid image") {
		return apperrors.ErrDeploymentImageInvalid
	}

	return fmt.Errorf("%w: %v", apperrors.ErrDeploymentExecutionFailed, err)
}
