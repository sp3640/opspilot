package executor

import (
	"errors"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestMapExecutionErrorFailurePaths is a direct, isolated regression test
// for the cluster-unreachable/timeout/permission-denied branches of
// mapExecutionError. These branches were previously exercised only
// indirectly, if at all, by the executor integration suite (which tests
// success, missing-namespace, permission-denied via a fake clientset, and
// invalid-kubeconfig, but nothing that reaches the connection-failed or
// timeout classification here). A regression that silently misclassifies a
// genuinely unreachable cluster as a generic execution failure (or vice
// versa) would change what apperror - and therefore what HTTP status and
// message - a caller sees, with no test catching it.
func TestMapExecutionErrorFailurePaths(t *testing.T) {
	gr := schema.GroupResource{Group: "apps", Resource: "deployments"}

	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "nil error passes through as nil",
			err:     nil,
			wantErr: nil,
		},
		{
			name:    "missing namespace",
			err:     errNamespaceMissing,
			wantErr: apperrors.ErrDeploymentNamespaceNotFound,
		},
		{
			name:    "invalid kubeconfig, typed error",
			err:     &intkube.ErrInvalidKubeconfig{Err: errors.New("malformed yaml")},
			wantErr: apperrors.ErrDeploymentInvalidKubeconfig,
		},
		{
			name:    "invalid kubeconfig, message-matched fallback",
			err:     errors.New("failed to parse kubeconfig: unexpected EOF"),
			wantErr: apperrors.ErrDeploymentInvalidKubeconfig,
		},
		{
			name:    "connection failed, unreachable cluster",
			err:     &intkube.ErrConnectionFailed{Err: errors.New("dial tcp 10.0.0.1:6443: connect: no route to host")},
			wantErr: apperrors.ErrDeploymentClusterUnreachable,
		},
		{
			name:    "connection failed, timeout-flavored message maps to timeout instead",
			err:     &intkube.ErrConnectionFailed{Err: errors.New("dial tcp 10.0.0.1:6443: i/o timeout")},
			wantErr: apperrors.ErrDeploymentExecutionTimeout,
		},
		{
			name:    "authentication failed",
			err:     &intkube.ErrAuthenticationFailed{Err: errors.New("invalid bearer token")},
			wantErr: apperrors.ErrDeploymentExecutionPermissionDenied,
		},
		{
			name:    "authorization failed",
			err:     &intkube.ErrAuthorizationFailed{Err: errors.New("user cannot create deployments")},
			wantErr: apperrors.ErrDeploymentExecutionPermissionDenied,
		},
		{
			name:    "apierrors forbidden",
			err:     apierrors.NewForbidden(gr, "my-deployment", errors.New("forbidden")),
			wantErr: apperrors.ErrDeploymentExecutionPermissionDenied,
		},
		{
			name:    "apierrors unauthorized",
			err:     apierrors.NewUnauthorized("unauthorized"),
			wantErr: apperrors.ErrDeploymentExecutionPermissionDenied,
		},
		{
			name:    "apierrors timeout",
			err:     apierrors.NewTimeoutError("timed out", 30),
			wantErr: apperrors.ErrDeploymentExecutionTimeout,
		},
		{
			name:    "apierrors server timeout",
			err:     apierrors.NewServerTimeout(gr, "create", 30),
			wantErr: apperrors.ErrDeploymentExecutionTimeout,
		},
		{
			name:    "invalid image name",
			err:     errors.New("InvalidImageName: could not parse image reference"),
			wantErr: apperrors.ErrDeploymentImageInvalid,
		},
		{
			name:    "unrecognized error falls back to generic execution failure",
			err:     errors.New("something unexpected happened"),
			wantErr: apperrors.ErrDeploymentExecutionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapExecutionError(tt.err)

			if tt.wantErr == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}

			if !errors.Is(got, tt.wantErr) {
				t.Fatalf("expected error wrapping %v, got %v", tt.wantErr, got)
			}
		})
	}
}
