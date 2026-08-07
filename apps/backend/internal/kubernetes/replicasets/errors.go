package replicasets

import (
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func mapReplicaSetError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case k8serrors.IsNotFound(err):
		return apperrors.ErrRuntimeDeploymentNotFound

	case k8serrors.IsForbidden(err):
		return apperrors.ErrRuntimeDeploymentForbidden

	case k8serrors.IsUnauthorized(err):
		return apperrors.ErrRuntimeDeploymentForbidden

	default:
		return err
	}
}
