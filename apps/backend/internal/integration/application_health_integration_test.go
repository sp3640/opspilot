package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// TestApplicationHealthIntegration drives the full GET /applications/:id/health
// endpoint end-to-end: an application with no data at all, then escalating
// real signals (a successful deployment, an active incident, an active
// alert), asserting the score/state move deterministically and that every
// factor without real backing data reports UNKNOWN rather than a fabricated
// HEALTHY.
func TestApplicationHealthIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "Health Admin", "health-admin@opspilot.dev", "password123")
	registerAndLogin(t, app.router, "Health Viewer", "health-viewer@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "health-admin@opspilot.dev")
	viewer := mustGetUserByEmail(t, app.userRepo, "health-viewer@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	organizationID := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(viewer.ID, organizationID, models.RoleViewer); err != nil {
		t.Fatalf("assign viewer to organization: %v", err)
	}
	viewerToken := loginOnly(t, app.router, "health-viewer@opspilot.dev", "password123")

	projectID := createProject(t, app.router, adminToken, "Health Project")
	clusterID := createCluster(t, app.router, adminToken, projectID, "health-cluster")

	applicationRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "Health App",
		"runtime": constants.ApplicationRuntimeGo,
		"port":    8080,
	})
	assertStatus(t, applicationRec, http.StatusCreated)
	applicationID, err := uuid.Parse(decodeDataMap(t, applicationRec)["id"].(string))
	if err != nil {
		t.Fatalf("parse application id: %v", err)
	}

	// 1. Never deployed, no alerts/incidents: deployment/availability/pod
	// health/error-rate/latency/resource-utilization must all be UNKNOWN.
	// Alerts and incidents are real, precise, zero-cost queries regardless
	// of deployment state, so they legitimately report HEALTHY (0 is a
	// real finding, not "unknown").
	initialRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/health", viewerToken, nil)
	assertStatus(t, initialRec, http.StatusOK)
	initial := decodeDataMap(t, initialRec)
	if initial["state"] != "HEALTHY" {
		t.Fatalf("expected HEALTHY with zero known problems, got %v", initial["state"])
	}
	initialFactors := healthFactorsByKey(t, initialRec)
	for _, key := range []string{"availability", "error_rate", "latency", "pod_health", "deployment", "resource_utilization"} {
		if initialFactors[key]["state"] != "UNKNOWN" {
			t.Fatalf("expected factor %s to be UNKNOWN with no data, got %v", key, initialFactors[key]["state"])
		}
		if initialFactors[key]["score"] != nil {
			t.Fatalf("expected factor %s to have no score while UNKNOWN, got %v", key, initialFactors[key]["score"])
		}
		if initialFactors[key]["reason"] == "" || initialFactors[key]["reason"] == nil {
			t.Fatalf("expected factor %s to explain why it's UNKNOWN", key)
		}
	}
	for _, key := range []string{"alerts", "incidents"} {
		if initialFactors[key]["state"] != "HEALTHY" {
			t.Fatalf("expected factor %s to be HEALTHY with zero real matches, got %v", key, initialFactors[key]["state"])
		}
	}

	// 2. A successful deployment makes the Deployment factor HEALTHY.
	createDeploymentRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
		"applicationId":      applicationID.String(),
		"projectId":          projectID.String(),
		"targetClusterId":    clusterID.String(),
		"image":              "ghcr.io/opspilot/health-app",
		"imageTag":           "v1.0.0",
		"environment":        "production",
		"namespace":          "health-app",
		"replicaCount":       2,
		"deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
	})
	assertStatus(t, createDeploymentRec, http.StatusCreated)
	deploymentID, err := uuid.Parse(decodeDataMap(t, createDeploymentRec)["id"].(string))
	if err != nil {
		t.Fatalf("parse deployment id: %v", err)
	}
	if _, err := app.deploymentService.UpdateDeploymentStatus(context.Background(), deploymentID, organizationID, admin.ID, constants.DeploymentStatusSucceeded); err != nil {
		t.Fatalf("mark deployment succeeded: %v", err)
	}

	afterDeployRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/health", adminToken, nil)
	assertStatus(t, afterDeployRec, http.StatusOK)
	afterDeployFactors := healthFactorsByKey(t, afterDeployRec)
	if afterDeployFactors["deployment"]["state"] != "HEALTHY" {
		t.Fatalf("expected deployment factor HEALTHY after a succeeded deployment, got %v", afterDeployFactors["deployment"]["state"])
	}
	afterDeploy := decodeDataMap(t, afterDeployRec)
	if afterDeploy["state"] != "HEALTHY" {
		t.Fatalf("expected overall HEALTHY, got %v", afterDeploy["state"])
	}

	// 3. An active incident scoped to this application must pull the score
	// down (Incident.ApplicationID is a real column - a precise match).
	incidentRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents", adminToken, map[string]any{
		"title":          "Health incident",
		"description":    "integration",
		"severity":       "P1",
		"status":         "OPEN",
		"project_id":     projectID.String(),
		"application_id": applicationID.String(),
	})
	assertStatus(t, incidentRec, http.StatusCreated)

	afterIncidentRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/health", adminToken, nil)
	assertStatus(t, afterIncidentRec, http.StatusOK)
	afterIncidentFactors := healthFactorsByKey(t, afterIncidentRec)
	if afterIncidentFactors["incidents"]["state"] != "CRITICAL" {
		t.Fatalf("expected incidents factor CRITICAL with an active incident, got %v", afterIncidentFactors["incidents"]["state"])
	}
	afterIncident := decodeDataMap(t, afterIncidentRec)
	afterIncidentScore := int(afterIncident["score"].(float64))
	afterDeployScore := int(afterDeploy["score"].(float64))
	if afterIncidentScore >= afterDeployScore {
		t.Fatalf("expected score to drop after an active incident: before=%d after=%d", afterDeployScore, afterIncidentScore)
	}
	if afterIncident["state"] == "HEALTHY" {
		t.Fatalf("expected overall state to move off HEALTHY once an incident is active, got %v", afterIncident["state"])
	}

	// 4. An active alert attributed to this application via the same
	// metadata.applicationId key the alerting engine itself writes must
	// also pull the score down further.
	now := time.Now().UTC().Format(time.RFC3339)
	alertRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts", adminToken, map[string]any{
		"project_id":       projectID.String(),
		"title":            "Health alert",
		"description":      "integration",
		"severity":         constants.AlertSeverityHigh,
		"status":           constants.AlertStatusOpen,
		"source":           constants.AlertSourceKubernetes,
		"resource_type":    constants.AlertResourceTypeDeployment,
		"resource_id":      deploymentID.String(),
		"fingerprint":      "health-app-fingerprint",
		"occurrence_count": 1,
		"labels":           map[string]any{},
		"metadata":         map[string]any{"applicationId": applicationID.String()},
		"first_seen_at":    now,
		"last_seen_at":     now,
	})
	assertStatus(t, alertRec, http.StatusCreated)

	afterAlertRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/health", adminToken, nil)
	assertStatus(t, afterAlertRec, http.StatusOK)
	afterAlertFactors := healthFactorsByKey(t, afterAlertRec)
	if afterAlertFactors["alerts"]["state"] != "DEGRADED" {
		t.Fatalf("expected alerts factor DEGRADED with one active alert, got %v", afterAlertFactors["alerts"]["state"])
	}
	afterAlert := decodeDataMap(t, afterAlertRec)
	afterAlertScore := int(afterAlert["score"].(float64))
	if afterAlertScore >= afterIncidentScore {
		t.Fatalf("expected score to drop further after an active alert: before=%d after=%d", afterIncidentScore, afterAlertScore)
	}

	// Repeated calls with unchanged data must return the identical score -
	// this is a deterministic computation, not something that drifts.
	repeatRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/health", adminToken, nil)
	assertStatus(t, repeatRec, http.StatusOK)
	repeat := decodeDataMap(t, repeatRec)
	if int(repeat["score"].(float64)) != afterAlertScore {
		t.Fatalf("expected an identical score on repeat evaluation, got %v vs %d", repeat["score"], afterAlertScore)
	}

	// An alert with no applicationId metadata at all (e.g. manually created,
	// not engine-generated) must never be counted - it cannot be honestly
	// attributed to this application.
	unattributedAlertRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts", adminToken, map[string]any{
		"project_id":       projectID.String(),
		"title":            "Unattributed alert",
		"description":      "integration",
		"severity":         constants.AlertSeverityHigh,
		"status":           constants.AlertStatusOpen,
		"source":           constants.AlertSourceManual,
		"resource_type":    constants.AlertResourceTypeDeployment,
		"resource_id":      uuid.NewString(),
		"fingerprint":      "unattributed-fingerprint",
		"occurrence_count": 1,
		"labels":           map[string]any{},
		"metadata":         map[string]any{},
		"first_seen_at":    now,
		"last_seen_at":     now,
	})
	assertStatus(t, unattributedAlertRec, http.StatusCreated)

	afterUnattributedRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/health", adminToken, nil)
	assertStatus(t, afterUnattributedRec, http.StatusOK)
	afterUnattributed := decodeDataMap(t, afterUnattributedRec)
	if int(afterUnattributed["score"].(float64)) != afterAlertScore {
		t.Fatalf("expected an unattributed alert to have no effect on the score: before=%d after=%v", afterAlertScore, afterUnattributed["score"])
	}

	// Cross-org access must never leak health data.
	orgBRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/organizations", adminToken, map[string]any{
		"name":        "Health Org B",
		"description": "secondary org",
	})
	assertStatus(t, orgBRec, http.StatusCreated)
	orgBID, err := uuid.Parse(decodeDataMap(t, orgBRec)["id"].(string))
	if err != nil {
		t.Fatalf("parse org b id: %v", err)
	}
	if err := app.userRepo.AssignOrganizationAndRole(viewer.ID, orgBID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign viewer to org b: %v", err)
	}
	viewerOrgBToken := loginOnly(t, app.router, "health-viewer@opspilot.dev", "password123")
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/health", viewerOrgBToken, nil), http.StatusForbidden)

	// An unknown application id reports not found, not a fabricated score.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+uuid.NewString()+"/health", adminToken, nil), http.StatusNotFound)
}

func healthFactorsByKey(t *testing.T, rec *httptest.ResponseRecorder) map[string]map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Factors []map[string]any `json:"factors"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode health factors: %v; data=%s", err, string(env.Data))
	}

	byKey := make(map[string]map[string]any, len(payload.Factors))
	for _, factor := range payload.Factors {
		key, ok := factor["key"].(string)
		if !ok {
			t.Fatalf("factor missing string key: %+v", factor)
		}
		byKey[key] = factor
	}

	return byKey
}
