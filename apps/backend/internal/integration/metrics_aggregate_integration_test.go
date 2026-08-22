package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestMetricsAggregateScoping covers Phase 13's addition of optional
// clusterId/resourceId scoping to GET /metrics/aggregate.
//
// The happy path (a request that actually reaches the repository's bucketed
// query) cannot be exercised here: MetricRepository.Aggregate buckets with
// Postgres' date_trunc(), which the SQLite test database used by this
// integration harness does not implement ("no such function: date_trunc").
// Production always runs against Postgres (internal/database/database.go
// only ever opens postgres.Open), so this is a pre-existing test-harness
// gap, not a production correctness issue - the same class of limitation
// already noted for the fake Kubernetes clientset in earlier phases. What
// IS fully covered here is everything that must be validated before the SQL
// executes: parameter parsing, project ownership, and time-range validation.
func TestMetricsAggregateScoping(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminAToken := registerAndLogin(t, app.router, "Aggregate Admin A", "aggregate-admin-a@opspilot.dev", "password123")
	projectA := createProject(t, app.router, adminAToken, "Aggregate Project A")
	clusterA := createCluster(t, app.router, adminAToken, projectA, "aggregate-cluster-a")

	adminBToken := registerAndLogin(t, app.router, "Aggregate Admin B", "aggregate-admin-b@opspilot.dev", "password123")

	now := time.Now().UTC()
	start := now.Add(-1 * time.Hour).Format(time.RFC3339)
	end := now.Format(time.RFC3339)

	basePath := "/api/v1/metrics/aggregate?projectId=" + projectA.String() + "&metricType=CPU&interval=hour&start=" + start + "&end=" + end

	t.Run("invalid clusterId is rejected before touching the repository", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, basePath+"&clusterId=not-a-uuid", adminAToken, nil)
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("invalid resourceId is rejected before touching the repository", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, basePath+"&resourceId=not-a-uuid", adminAToken, nil)
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("a well-formed but unowned clusterId is still accepted at the parsing layer", func(t *testing.T) {
		// Ownership of the cluster/resource ID itself isn't separately
		// checked here (only project ownership is) - an unrelated cluster's
		// aggregate would just and up empty. Confirm this at least parses
		// and reaches project-ownership validation rather than 400ing.
		rec := doJSONRequest(t, app.router, http.MethodGet, basePath+"&clusterId="+uuid.NewString(), adminAToken, nil)
		if rec.Code == http.StatusBadRequest {
			t.Fatalf("expected parsing to succeed for a well-formed UUID, got 400: %s", rec.Body.String())
		}
	})

	t.Run("missing metricType is rejected", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/metrics/aggregate?projectId="+projectA.String()+"&clusterId="+clusterA.String()+"&interval=hour&start="+start+"&end="+end, adminAToken, nil)
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("end before start is rejected", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/metrics/aggregate?projectId="+projectA.String()+"&clusterId="+clusterA.String()+"&metricType=CPU&interval=hour&start="+end+"&end="+start, adminAToken, nil)
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("cross-organization project access is denied", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, basePath+"&clusterId="+clusterA.String(), adminBToken, nil)
		assertStatus(t, rec, http.StatusForbidden)
	})
}
