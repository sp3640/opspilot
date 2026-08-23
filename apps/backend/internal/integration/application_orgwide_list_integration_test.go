package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// TestApplicationOrgWideListIntegration exercises GET /applications (Phase
// 22): unscoped it must return every application across every project in
// the caller's own organization (used by the operations dashboard for a
// real org-wide application count), scoped by ?projectId= it must behave
// identically to the existing per-project endpoint, and it must never leak
// another organization's applications.
func TestApplicationOrgWideListIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "OrgWide Admin", "orgwide-admin@opspilot.dev", "password123")
	registerAndLogin(t, app.router, "OrgWide Viewer", "orgwide-viewer@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "orgwide-admin@opspilot.dev")
	viewer := mustGetUserByEmail(t, app.userRepo, "orgwide-viewer@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	organizationA := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(viewer.ID, organizationA, models.RoleViewer); err != nil {
		t.Fatalf("assign viewer to organization: %v", err)
	}
	viewerToken := loginOnly(t, app.router, "orgwide-viewer@opspilot.dev", "password123")

	projectOneID := createProject(t, app.router, adminToken, "OrgWide Project One")
	projectTwoID := createProject(t, app.router, adminToken, "OrgWide Project Two")

	createApp := func(projectID, name string) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID+"/applications", adminToken, map[string]any{
			"name":    name,
			"runtime": constants.ApplicationRuntimeGo,
			"port":    8080,
		})
		assertStatus(t, rec, http.StatusCreated)
	}
	createApp(projectOneID.String(), "Org Wide App One")
	createApp(projectOneID.String(), "Org Wide App Two")
	createApp(projectTwoID.String(), "Org Wide App Three")

	// Unscoped: every application across both projects, readable by a
	// Viewer (visibility is org-membership-based, not permission-gated).
	allRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications?limit=100", viewerToken, nil)
	assertStatus(t, allRec, http.StatusOK)
	allData := decodeDataMap(t, allRec)
	if int(allData["total"].(float64)) != 3 {
		t.Fatalf("expected 3 applications org-wide, got %v", allData["total"])
	}

	// Scoped to one project: identical semantics to the dedicated
	// per-project endpoint.
	scopedRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications?projectId="+projectOneID.String()+"&limit=100", adminToken, nil)
	assertStatus(t, scopedRec, http.StatusOK)
	scopedData := decodeDataMap(t, scopedRec)
	if int(scopedData["total"].(float64)) != 2 {
		t.Fatalf("expected 2 applications scoped to project one, got %v", scopedData["total"])
	}

	// A second organization must never see the first organization's
	// applications via the unscoped endpoint - isolation is automatic
	// (derived from the caller's own JWT organization claim), not an
	// explicit per-request check.
	orgBRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/organizations", adminToken, map[string]any{
		"name":        "OrgWide Org B",
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
	viewerOrgBToken := loginOnly(t, app.router, "orgwide-viewer@opspilot.dev", "password123")

	orgBListRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications?limit=100", viewerOrgBToken, nil)
	assertStatus(t, orgBListRec, http.StatusOK)
	orgBListData := decodeDataMap(t, orgBListRec)
	if int(orgBListData["total"].(float64)) != 0 {
		t.Fatalf("expected 0 applications visible from a different organization, got %v", orgBListData["total"])
	}
}
