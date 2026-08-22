package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/sp3640/opspilot/backend/internal/models"
)

// TestInvitationAcceptanceIntegration exercises the end-to-end invitation
// acceptance flow over HTTP: validating an invitation, accepting it, and the
// authorization/state rules that must hold around acceptance.
func TestInvitationAcceptanceIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "Accept Admin", "accept-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "accept-admin@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	orgID := *admin.OrganizationID

	t.Run("Valid acceptance updates organization and role", func(t *testing.T) {
		memberToken := registerAndLogin(t, app.router, "Valid Member", "valid-member@opspilot.dev", "password123")

		inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "valid-member@opspilot.dev",
			"role":  models.RolePlatformAdmin,
		})
		assertStatus(t, inviteRec, http.StatusCreated)
		token, _ := decodeDataMap(t, inviteRec)["token"].(string)
		if token == "" {
			t.Fatalf("expected invitation token")
		}

		validateRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/invitations/validate?token="+token, memberToken, nil)
		assertStatus(t, validateRec, http.StatusOK)
		validated := decodeDataMap(t, validateRec)
		if validated["email"] != "valid-member@opspilot.dev" {
			t.Fatalf("expected validated email to match invitation, got %v", validated["email"])
		}
		if validated["role"] != models.RolePlatformAdmin {
			t.Fatalf("expected validated role Platform Admin, got %v", validated["role"])
		}
		if validated["organizationName"] == "" || validated["organizationName"] == nil {
			t.Fatalf("expected validated organization name to be populated")
		}
		if _, hasToken := validated["token"]; hasToken {
			t.Fatalf("expected validate response to omit the raw token")
		}

		acceptRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", memberToken, map[string]any{
			"token": token,
		})
		assertStatus(t, acceptRec, http.StatusOK)
		accepted := decodeDataMap(t, acceptRec)
		if accepted["status"] != models.InvitationStatusAccepted {
			t.Fatalf("expected accepted status, got %v", accepted["status"])
		}

		member := mustGetUserByEmail(t, app.userRepo, "valid-member@opspilot.dev")
		if member.OrganizationID == nil || *member.OrganizationID != orgID {
			t.Fatalf("expected member to join the inviting organization")
		}
		if member.Role != models.RolePlatformAdmin {
			t.Fatalf("expected member role updated to Platform Admin, got %s", member.Role)
		}
	})

	t.Run("Wrong authenticated email cannot validate or accept", func(t *testing.T) {
		registerAndLogin(t, app.router, "Target User", "wrong-email-target@opspilot.dev", "password123")
		attackerToken := registerAndLogin(t, app.router, "Attacker", "wrong-email-attacker@opspilot.dev", "password123")

		inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "wrong-email-target@opspilot.dev",
			"role":  models.RoleViewer,
		})
		assertStatus(t, inviteRec, http.StatusCreated)
		token, _ := decodeDataMap(t, inviteRec)["token"].(string)

		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/invitations/validate?token="+token, attackerToken, nil), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", attackerToken, map[string]any{
			"token": token,
		}), http.StatusForbidden)

		// The attacker legitimately becomes Platform Admin of their OWN new
		// workspace via normal registration — that's expected and unrelated
		// to this attack. What must never happen is the attacker ending up
		// associated with the target's/admin's organization.
		attacker := mustGetUserByEmail(t, app.userRepo, "wrong-email-attacker@opspilot.dev")
		if attacker.OrganizationID != nil && *attacker.OrganizationID == orgID {
			t.Fatalf("attacker must not gain access to another organization's invitation")
		}
	})

	t.Run("Expired invitation cannot be accepted", func(t *testing.T) {
		memberToken := registerAndLogin(t, app.router, "Expired Member", "expired-member@opspilot.dev", "password123")

		expired := &models.Invitation{
			OrganizationID: orgID,
			Email:          "expired-member@opspilot.dev",
			Role:           models.RoleViewer,
			Token:          "expired-accept-token",
			Status:         models.InvitationStatusPending,
			InvitedBy:      admin.ID,
			ExpiresAt:      time.Now().UTC().Add(-1 * time.Hour),
		}
		if err := app.invitationRepo.Create(expired); err != nil {
			t.Fatalf("seed expired invitation: %v", err)
		}

		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", memberToken, map[string]any{
			"token": expired.Token,
		}), http.StatusBadRequest)
	})

	t.Run("Revoked invitation cannot be accepted", func(t *testing.T) {
		memberToken := registerAndLogin(t, app.router, "Revoked Member", "revoked-member@opspilot.dev", "password123")

		inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "revoked-member@opspilot.dev",
			"role":  models.RoleViewer,
		})
		assertStatus(t, inviteRec, http.StatusCreated)
		invited := decodeDataMap(t, inviteRec)
		invitationID, _ := invited["id"].(string)
		token, _ := invited["token"].(string)

		assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/invitations/"+invitationID, adminToken, nil), http.StatusOK)

		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", memberToken, map[string]any{
			"token": token,
		}), http.StatusBadRequest)
	})

	t.Run("Already accepted invitation cannot be accepted twice", func(t *testing.T) {
		memberToken := registerAndLogin(t, app.router, "Twice Member", "twice-member@opspilot.dev", "password123")

		inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "twice-member@opspilot.dev",
			"role":  models.RoleViewer,
		})
		assertStatus(t, inviteRec, http.StatusCreated)
		token, _ := decodeDataMap(t, inviteRec)["token"].(string)

		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", memberToken, map[string]any{
			"token": token,
		}), http.StatusOK)

		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", memberToken, map[string]any{
			"token": token,
		}), http.StatusBadRequest)
	})

	t.Run("Invited registration followed by acceptance stays consistent", func(t *testing.T) {
		inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "invited-then-registered@opspilot.dev",
			"role":  models.RolePlatformAdmin,
		})
		assertStatus(t, inviteRec, http.StatusCreated)
		token, _ := decodeDataMap(t, inviteRec)["token"].(string)

		// Invitation-aware registration already assigns the invited org/role
		// before acceptance ever runs.
		newMemberToken := registerAndLogin(t, app.router, "Invited Registrant", "invited-then-registered@opspilot.dev", "password123")

		newMember := mustGetUserByEmail(t, app.userRepo, "invited-then-registered@opspilot.dev")
		if newMember.OrganizationID == nil || *newMember.OrganizationID != orgID {
			t.Fatalf("expected invitation-aware registration to already assign the invited organization")
		}
		if newMember.Role != models.RolePlatformAdmin {
			t.Fatalf("expected invitation-aware registration to already assign the invited role")
		}

		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", newMemberToken, map[string]any{
			"token": token,
		}), http.StatusOK)

		reloaded := mustGetUserByEmail(t, app.userRepo, "invited-then-registered@opspilot.dev")
		if reloaded.OrganizationID == nil || *reloaded.OrganizationID != orgID {
			t.Fatalf("expected acceptance after invitation-aware registration to remain safe/idempotent")
		}
		if reloaded.Role != models.RolePlatformAdmin {
			t.Fatalf("expected role to remain Platform Admin after acceptance")
		}
	})

	t.Run("Existing user login followed by acceptance moves them into the invited org", func(t *testing.T) {
		existingToken := registerAndLogin(t, app.router, "Existing User", "existing-then-invited@opspilot.dev", "password123")
		existingUser := mustGetUserByEmail(t, app.userRepo, "existing-then-invited@opspilot.dev")

		// Put the existing user in a different organization first, so
		// acceptance is the thing that visibly moves them into orgID.
		otherOrgID := mustCreateTeamTestOrganization(t, app.organizationRepo, "Existing User Org", "existing-user-org-accept", existingUser.ID)
		if err := app.userRepo.AssignOrganizationAndRole(existingUser.ID, otherOrgID, models.RoleViewer); err != nil {
			t.Fatalf("assign existing user to a different organization: %v", err)
		}

		inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "existing-then-invited@opspilot.dev",
			"role":  models.RolePlatformAdmin,
		})
		assertStatus(t, inviteRec, http.StatusCreated)
		token, _ := decodeDataMap(t, inviteRec)["token"].(string)

		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", existingToken, map[string]any{
			"token": token,
		}), http.StatusOK)

		existingAfter := mustGetUserByEmail(t, app.userRepo, "existing-then-invited@opspilot.dev")
		if existingAfter.OrganizationID == nil || *existingAfter.OrganizationID != orgID {
			t.Fatalf("expected existing user to move into the inviting organization after acceptance")
		}
		if existingAfter.Role != models.RolePlatformAdmin {
			t.Fatalf("expected existing user's role updated to Platform Admin after acceptance")
		}
	})
}
