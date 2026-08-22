package logs

import (
	"errors"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
)

// The kubelet reports a previous=true request for a container with no prior
// terminated instance as a plain error string, not a structured Kubernetes
// API status - this can't be exercised through the fake clientset (its
// GetLogs implementation discards whatever a reactor returns), so it is
// covered directly against mapLogError instead.
func TestMapLogError_PreviousInstanceNotFound(t *testing.T) {
	err := errors.New(`previous terminated container "app" in pod "pod-bounds" not found`)

	mapped := mapLogError(err)

	if !errors.Is(mapped, apperrors.ErrLogPreviousNotFound) {
		t.Fatalf("expected ErrLogPreviousNotFound, got %v", mapped)
	}
}

func TestMapLogError_PassesThroughUnrecognizedErrors(t *testing.T) {
	err := errors.New("some other kubelet failure")

	mapped := mapLogError(err)

	if !errors.Is(mapped, err) {
		t.Fatalf("expected the original error to pass through, got %v", mapped)
	}
}
