package integration

import (
	"net/http"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/models"
)

// TestOrganizationCrossTenantIsolationIntegration reproduces, over HTTP, the
// exact attack the service-level fix in organization_service.go closes: a
// Platform Admin authenticated against their own organization must not be
// able to read, rename, or delete a DIFFERENT organization merely by putting
// its id in the URL path.
func TestOrganizationCrossTenantIsolationIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminAToken := registerAndLogin(t, app.router, "Org A Admin", "org-a-admin@opspilot.dev", "password123")
	adminA := mustGetUserByEmail(t, app.userRepo, "org-a-admin@opspilot.dev")
	if adminA.OrganizationID == nil {
		t.Fatalf("expected org A admin organization")
	}
	orgAID := *adminA.OrganizationID

	registerAndLogin(t, app.router, "Org B Admin", "org-b-admin@opspilot.dev", "password123")
	adminB := mustGetUserByEmail(t, app.userRepo, "org-b-admin@opspilot.dev")

	// registerAndLogin lands non-first, non-invited users in the same
	// default organization as everyone else in this test's DB, so a
	// genuinely separate organization must be created explicitly.
	orgBID := mustCreateTeamTestOrganization(t, app.organizationRepo, "Org B Cross Tenant", "org-b-cross-tenant", adminB.ID)
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, orgBID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign org B admin role: %v", err)
	}
	adminBToken := loginOnly(t, app.router, "org-b-admin@opspilot.dev", "password123")

	t.Run("Org B admin cannot read org A by id", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/organizations/"+orgAID.String(), adminBToken, nil)
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("Org B admin cannot rename org A by id", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/organizations/"+orgAID.String(), adminBToken, map[string]any{
			"name":        "Hijacked Org A",
			"description": "renamed by org B",
		})
		assertStatus(t, rec, http.StatusForbidden)

		// Org A must be unaffected by the rejected attempt.
		unaffected := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/organizations/"+orgAID.String(), adminAToken, nil)
		assertStatus(t, unaffected, http.StatusOK)
		data := decodeDataMap(t, unaffected)
		if data["name"] == "Hijacked Org A" {
			t.Fatalf("expected org A name unaffected by rejected cross-tenant update")
		}
	})

	t.Run("Org B admin cannot delete org A by id", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/organizations/"+orgAID.String(), adminBToken, nil)
		assertStatus(t, rec, http.StatusForbidden)

		// Org A must still exist afterward.
		stillThere := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/organizations/"+orgAID.String(), adminAToken, nil)
		assertStatus(t, stillThere, http.StatusOK)
	})

	t.Run("Each admin can still access their own organization", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/organizations/"+orgAID.String(), adminAToken, nil), http.StatusOK)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/organizations/"+orgBID.String(), adminBToken, nil), http.StatusOK)
	})
}
