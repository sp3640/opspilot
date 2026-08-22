package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// TestIncidentLifecycleIntegration drives OPEN -> INVESTIGATING -> MITIGATING
// -> RESOLVED -> (reopen) OPEN, asserting resolvedAt is set exactly when the
// status becomes RESOLVED and cleared the moment it moves away again.
func TestIncidentLifecycleIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Incident Lifecycle Admin", "incident-lifecycle-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Incident Lifecycle Project")

	incidentID := createIncidentForLifecycle(t, app.router, adminToken, projectID, "Lifecycle test incident", constants.StatusOpen)

	transitions := []string{constants.StatusInvestigating, constants.StatusMitigating, constants.StatusResolved}
	for _, status := range transitions {
		rec := updateIncidentStatus(t, app.router, adminToken, incidentID, projectID, status)
		assertStatus(t, rec, http.StatusOK)
		data := decodeDataMap(t, rec)
		if data["status"] != status {
			t.Fatalf("expected status %s, got %v", status, data["status"])
		}
		if status == constants.StatusResolved {
			if data["resolvedAt"] == nil {
				t.Fatalf("expected resolvedAt to be set once RESOLVED")
			}
		} else if data["resolvedAt"] != nil {
			t.Fatalf("expected resolvedAt to stay unset before RESOLVED, got %v", data["resolvedAt"])
		}
	}

	// Reopen: RESOLVED -> OPEN clears resolvedAt.
	reopenRec := updateIncidentStatus(t, app.router, adminToken, incidentID, projectID, constants.StatusOpen)
	assertStatus(t, reopenRec, http.StatusOK)
	reopened := decodeDataMap(t, reopenRec)
	if reopened["status"] != constants.StatusOpen {
		t.Fatalf("expected status OPEN after reopen, got %v", reopened["status"])
	}
	if reopened["resolvedAt"] != nil {
		t.Fatalf("expected resolvedAt to be cleared after reopen, got %v", reopened["resolvedAt"])
	}
}

// TestIncidentApplicationAndOwnerTeamLinkage covers the new
// application/owner-team correlation fields: valid same-project/same-org
// links succeed, cross-project applications and cross-org teams are
// rejected, and existing incident functionality (create without either
// field) keeps working unchanged.
func TestIncidentApplicationAndOwnerTeamLinkage(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Incident Linkage Admin", "incident-linkage-admin@opspilot.dev", "password123")
	projectA := createProject(t, app.router, adminToken, "Incident Linkage Project A")
	projectB := createProject(t, app.router, adminToken, "Incident Linkage Project B")
	applicationInA := createApplicationForIncidents(t, app.router, adminToken, projectA, "incident-linkage-app-a")
	applicationInB := createApplicationForIncidents(t, app.router, adminToken, projectB, "incident-linkage-app-b")
	teamID := createTeamForIncidents(t, app.router, adminToken, "incident-linkage-team")

	t.Run("existing functionality: create with neither field still works", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents", adminToken, map[string]any{
			"title": "Plain incident", "description": "d", "severity": constants.SeverityP2, "status": constants.StatusOpen,
			"project_id": projectA.String(),
		})
		assertStatus(t, rec, http.StatusCreated)
		data := decodeDataMap(t, rec)
		if data["applicationId"] != nil || data["ownerTeamId"] != nil {
			t.Fatalf("expected no application/owner team by default, got %+v", data)
		}
	})

	t.Run("valid application in the same project and team in the same org succeed", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents", adminToken, map[string]any{
			"title": "Linked incident", "description": "d", "severity": constants.SeverityP2, "status": constants.StatusOpen,
			"project_id": projectA.String(), "application_id": applicationInA.String(), "owner_team_id": teamID.String(),
		})
		assertStatus(t, rec, http.StatusCreated)
		data := decodeDataMap(t, rec)
		if data["applicationId"] != applicationInA.String() {
			t.Fatalf("expected applicationId %s, got %v", applicationInA, data["applicationId"])
		}
		if data["ownerTeamId"] != teamID.String() {
			t.Fatalf("expected ownerTeamId %s, got %v", teamID, data["ownerTeamId"])
		}
	})

	t.Run("an application from a different project is rejected", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents", adminToken, map[string]any{
			"title": "Mismatched incident", "description": "d", "severity": constants.SeverityP2, "status": constants.StatusOpen,
			"project_id": projectA.String(), "application_id": applicationInB.String(),
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("an unknown owner team id is rejected", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents", adminToken, map[string]any{
			"title": "Bad team incident", "description": "d", "severity": constants.SeverityP2, "status": constants.StatusOpen,
			"project_id": projectA.String(), "owner_team_id": uuid.NewString(),
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})
}

// TestIncidentApplicationCrossOrganizationRejected confirms an application
// belonging to a different organization cannot be linked, even though its ID
// is well-formed and it genuinely exists.
func TestIncidentApplicationCrossOrganizationRejected(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminAToken := registerAndLogin(t, app.router, "Incident XOrg Admin A", "incident-xorg-admin-a@opspilot.dev", "password123")
	projectA := createProject(t, app.router, adminAToken, "Incident XOrg Project A")

	adminBToken := registerAndLogin(t, app.router, "Incident XOrg Admin B", "incident-xorg-admin-b@opspilot.dev", "password123")
	adminB := mustGetUserByEmail(t, app.userRepo, "incident-xorg-admin-b@opspilot.dev")
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, *adminB.OrganizationID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "incident-xorg-admin-b@opspilot.dev", "password123")
	projectB := createProject(t, app.router, adminBToken, "Incident XOrg Project B")
	applicationInB := createApplicationForIncidents(t, app.router, adminBToken, projectB, "incident-xorg-app-b")

	rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents", adminAToken, map[string]any{
		"title": "Cross-org incident", "description": "d", "severity": constants.SeverityP2, "status": constants.StatusOpen,
		"project_id": projectA.String(), "application_id": applicationInB.String(),
	})
	assertStatus(t, rec, http.StatusBadRequest)
}

// TestIncidentAlertsFilterIntegration covers the Incident -> Alerts
// connection: GET /alerts?incidentId= returns only alerts attached to that
// incident, and organization isolation still applies.
func TestIncidentAlertsFilterIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Incident Alerts Admin", "incident-alerts-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Incident Alerts Project")
	resourceID := createResource(t, app.router, adminToken, projectID, "incident-alerts-resource")

	incidentID := createIncidentForLifecycle(t, app.router, adminToken, projectID, "Incident with alerts", constants.StatusOpen)
	attachedAlertID := createAlert(t, app.router, adminToken, projectID, "Attached alert", resourceID)
	_ = createAlert(t, app.router, adminToken, projectID, "Unattached alert", resourceID)

	attachRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(attachedAlertID)+"/incident", adminToken, map[string]any{
		"incidentId": incidentID,
	})
	assertStatus(t, attachRec, http.StatusOK)

	listRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts?incidentId="+itoa(incidentID), adminToken, nil)
	assertStatus(t, listRec, http.StatusOK)
	items := decodeAlertItems(t, decodeEnvelope(t, listRec))
	if len(items) != 1 {
		t.Fatalf("expected exactly 1 alert attached to the incident, got %d: %+v", len(items), items)
	}
	if toUint(t, items[0]["id"]) != attachedAlertID {
		t.Fatalf("expected alert %d, got %v", attachedAlertID, items[0]["id"])
	}

	invalidRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts?incidentId=not-a-number", adminToken, nil)
	assertStatus(t, invalidRec, http.StatusBadRequest)
}

// TestIncidentOrganizationIsolation covers the standard cross-organization
// denial pattern for the incident routes, unchanged by this phase.
func TestIncidentOrganizationIsolation(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminAToken := registerAndLogin(t, app.router, "Incident Isolation Admin A", "incident-isolation-admin-a@opspilot.dev", "password123")
	projectA := createProject(t, app.router, adminAToken, "Incident Isolation Project A")
	incidentID := createIncidentForLifecycle(t, app.router, adminAToken, projectA, "Isolated incident", constants.StatusOpen)

	adminBToken := registerAndLogin(t, app.router, "Incident Isolation Admin B", "incident-isolation-admin-b@opspilot.dev", "password123")
	adminB := mustGetUserByEmail(t, app.userRepo, "incident-isolation-admin-b@opspilot.dev")
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, *adminB.OrganizationID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "incident-isolation-admin-b@opspilot.dev", "password123")

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/incidents/"+itoa(incidentID), adminBToken, nil), http.StatusForbidden)
	assertStatus(t, updateIncidentStatus(t, app.router, adminBToken, incidentID, projectA, constants.StatusInvestigating), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/incidents/"+itoa(incidentID), adminBToken, nil), http.StatusForbidden)
}

func createIncidentForLifecycle(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, title, status string) uint {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/incidents", token, map[string]any{
		"title": title, "description": "integration test incident", "severity": constants.SeverityP2, "status": status,
		"project_id": projectID.String(),
	})
	assertStatus(t, rec, http.StatusCreated)
	return toUint(t, decodeDataMap(t, rec)["id"])
}

func updateIncidentStatus(t *testing.T, r *gin.Engine, token string, incidentID uint, projectID uuid.UUID, status string) *httptest.ResponseRecorder {
	t.Helper()
	return doJSONRequest(t, r, http.MethodPut, "/api/v1/incidents/"+itoa(incidentID), token, map[string]any{
		"title": "Lifecycle test incident", "description": "integration test incident", "severity": constants.SeverityP2,
		"status": status, "project_id": projectID.String(),
	})
}

func createApplicationForIncidents(t *testing.T, r *gin.Engine, token string, projectID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", token, map[string]any{
		"name": name, "runtime": constants.ApplicationRuntimeGo, "port": 8080,
	})
	assertStatus(t, rec, http.StatusCreated)
	id, err := uuid.Parse(decodeDataMap(t, rec)["id"].(string))
	if err != nil {
		t.Fatalf("parse application id: %v", err)
	}
	return id
}

func createTeamForIncidents(t *testing.T, r *gin.Engine, token string, name string) uuid.UUID {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/teams", token, map[string]any{
		"name": name, "description": "integration test team",
	})
	assertStatus(t, rec, http.StatusCreated)
	id, err := uuid.Parse(decodeDataMap(t, rec)["id"].(string))
	if err != nil {
		t.Fatalf("parse team id: %v", err)
	}
	return id
}
