package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func decodeMetricsStatus(t *testing.T, rec *httptest.ResponseRecorder) dto.MetricsStatusResponse {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var status dto.MetricsStatusResponse
	if err := json.Unmarshal(env.Data, &status); err != nil {
		t.Fatalf("decode metrics status: %v; data=%s", err, string(env.Data))
	}
	return status
}

// TestClusterMetricsStatusIntegration covers Phase 12's provider-status
// endpoint: it must perform a real (live) reachability check rather than
// just reflecting whatever is in the database, honestly report that no
// application-metrics provider exists, and preserve organization isolation.
func TestClusterMetricsStatusIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminAToken := registerAndLogin(t, app.router, "Metrics Status Admin A", "metrics-status-admin-a@opspilot.dev", "password123")
	adminA := mustGetUserByEmail(t, app.userRepo, "metrics-status-admin-a@opspilot.dev")
	projectA := createProject(t, app.router, adminAToken, "Metrics Status Project A")

	adminBToken := registerAndLogin(t, app.router, "Metrics Status Admin B", "metrics-status-admin-b@opspilot.dev", "password123")
	adminB := mustGetUserByEmail(t, app.userRepo, "metrics-status-admin-b@opspilot.dev")
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, *adminB.OrganizationID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "metrics-status-admin-b@opspilot.dev", "password123")

	server := fakeReachableKubernetesServer(t)
	clusterID := createClusterWithCredential(t, app.router, adminAToken, projectA, "metrics-status-cluster", fakeKubeconfigYAML(server.URL))

	t.Run("reachable cluster with no metrics-server", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/metrics/status", adminAToken, nil)
		assertStatus(t, rec, http.StatusOK)

		status := decodeMetricsStatus(t, rec)

		if !status.KubernetesReachable {
			t.Fatalf("expected the fake reachable server to report kubernetesReachable=true, got %+v", status)
		}
		if status.MetricsServerAvailable {
			t.Fatalf("expected metricsServerAvailable=false against a server with no metrics.k8s.io API, got %+v", status)
		}
		if status.ApplicationMetrics.Available {
			t.Fatalf("expected application metrics to be honestly reported unavailable, got %+v", status.ApplicationMetrics)
		}
		if status.ApplicationMetrics.Reason == "" {
			t.Fatalf("expected a reason to be given for unavailable application metrics")
		}
		if status.LastCollectedAt != nil {
			t.Fatalf("expected no collected metrics yet, got %v", status.LastCollectedAt)
		}
	})

	t.Run("reflects the most recent stored collection", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		_, err := app.metricService.StoreMetrics(t.Context(), projectA, adminA.ID, []dto.CreateMetricRequest{{
			ProjectID:    projectA,
			ClusterID:    clusterID,
			ResourceID:   clusterID,
			ResourceKind: constants.ResourceKindCluster,
			MetricType:   constants.MetricTypeCPU,
			MetricName:   "cluster.cpu.usage.millicores",
			Value:        123,
			Unit:         "m",
			Timestamp:    now,
			Labels:       json.RawMessage(`{"source":"metrics-server"}`),
			Metadata:     json.RawMessage(`{}`),
		}})
		if err != nil {
			t.Fatalf("store metric: %v", err)
		}

		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/metrics/status", adminAToken, nil)
		assertStatus(t, rec, http.StatusOK)

		status := decodeMetricsStatus(t, rec)

		if status.LastCollectedAt == nil || !status.LastCollectedAt.Equal(now) {
			t.Fatalf("expected lastCollectedAt %v, got %v", now, status.LastCollectedAt)
		}
	})

	t.Run("cross-organization access is denied", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String()+"/metrics/status", adminBToken, nil)
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("invalid cluster id", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/not-a-uuid/metrics/status", adminAToken, nil)
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("unknown cluster id", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+uuid.NewString()+"/metrics/status", adminAToken, nil)
		assertStatus(t, rec, http.StatusNotFound)
	})
}
