package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func TestApplicationManagementIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "App Admin", "app-admin@opspilot.dev", "password123")
	memberToken := registerAndLogin(t, app.router, "App Member", "app-member@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "app-admin@opspilot.dev")
	member := mustGetUserByEmail(t, app.userRepo, "app-member@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	organizationA := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(member.ID, organizationA, models.RoleUser); err != nil {
		t.Fatalf("assign member to organization: %v", err)
	}

	projectID := createProject(t, app.router, adminToken, "Application Project")

	createRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":           "Orders API",
		"description":    "Primary workload",
		"runtime":        constants.ApplicationRuntimeNodeJS,
		"port":           3000,
		"environment":    "production",
		"build_command":  "npm run build",
		"start_command":  "npm start",
		"repository_url": "https://github.com/example/orders-api",
	})
	assertStatus(t, createRec, http.StatusCreated)
	application := decodeDataMap(t, createRec)
	applicationID, err := uuid.Parse(application["id"].(string))
	if err != nil {
		t.Fatalf("parse application id: %v", err)
	}
	generatedSlug := application["slug"].(string)
	if generatedSlug == "" {
		t.Fatalf("expected auto-generated slug")
	}
	if application["defaultBranch"].(string) != constants.ApplicationDefaultBranch {
		t.Fatalf("expected default branch %q, got %v", constants.ApplicationDefaultBranch, application["defaultBranch"])
	}
	if application["status"].(string) != constants.ApplicationStatusDraft {
		t.Fatalf("expected default status %q, got %v", constants.ApplicationStatusDraft, application["status"])
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "Orders API copy",
		"slug":    generatedSlug,
		"runtime": constants.ApplicationRuntimeNodeJS,
		"port":    3001,
	}), http.StatusConflict)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "Broken Runtime",
		"runtime": "Rust",
		"port":    3002,
	}), http.StatusBadRequest)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "Bad Port",
		"runtime": constants.ApplicationRuntimeGo,
		"port":    70000,
	}), http.StatusBadRequest)

	updateRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/applications/"+applicationID.String(), adminToken, map[string]any{
		"name":        "Orders API v2",
		"slug":        "orders-api-v2",
		"description": "Updated workload",
		"runtime":     constants.ApplicationRuntimeGo,
		"port":        8080,
		"status":      constants.ApplicationStatusReady,
	})
	assertStatus(t, updateRec, http.StatusOK)
	updated := decodeDataMap(t, updateRec)
	if updated["name"].(string) != "Orders API v2" {
		t.Fatalf("expected updated name, got %v", updated["name"])
	}
	if updated["slug"].(string) != "orders-api-v2" {
		t.Fatalf("expected updated slug, got %v", updated["slug"])
	}
	if updated["runtime"].(string) != constants.ApplicationRuntimeGo {
		t.Fatalf("expected updated runtime, got %v", updated["runtime"])
	}
	if int(updated["port"].(float64)) != 8080 {
		t.Fatalf("expected updated port, got %v", updated["port"])
	}
	if updated["status"].(string) != constants.ApplicationStatusReady {
		t.Fatalf("expected updated status, got %v", updated["status"])
	}
	if updated["projectId"].(string) != projectID.String() {
		t.Fatalf("expected project id to remain unchanged")
	}
	if updated["organizationId"].(string) != organizationA.String() {
		t.Fatalf("expected organization id to remain unchanged")
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectID.String()+"/applications", memberToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String(), memberToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", memberToken, map[string]any{
		"name":    "Member Forbidden",
		"runtime": constants.ApplicationRuntimePython,
		"port":    3003,
	}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/applications/"+applicationID.String(), memberToken, map[string]any{
		"name":    "Member Forbidden Update",
		"runtime": constants.ApplicationRuntimePython,
		"port":    3004,
	}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/applications/"+applicationID.String(), memberToken, nil), http.StatusForbidden)

	orgBRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/organizations", adminToken, map[string]any{
		"name":        "App Org B",
		"description": "secondary org",
	})
	assertStatus(t, orgBRec, http.StatusCreated)
	orgBData := decodeDataMap(t, orgBRec)
	orgBID, err := uuid.Parse(orgBData["id"].(string))
	if err != nil {
		t.Fatalf("parse org b id: %v", err)
	}
	if err := app.userRepo.AssignOrganizationAndRole(member.ID, orgBID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign member to org b: %v", err)
	}
	memberOrgBToken := loginOnly(t, app.router, "app-member@opspilot.dev", "password123")
	projectBID := createProject(t, app.router, memberOrgBToken, "Org B Project")

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectID.String()+"/applications", memberOrgBToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String(), memberOrgBToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectBID.String()+"/applications", adminToken, map[string]any{
		"name":    "Cross Org Blocked",
		"runtime": constants.ApplicationRuntimeDocker,
		"port":    4000,
	}), http.StatusForbidden)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/applications/"+applicationID.String(), adminToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String(), adminToken, nil), http.StatusNotFound)
}
