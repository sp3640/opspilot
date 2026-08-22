package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// fakeKubernetesAPIServer stands in for a real Kubernetes API server, serving
// just enough of the wire protocol (readyz, SelfSubjectAccessReview, version)
// for Validator.ValidateConnection to exercise the real client-go request
// path end to end over loopback HTTP, rather than mocking the validator away.
func fakeKubernetesAPIServer(t *testing.T, allowed bool) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/apis/authorization.k8s.io/v1/selfsubjectaccessreviews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"apiVersion": "authorization.k8s.io/v1",
			"kind":       "SelfSubjectAccessReview",
			"status": map[string]any{
				"allowed": allowed,
				"reason":  "fake server decision",
			},
		})
	})
	mux.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"major":      "1",
			"minor":      "30",
			"gitVersion": "v1.30.4",
			"platform":   "linux/amd64",
		})
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func fakeKubeconfig(serverURL string) []byte {
	return []byte(fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: fake
  cluster:
    server: %s
    insecure-skip-tls-verify: true
contexts:
- name: fake
  context:
    cluster: fake
    user: fake
current-context: fake
users:
- name: fake
  user:
    token: fake-token
`, serverURL))
}

func TestValidateConnectionSucceedsAgainstReachableCluster(t *testing.T) {
	server := fakeKubernetesAPIServer(t, true)

	client := NewClient(fakeKubeconfig(server.URL))
	validator := NewValidator(client)

	result, err := validator.ValidateConnection(context.Background())
	if err != nil {
		t.Fatalf("expected validation to succeed, got error: %v", err)
	}
	if !result.Connected {
		t.Fatalf("expected Connected to be true")
	}
	if result.ClusterVersion == nil || result.ClusterVersion.GitVersion != "v1.30.4" {
		t.Fatalf("expected cluster version v1.30.4, got %+v", result.ClusterVersion)
	}
	if result.APIServerURL != server.URL {
		t.Fatalf("expected api server url %s, got %s", server.URL, result.APIServerURL)
	}
}

func TestValidateConnectionFailsWhenPermissionDenied(t *testing.T) {
	server := fakeKubernetesAPIServer(t, false)

	client := NewClient(fakeKubeconfig(server.URL))
	validator := NewValidator(client)

	_, err := validator.ValidateConnection(context.Background())
	if err == nil {
		t.Fatalf("expected validation to fail when permission is denied")
	}
	var authzErr *ErrAuthorizationFailed
	if !errors.As(err, &authzErr) {
		t.Fatalf("expected ErrAuthorizationFailed, got %T: %v", err, err)
	}
}

func TestValidateConnectionFailsWithInvalidKubeconfig(t *testing.T) {
	client := NewClient([]byte("this is not a kubeconfig"))
	validator := NewValidator(client)

	_, err := validator.ValidateConnection(context.Background())
	if err == nil {
		t.Fatalf("expected validation to fail with invalid kubeconfig")
	}
	var invalidErr *ErrInvalidKubeconfig
	if !errors.As(err, &invalidErr) {
		t.Fatalf("expected ErrInvalidKubeconfig, got %T: %v", err, err)
	}
}

func TestValidateConnectionFailsWhenUnreachable(t *testing.T) {
	// Port 0's ephemeral allocation is closed immediately after Close(), so
	// this address is guaranteed to refuse connections without needing a
	// real unreachable host (keeps the test hermetic and fast).
	server := httptest.NewServer(http.NewServeMux())
	unreachableURL := server.URL
	server.Close()

	client := NewClient(fakeKubeconfig(unreachableURL))
	validator := NewValidator(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := validator.ValidateConnection(ctx)
	if err == nil {
		t.Fatalf("expected validation to fail against an unreachable server")
	}
	var connErr *ErrConnectionFailed
	if !errors.As(err, &connErr) {
		t.Fatalf("expected ErrConnectionFailed, got %T: %v", err, err)
	}
}
