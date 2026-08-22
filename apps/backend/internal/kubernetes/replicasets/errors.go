package replicasets

import (
	"errors"
	"strings"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func mapReplicaSetError(err error) error {
	if err == nil {
		return nil
	}

	var invalidConfigErr *intkube.ErrInvalidKubeconfig
	if errors.As(err, &invalidConfigErr) {
		return apperrors.ErrRuntimeDeploymentInvalidKubeconfig
	}
	if strings.Contains(strings.ToLower(err.Error()), "kubeconfig") {
		return apperrors.ErrRuntimeDeploymentInvalidKubeconfig
	}

	var connectionErr *intkube.ErrConnectionFailed
	if errors.As(err, &connectionErr) {
		return apperrors.ErrRuntimeDeploymentClusterUnavailable
	}

	var authErr *intkube.ErrAuthenticationFailed
	if errors.As(err, &authErr) {
		return apperrors.ErrRuntimeDeploymentForbidden
	}
	var authorizationErr *intkube.ErrAuthorizationFailed
	if errors.As(err, &authorizationErr) {
		return apperrors.ErrRuntimeDeploymentForbidden
	}

	switch {
	case k8serrors.IsNotFound(err):
		return apperrors.ErrRuntimeDeploymentNotFound
	case k8serrors.IsForbidden(err):
		return apperrors.ErrRuntimeDeploymentForbidden
	case k8serrors.IsUnauthorized(err):
		return apperrors.ErrRuntimeDeploymentForbidden
	case k8serrors.IsTimeout(err), k8serrors.IsServerTimeout(err):
		return apperrors.ErrRuntimeDeploymentTimeout
	default:
		return err
	}
}
