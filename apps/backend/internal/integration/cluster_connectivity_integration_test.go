package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
)

// fakeReachableKubernetesServer stands in for a real Kubernetes API server:
// enough of the wire protocol (readyz, SelfSubjectAccessReview, version) for
// the real connectivity validator to succeed against it over loopback HTTP.
func fakeReachableKubernetesServer(t *testing.T) *httptest.Server {
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
			"status":     map[string]any{"allowed": true},
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

func fakeKubeconfigYAML(serverURL string) string {
	return fmt.Sprintf(`apiVersion: v1
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
`, serverURL)
}

// createClusterWithCredential mirrors createCluster but lets the caller
// supply the exact kubeconfig plaintext, needed to point at a fake server or
// submit deliberately invalid content.
func createClusterWithCredential(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name, kubeconfig string) uuid.UUID {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/clusters", token, map[string]any{
		"project_id":           projectID.String(),
		"name":                 name,
		"provider":             constants.ClusterProviderKubernetes,
		"connection_type":      constants.ClusterConnectionTypeKubeconfig,
		"kubeconfig_encrypted": kubeconfig,
		"api_endpoint":         "https://api.example",
		"region":               "us-east-1",
		"metadata":             map[string]any{"env": "test"},
	})
	assertStatus(t, rec, http.StatusCreated)
	data := decodeDataMap(t, rec)
	id, _ := data["id"].(string)
	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("parse cluster id: %v", err)
	}
	return parsed
}

func TestClusterConnectivityIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "Cluster Conn Admin", "cluster-conn-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "cluster-conn-admin@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	orgID := *admin.OrganizationID
	projectID := createProject(t, app.router, adminToken, "Cluster Connectivity Project")

	t.Run("valid cluster reports connected with version", func(t *testing.T) {
		server := fakeReachableKubernetesServer(t)
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "valid-cluster", fakeKubeconfigYAML(server.URL))

		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters/"+clusterID.String()+"/validate", adminToken, nil)
		assertStatus(t, rec, http.StatusOK)
		data := decodeDataMap(t, rec)

		if connected, _ := data["connected"].(bool); !connected {
			t.Fatalf("expected connected=true, got %+v", data)
		}
		if status, _ := data["status"].(string); status != constants.ClusterStatusHealthy {
			t.Fatalf("expected status HEALTHY, got %v", data["status"])
		}
		if version, _ := data["kubernetesVersion"].(string); version != "v1.30.4" {
			t.Fatalf("expected kubernetesVersion v1.30.4, got %v", data["kubernetesVersion"])
		}

		getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String(), adminToken, nil)
		assertStatus(t, getRec, http.StatusOK)
		getData := decodeDataMap(t, getRec)
		if status, _ := getData["status"].(string); status != constants.ClusterStatusHealthy {
			t.Fatalf("expected persisted status HEALTHY, got %v", getData["status"])
		}
		if _, hasCredential := getData["kubeconfigEncrypted"]; hasCredential {
			t.Fatalf("cluster response must never include the credential field")
		}
		if _, hasCredential := getData["encryptedCredential"]; hasCredential {
			t.Fatalf("cluster response must never include the credential field")
		}
	})

	t.Run("invalid credentials report disconnected with a clear error", func(t *testing.T) {
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "invalid-cred-cluster", "this is not valid kubeconfig content")

		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters/"+clusterID.String()+"/validate", adminToken, nil)
		assertStatus(t, rec, http.StatusOK)
		data := decodeDataMap(t, rec)

		if connected, _ := data["connected"].(bool); connected {
			t.Fatalf("expected connected=false for invalid kubeconfig")
		}
		if status, _ := data["status"].(string); status != constants.ClusterStatusInvalid {
			t.Fatalf("expected status INVALID, got %v", data["status"])
		}
		errMsg, _ := data["error"].(string)
		if errMsg == "" {
			t.Fatalf("expected a non-empty validation error message")
		}
	})

	t.Run("disconnected cluster reports unreachable", func(t *testing.T) {
		unreachable := httptest.NewServer(http.NewServeMux())
		unreachableURL := unreachable.URL
		unreachable.Close()

		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "disconnected-cluster", fakeKubeconfigYAML(unreachableURL))

		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters/"+clusterID.String()+"/validate", adminToken, nil)
		assertStatus(t, rec, http.StatusOK)
		data := decodeDataMap(t, rec)

		if connected, _ := data["connected"].(bool); connected {
			t.Fatalf("expected connected=false for an unreachable cluster")
		}
		if status, _ := data["status"].(string); status != constants.ClusterStatusInvalid {
			t.Fatalf("expected status INVALID, got %v", data["status"])
		}
	})

	t.Run("cluster creation requires a non-empty credential", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters", adminToken, map[string]any{
			"project_id":           projectID.String(),
			"name":                 "no-credential-cluster",
			"provider":             constants.ClusterProviderKubernetes,
			"connection_type":      constants.ClusterConnectionTypeKubeconfig,
			"kubeconfig_encrypted": "",
			"api_endpoint":         "https://api.example",
			"region":               "us-east-1",
			"metadata":             map[string]any{},
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("cross-organization access is forbidden", func(t *testing.T) {
		server := fakeReachableKubernetesServer(t)
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "cross-org-cluster", fakeKubeconfigYAML(server.URL))

		otherToken := registerAndLogin(t, app.router, "Cluster Conn Other Admin", "cluster-conn-other-admin@opspilot.dev", "password123")

		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String(), otherToken, nil), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters/"+clusterID.String()+"/validate", otherToken, nil), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/clusters/"+clusterID.String(), otherToken, map[string]any{
			"project_id":      projectID.String(),
			"name":            "hijacked",
			"provider":        constants.ClusterProviderKubernetes,
			"connection_type": constants.ClusterConnectionTypeKubeconfig,
			"api_endpoint":    "https://api.example",
			"region":          "us-east-1",
			"metadata":        map[string]any{},
		}), http.StatusForbidden)
		assertListExcludesID(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters", otherToken, nil), clusterID.String())
	})

	t.Run("credential is preserved when editing unrelated fields", func(t *testing.T) {
		server := fakeReachableKubernetesServer(t)
		// CreateCluster trims surrounding whitespace before encrypting, so
		// the round-tripped value is compared against the trimmed original.
		originalKubeconfig := strings.TrimSpace(fakeKubeconfigYAML(server.URL))
		clusterID := createClusterWithCredential(t, app.router, adminToken, projectID, "edit-preserve-cluster", originalKubeconfig)

		cipher := testClusterCredentialCipher(t)
		before, err := app.clusterRepo.FindByID(clusterID, orgID)
		if err != nil {
			t.Fatalf("load cluster before edit: %v", err)
		}
		beforePlaintext, err := cipher.Decrypt(before.KubeconfigEncrypted)
		if err != nil {
			t.Fatalf("decrypt credential before edit: %v", err)
		}
		if beforePlaintext != originalKubeconfig {
			t.Fatalf("stored credential does not match what was submitted at creation")
		}

		// Rename-only edit: the JSON body intentionally omits
		// kubeconfig_encrypted entirely, exactly like a frontend that never
		// has the plaintext/ciphertext to round-trip.
		renameRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/clusters/"+clusterID.String(), adminToken, map[string]any{
			"project_id":      projectID.String(),
			"name":            "edit-preserve-cluster-renamed",
			"provider":        constants.ClusterProviderKubernetes,
			"connection_type": constants.ClusterConnectionTypeKubeconfig,
			"api_endpoint":    "https://api.example",
			"region":          "us-east-1",
			"metadata":        map[string]any{},
		})
		assertStatus(t, renameRec, http.StatusOK)

		after, err := app.clusterRepo.FindByID(clusterID, orgID)
		if err != nil {
			t.Fatalf("load cluster after edit: %v", err)
		}
		afterPlaintext, err := cipher.Decrypt(after.KubeconfigEncrypted)
		if err != nil {
			t.Fatalf("decrypt credential after edit: %v", err)
		}
		if afterPlaintext != originalKubeconfig {
			t.Fatalf("credential was altered by an edit that never sent kubeconfig_encrypted")
		}
		if after.Name != "edit-preserve-cluster-renamed" {
			t.Fatalf("expected name to be updated, got %s", after.Name)
		}

		// The preserved credential must still be usable: validating again
		// against the same still-reachable fake server must still succeed.
		validateRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters/"+clusterID.String()+"/validate", adminToken, nil)
		assertStatus(t, validateRec, http.StatusOK)
		validateData := decodeDataMap(t, validateRec)
		if connected, _ := validateData["connected"].(bool); !connected {
			t.Fatalf("expected preserved credential to still validate successfully")
		}

		// Explicitly sending an empty credential is rejected rather than
		// silently wiping or silently ignoring it.
		blankRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/clusters/"+clusterID.String(), adminToken, map[string]any{
			"project_id":           projectID.String(),
			"name":                 "edit-preserve-cluster-renamed",
			"provider":             constants.ClusterProviderKubernetes,
			"connection_type":      constants.ClusterConnectionTypeKubeconfig,
			"kubeconfig_encrypted": "",
			"api_endpoint":         "https://api.example",
			"region":               "us-east-1",
			"metadata":             map[string]any{},
		})
		assertStatus(t, blankRec, http.StatusBadRequest)

		stillIntact, err := app.clusterRepo.FindByID(clusterID, orgID)
		if err != nil {
			t.Fatalf("load cluster after rejected blank edit: %v", err)
		}
		stillIntactPlaintext, err := cipher.Decrypt(stillIntact.KubeconfigEncrypted)
		if err != nil {
			t.Fatalf("decrypt credential after rejected blank edit: %v", err)
		}
		if stillIntactPlaintext != originalKubeconfig {
			t.Fatalf("credential was altered by a rejected blank-credential edit")
		}

		// Replacing the credential with a new, different one is honored, and
		// resets validation state to PENDING_VALIDATION.
		replacementKubeconfig := "still not a valid kubeconfig"
		replaceRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/clusters/"+clusterID.String(), adminToken, map[string]any{
			"project_id":           projectID.String(),
			"name":                 "edit-preserve-cluster-renamed",
			"provider":             constants.ClusterProviderKubernetes,
			"connection_type":      constants.ClusterConnectionTypeKubeconfig,
			"kubeconfig_encrypted": replacementKubeconfig,
			"api_endpoint":         "https://api.example",
			"region":               "us-east-1",
			"metadata":             map[string]any{},
		})
		assertStatus(t, replaceRec, http.StatusOK)
		replaceData := decodeDataMap(t, replaceRec)
		if status, _ := replaceData["status"].(string); status != constants.ClusterStatusPendingValidation {
			t.Fatalf("expected status PENDING_VALIDATION after credential replacement, got %v", replaceData["status"])
		}

		replaced, err := app.clusterRepo.FindByID(clusterID, orgID)
		if err != nil {
			t.Fatalf("load cluster after replacement: %v", err)
		}
		replacedPlaintext, err := cipher.Decrypt(replaced.KubeconfigEncrypted)
		if err != nil {
			t.Fatalf("decrypt credential after replacement: %v", err)
		}
		if replacedPlaintext != replacementKubeconfig {
			t.Fatalf("expected credential to be replaced with the new value")
		}
	})
}
