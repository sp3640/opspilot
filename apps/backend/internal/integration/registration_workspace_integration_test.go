package integration

import (
	"net/http"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/models"
)

// TestRegistrationWorkspaceIntegration exercises the onboarding contract
// over HTTP: POST /auth/register optionally names the workspace the
// registrant creates, and GET /users/me reflects the resulting organization.
// A pending invitation for the registrant's email never changes this -
// registration alone doesn't prove the registrant controls that email
// address, so it must never be sufficient to join someone else's
// organization (see invitation_acceptance_integration_test.go for the
// token-verified accept flow that actually does).
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

	t.Run("Register with a pending invitation still creates its own workspace, not the invited org", func(t *testing.T) {
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
			"organizationName": "Invited Founder's Own Workspace",
		})
		assertStatus(t, registerRec, http.StatusCreated)

		token := loginOnly(t, app.router, "invited-founder@opspilot.dev", "password123")
		me := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", token, nil)
		assertStatus(t, me, http.StatusOK)
		data := decodeDataMap(t, me)

		// A pending invitation must never be enough, by itself, to grant
		// membership in its organization at registration time - only the
		// token-verified /invitations/accept flow may do that.
		if data["role"] != models.RolePlatformAdmin {
			t.Fatalf("expected registrant to be Platform Admin of their own workspace, got %v", data["role"])
		}
		if data["organizationId"] == admin.OrganizationID.String() {
			t.Fatalf("expected registration to create its own workspace, not join the inviting organization via email match alone")
		}
		if data["organizationName"] != "Invited Founder's Own Workspace" {
			t.Fatalf("expected requested organization name to be honored, got %v", data["organizationName"])
		}

		invited := mustGetUserByEmail(t, app.userRepo, "invited-founder@opspilot.dev")
		if invited.OrganizationID == nil || *invited.OrganizationID == *admin.OrganizationID {
			t.Fatalf("expected invited registrant's organization to differ from the inviting admin's")
		}
	})
}
