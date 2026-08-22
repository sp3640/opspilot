package integration

import (
	"net/http"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/models"
)

// TestRegistrationWorkspaceIntegration exercises the Phase 2 onboarding
// contract over HTTP: POST /auth/register optionally names the workspace an
// uninvited registrant creates, that name is ignored when a pending
// invitation determines the organization instead, and GET /users/me reflects
// the resulting organization either way.
func TestRegistrationWorkspaceIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	t.Run("Register with an explicit organization name creates that workspace", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"name":             "Priya Founder",
			"email":            "priya-founder@opspilot.dev",
			"password":         "password123",
			"organizationName": "Priya Industries",
		})
		assertStatus(t, rec, http.StatusCreated)

		token := loginOnly(t, app.router, "priya-founder@opspilot.dev", "password123")
		me := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", token, nil)
		assertStatus(t, me, http.StatusOK)
		data := decodeDataMap(t, me)

		if data["role"] != models.RolePlatformAdmin {
			t.Fatalf("expected new registrant to be Platform Admin, got %v", data["role"])
		}
		if data["organizationName"] != "Priya Industries" {
			t.Fatalf("expected requested organization name to be honored, got %v", data["organizationName"])
		}
	})

	t.Run("Register without an organization name still creates a default workspace", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"name":     "Quentin Founder",
			"email":    "quentin-founder@opspilot.dev",
			"password": "password123",
		})
		assertStatus(t, rec, http.StatusCreated)

		token := loginOnly(t, app.router, "quentin-founder@opspilot.dev", "password123")
		me := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", token, nil)
		assertStatus(t, me, http.StatusOK)
		data := decodeDataMap(t, me)

		if data["role"] != models.RolePlatformAdmin {
			t.Fatalf("expected new registrant to be Platform Admin, got %v", data["role"])
		}
		if data["organizationName"] != "Quentin Founder's Workspace" {
			t.Fatalf("expected default workspace name, got %v", data["organizationName"])
		}
	})

	t.Run("Register with a pending invitation ignores the organization name and joins the invited org", func(t *testing.T) {
		adminToken := registerAndLogin(t, app.router, "Workspace Admin", "workspace-admin@opspilot.dev", "password123")
		admin := mustGetUserByEmail(t, app.userRepo, "workspace-admin@opspilot.dev")
		if admin.OrganizationID == nil {
			t.Fatalf("expected admin organization")
		}

		inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "invited-founder@opspilot.dev",
			"role":  models.RoleDeveloper,
		})
		assertStatus(t, inviteRec, http.StatusCreated)

		registerRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
			"name":             "Invited Founder",
			"email":            "invited-founder@opspilot.dev",
			"password":         "password123",
			"organizationName": "Should Be Ignored Workspace",
		})
		assertStatus(t, registerRec, http.StatusCreated)

		token := loginOnly(t, app.router, "invited-founder@opspilot.dev", "password123")
		me := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", token, nil)
		assertStatus(t, me, http.StatusOK)
		data := decodeDataMap(t, me)

		if data["role"] != models.RoleDeveloper {
			t.Fatalf("expected invited role Developer, got %v", data["role"])
		}
		if data["organizationId"] != admin.OrganizationID.String() {
			t.Fatalf("expected invited registrant to join the inviting organization, got %v", data["organizationId"])
		}

		invited := mustGetUserByEmail(t, app.userRepo, "invited-founder@opspilot.dev")
		if invited.OrganizationID == nil || *invited.OrganizationID != *admin.OrganizationID {
			t.Fatalf("expected invited registrant's organization to match the inviting admin's")
		}
	})
}
