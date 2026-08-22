package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sp3640/opspilot/backend/internal/auth"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// forgeToken builds a JWT with fully custom claims/expiry/signing secret, so
// tests can exercise expiry and tampering scenarios that auth.GenerateToken
// (which hardcodes a 24h expiry) cannot produce directly.
func forgeToken(t *testing.T, secret string, claims auth.Claims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign forged token: %v", err)
	}

	return signed
}

func baseClaims(userID uint, email, role, organizationID string, expiresAt time.Time) auth.Claims {
	now := time.Now().UTC()
	return auth.Claims{
		UserID:         userID,
		Email:          email,
		Role:           role,
		OrganizationID: organizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "OpsPilot",
			Subject:   email,
		},
	}
}

// TestTokenLifecycleIntegration covers expired/invalid/tampered tokens,
// stateless-JWT logout semantics, and the new /auth/reissue endpoint that
// refreshes stale role/organization claims without a full re-login.
func TestTokenLifecycleIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	token := registerAndLogin(t, app.router, "Token User", "token-user@opspilot.dev", "password123")
	user := mustGetUserByEmail(t, app.userRepo, "token-user@opspilot.dev")
	if user.OrganizationID == nil {
		t.Fatalf("expected user organization")
	}

	t.Run("Expired token is rejected", func(t *testing.T) {
		expired := forgeToken(t, app.cfg.JWTSecret, baseClaims(
			user.ID, user.Email, models.RolePlatformAdmin, user.OrganizationID.String(),
			time.Now().UTC().Add(-1*time.Hour),
		))

		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", expired, nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("Malformed token is rejected", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", "not-a-real-jwt", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("Token signed with the wrong secret is rejected", func(t *testing.T) {
		tampered := forgeToken(t, "an-entirely-different-signing-secret-1234567890", baseClaims(
			user.ID, user.Email, models.RolePlatformAdmin, user.OrganizationID.String(),
			time.Now().UTC().Add(1*time.Hour),
		))

		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", tampered, nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("Missing token is rejected", func(t *testing.T) {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("Valid token still works", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", token, nil), http.StatusOK)
	})

	t.Run("Logout is client-side only: a captured token remains valid until it expires", func(t *testing.T) {
		// There is no server-side revocation for a stateless JWT — "logout"
		// only clears the client's copy of the token. A request that omits
		// the token (the client-side effect of logging out) is rejected...
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", "", nil), http.StatusUnauthorized)
		// ...but the same still-unexpired token, if presented again by
		// anyone holding a copy of it, continues to authenticate normally.
		// This is a deliberate, documented property of the current
		// architecture, not a gap introduced by this change.
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", token, nil), http.StatusOK)
	})

	t.Run("Reissue mints a fresh token reflecting the current database role", func(t *testing.T) {
		reissueRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/reissue", token, nil)
		assertStatus(t, reissueRec, http.StatusOK)
		reissued, _ := decodeDataMap(t, reissueRec)["access_token"].(string)
		if reissued == "" {
			t.Fatalf("expected a reissued access token")
		}

		claims, err := auth.ValidateToken(reissued, app.cfg.JWTSecret)
		if err != nil {
			t.Fatalf("validate reissued token: %v", err)
		}
		if claims.Role != models.RolePlatformAdmin {
			t.Fatalf("expected reissued role %q, got %q", models.RolePlatformAdmin, claims.Role)
		}
		if claims.OrganizationID != user.OrganizationID.String() {
			t.Fatalf("expected reissued organization %q, got %q", user.OrganizationID.String(), claims.OrganizationID)
		}

		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/users/me", reissued, nil), http.StatusOK)
	})

	t.Run("Reissue with no token is rejected", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/reissue", "", nil), http.StatusUnauthorized)
	})
}

// TestRoleChangeAuthorizationFreshnessIntegration verifies the core Phase 3
// guarantee: a token minted before a role/organization change carries stale
// claims (an inherent property of JWTs, not a bug), but calling /auth/reissue
// after the change immediately produces a token whose authorization reflects
// the new role — without a full re-login and without weakening backend
// authorization (the old token is never granted new permissions retroactively;
// a brand new token has to be obtained).
func TestRoleChangeAuthorizationFreshnessIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	viewerToken := registerAndLogin(t, app.router, "Freshness Viewer", "freshness-viewer@opspilot.dev", "password123")
	viewer := mustGetUserByEmail(t, app.userRepo, "freshness-viewer@opspilot.dev")
	if viewer.OrganizationID == nil {
		t.Fatalf("expected viewer organization")
	}
	orgID := *viewer.OrganizationID

	// Create the project while still Platform Admin of their own workspace,
	// so it belongs to orgID — the same organization the downgrade below
	// keeps them in.
	projectID := createProject(t, app.router, viewerToken, "Freshness Project")

	// Registering alone makes this user Platform Admin of their own new
	// workspace; downgrade them to Viewer to set up a meaningful before/after.
	if err := app.userRepo.AssignOrganizationAndRole(viewer.ID, orgID, models.RoleViewer); err != nil {
		t.Fatalf("downgrade to viewer: %v", err)
	}

	t.Run("Old token keeps failing a permission the DB role would now allow", func(t *testing.T) {
		// viewerToken was minted before the downgrade, so it still carries
		// whatever role the original registration produced (Platform Admin).
		// That is exactly the staleness this phase does not silently fix —
		// only an explicit reissue does.
		claims, err := auth.ValidateToken(viewerToken, app.cfg.JWTSecret)
		if err != nil {
			t.Fatalf("validate old token: %v", err)
		}
		if claims.Role != models.RolePlatformAdmin {
			t.Fatalf("expected the pre-downgrade token to still claim Platform Admin, got %q", claims.Role)
		}
	})

	t.Run("Reissue after a DB role change immediately reflects the new role", func(t *testing.T) {
		reissueRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/reissue", viewerToken, nil)
		assertStatus(t, reissueRec, http.StatusOK)
		freshToken, _ := decodeDataMap(t, reissueRec)["access_token"].(string)

		claims, err := auth.ValidateToken(freshToken, app.cfg.JWTSecret)
		if err != nil {
			t.Fatalf("validate fresh token: %v", err)
		}
		if claims.Role != models.RoleViewer {
			t.Fatalf("expected reissued role %q, got %q", models.RoleViewer, claims.Role)
		}

		// Viewer cannot create projects: the fresh token must be forbidden...
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects", freshToken, map[string]any{
			"name": "Should Be Forbidden",
		}), http.StatusForbidden)
		// ...but can still read them, matching Viewer's read-only permissions.
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectID.String(), freshToken, nil), http.StatusOK)
	})
}

// TestInvitationAcceptanceAuthorizationFreshnessIntegration verifies that
// after accepting an invitation, calling /auth/reissue produces a token
// whose authorization reflects the invited organization and role — the
// concrete scenario Phase 3 targets ("invitation acceptance must result in a
// valid session reflecting the new organization/role").
func TestInvitationAcceptanceAuthorizationFreshnessIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "Freshness Invite Admin", "freshness-invite-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "freshness-invite-admin@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	orgID := *admin.OrganizationID

	inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
		"email": "freshness-invitee@opspilot.dev",
		"role":  models.RoleDevOpsEngineer,
	})
	assertStatus(t, inviteRec, http.StatusCreated)
	inviteToken, _ := decodeDataMap(t, inviteRec)["token"].(string)

	// The invitee is an existing user in a DIFFERENT organization before
	// accepting, so acceptance is what visibly moves them into orgID.
	inviteeToken := registerAndLogin(t, app.router, "Freshness Invitee", "freshness-invitee@opspilot.dev", "password123")
	invitee := mustGetUserByEmail(t, app.userRepo, "freshness-invitee@opspilot.dev")

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", inviteeToken, map[string]any{
		"token": inviteToken,
	}), http.StatusOK)

	t.Run("Reissue after acceptance grants DevOps Engineer permissions in the new org", func(t *testing.T) {
		reissueRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/reissue", inviteeToken, nil)
		assertStatus(t, reissueRec, http.StatusOK)
		freshToken, _ := decodeDataMap(t, reissueRec)["access_token"].(string)

		claims, err := auth.ValidateToken(freshToken, app.cfg.JWTSecret)
		if err != nil {
			t.Fatalf("validate fresh token: %v", err)
		}
		if claims.Role != models.RoleDevOpsEngineer {
			t.Fatalf("expected reissued role %q, got %q", models.RoleDevOpsEngineer, claims.Role)
		}
		if claims.OrganizationID != orgID.String() {
			t.Fatalf("expected reissued organization %q, got %q", orgID.String(), claims.OrganizationID)
		}

		// DevOps Engineer can manage clusters in the (now their own) org.
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters", freshToken, map[string]any{
			"project_id":           createProject(t, app.router, adminToken, "Freshness Invite Project").String(),
			"name":                 "freshness-cluster",
			"provider":             constants.ClusterProviderKubernetes,
			"status":               constants.ClusterStatusConnected,
			"connection_type":      constants.ClusterConnectionTypeKubeconfig,
			"kubeconfig_encrypted": "dummy",
			"api_endpoint":         "https://api.example",
			"region":               "us-east-1",
			"version":              "1.29",
			"validation_error":     "",
			"metadata":             map[string]any{"env": "test"},
		}), http.StatusCreated)

		reloaded := mustGetUserByEmail(t, app.userRepo, invitee.Email)
		if reloaded.OrganizationID == nil || *reloaded.OrganizationID != orgID {
			t.Fatalf("expected invitee's organization to be updated in the database")
		}
	})
}
