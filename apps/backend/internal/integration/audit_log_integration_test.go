package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// TestAuditLoginIntegration confirms Phase 23's login auditing rules: a
// successful login and a failed login against a KNOWN account both produce
// an audit row (distinguished by Result), while a failed login against an
// email with no account produces none - there is no organization to attach
// it to, and inventing one would violate organization isolation.
func TestAuditLoginIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "Audit Admin", "audit-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "audit-admin@opspilot.dev")

	// Failed login against a real account.
	failRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email":    "audit-admin@opspilot.dev",
		"password": "wrong-password",
	})
	assertStatus(t, failRec, http.StatusUnauthorized)

	// Failed login against an email with no account at all.
	unknownRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email":    "nobody-at-all@opspilot.dev",
		"password": "whatever123",
	})
	assertStatus(t, unknownRec, http.StatusUnauthorized)

	items := fetchAllAuditItems(t, app.router, adminToken)

	var successCount, failureCount int
	for _, item := range items {
		if item["entity_type"] != "auth" {
			continue
		}
		switch item["result"] {
		case "SUCCESS":
			successCount++
		case "FAILURE":
			failureCount++
		}
		if uint64(item["user_id"].(float64)) != uint64(admin.ID) {
			t.Fatalf("expected auth audit entry to attribute to admin, got %v", item["user_id"])
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful login audit entry (from registerAndLogin), got %d", successCount)
	}
	if failureCount != 1 {
		t.Fatalf("expected exactly 1 failed login audit entry (known account only), got %d", failureCount)
	}
}

// TestAuditInvitationTrailIntegration verifies invitation create/revoke are
// audited, that the granted role is captured (standing in for "role change"
// since there is no separate role-change endpoint), and that the raw
// invitation token never appears in any audit record.
func TestAuditInvitationTrailIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "Invite Admin", "invite-admin@opspilot.dev", "password123")

	inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
		"email": "invitee@opspilot.dev",
		"role":  string(models.RoleViewer),
	})
	assertStatus(t, inviteRec, http.StatusCreated)
	invitationData := decodeDataMap(t, inviteRec)
	invitationID, _ := invitationData["id"].(string)
	token, _ := invitationData["token"].(string)
	if invitationID == "" || token == "" {
		t.Fatalf("expected invitation id and token, got %v", invitationData)
	}

	revokeRec := doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/invitations/"+invitationID, adminToken, nil)
	assertStatus(t, revokeRec, http.StatusOK)

	items := fetchAllAuditItems(t, app.router, adminToken)

	var foundCreate, foundRevoke bool
	for _, item := range items {
		if item["entity_type"] != "invitation" {
			continue
		}
		afterState, _ := item["after_state"].(string)
		beforeState, _ := item["before_state"].(string)
		oldValue, _ := item["old_value"].(string)
		newValue, _ := item["new_value"].(string)
		if strings.Contains(afterState, token) || strings.Contains(beforeState, token) ||
			strings.Contains(oldValue, token) || strings.Contains(newValue, token) {
			t.Fatalf("invitation token leaked into audit record: %+v", item)
		}
		if item["action"] == "CREATE" {
			foundCreate = true
			if !strings.Contains(afterState, "invitee@opspilot.dev") {
				t.Fatalf("expected create audit AfterState to capture invited email, got %q", afterState)
			}
		}
		if item["action"] == "UPDATE" && item["field_name"] == "status" {
			foundRevoke = true
			if newValue != models.InvitationStatusRevoked {
				t.Fatalf("expected revoke audit to record status=%s, got %q", models.InvitationStatusRevoked, newValue)
			}
		}
	}

	if !foundCreate {
		t.Fatalf("expected an invitation create audit entry")
	}
	if !foundRevoke {
		t.Fatalf("expected an invitation revoke audit entry")
	}
}

// TestAuditApplicationTrailIntegration verifies application CRUD produces
// before/after-state audit entries with organization/application isolation.
func TestAuditApplicationTrailIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "App Audit Admin", "app-audit-admin@opspilot.dev", "password123")

	projectID := createProject(t, app.router, adminToken, "Audit Project")

	createRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "Audit App",
		"runtime": constants.ApplicationRuntimeGo,
		"port":    8080,
	})
	assertStatus(t, createRec, http.StatusCreated)
	applicationData := decodeDataMap(t, createRec)
	applicationID, _ := applicationData["id"].(string)

	updateRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/applications/"+applicationID, adminToken, map[string]any{
		"name":    "Audit App Renamed",
		"runtime": constants.ApplicationRuntimeGo,
		"port":    9090,
	})
	assertStatus(t, updateRec, http.StatusOK)

	deleteRec := doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/applications/"+applicationID, adminToken, nil)
	assertStatus(t, deleteRec, http.StatusOK)

	items := fetchAllAuditItems(t, app.router, adminToken)

	var foundCreate, foundUpdate, foundDelete bool
	for _, item := range items {
		if item["entity_type"] != "application" || item["entity_id"] != applicationID {
			continue
		}
		afterState, _ := item["after_state"].(string)
		beforeState, _ := item["before_state"].(string)
		switch item["action"] {
		case "CREATE":
			foundCreate = true
			if !strings.Contains(afterState, "Audit App") {
				t.Fatalf("expected create AfterState to snapshot application name, got %q", afterState)
			}
		case "UPDATE":
			foundUpdate = true
			if !strings.Contains(beforeState, "Audit App") || !strings.Contains(afterState, "Audit App Renamed") {
				t.Fatalf("expected update before/after snapshots to capture the rename, got before=%q after=%q", beforeState, afterState)
			}
		case "DELETE":
			foundDelete = true
			if !strings.Contains(beforeState, "Audit App Renamed") {
				t.Fatalf("expected delete BeforeState to snapshot the final application state, got %q", beforeState)
			}
		}
	}

	if !foundCreate || !foundUpdate || !foundDelete {
		t.Fatalf("expected create+update+delete audit entries for application, got create=%v update=%v delete=%v", foundCreate, foundUpdate, foundDelete)
	}
}

// TestOrganizationAuditEndpointIntegration exercises the new GET
// /api/v1/audit-logs (Phase 23 Organization Audit view): membership gating,
// organization isolation, and the new filter dimensions.
func TestOrganizationAuditEndpointIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminAToken := registerAndLogin(t, app.router, "Org Audit Admin A", "org-audit-admin-a@opspilot.dev", "password123")
	adminA := mustGetUserByEmail(t, app.userRepo, "org-audit-admin-a@opspilot.dev")

	adminBToken := registerAndLogin(t, app.router, "Org Audit Admin B", "org-audit-admin-b@opspilot.dev", "password123")

	createProject(t, app.router, adminAToken, "Org Audit Project")

	listRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/audit-logs", adminAToken, nil)
	assertStatus(t, listRec, http.StatusOK)
	items := decodeAuditItems(t, listRec)

	var foundProjectCreate bool
	for _, item := range items {
		if item["entity_type"] == "project" && item["action"] == "CREATE" {
			foundProjectCreate = true
		}
	}
	if !foundProjectCreate {
		t.Fatalf("expected organization audit view to include the project create entry, got %v", items)
	}

	// Organization isolation: org B's admin must never see org A's entries,
	// even though both hit the same unscoped endpoint.
	crossOrgRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/audit-logs", adminBToken, nil)
	assertStatus(t, crossOrgRec, http.StatusOK)
	for _, item := range decodeAuditItems(t, crossOrgRec) {
		if item["entity_type"] == "project" {
			t.Fatalf("org B must not see org A's project audit entries, got %v", item)
		}
	}

	// Filter by action.
	actionFilterRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/audit-logs?action=CREATE", adminAToken, nil)
	assertStatus(t, actionFilterRec, http.StatusOK)
	for _, item := range decodeAuditItems(t, actionFilterRec) {
		if item["action"] != "CREATE" {
			t.Fatalf("expected only CREATE entries with action filter, got %v", item)
		}
	}

	// Filter by entity type.
	entityFilterRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/audit-logs?entityType=project", adminAToken, nil)
	assertStatus(t, entityFilterRec, http.StatusOK)
	entityItems := decodeAuditItems(t, entityFilterRec)
	if len(entityItems) == 0 {
		t.Fatalf("expected at least one project entry with entityType filter")
	}
	for _, item := range entityItems {
		if item["entity_type"] != "project" {
			t.Fatalf("expected only project entries with entityType filter, got %v", item)
		}
	}

	// Filter by userId.
	userFilterRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/audit-logs?userId="+strconv.FormatUint(uint64(adminA.ID), 10), adminAToken, nil)
	assertStatus(t, userFilterRec, http.StatusOK)
	if len(decodeAuditItems(t, userFilterRec)) == 0 {
		t.Fatalf("expected at least one entry for adminA's own userId filter")
	}

	// Filter by result.
	resultFilterRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/audit-logs?result=SUCCESS", adminAToken, nil)
	assertStatus(t, resultFilterRec, http.StatusOK)
	for _, item := range decodeAuditItems(t, resultFilterRec) {
		if item["result"] != "SUCCESS" {
			t.Fatalf("expected only SUCCESS entries with result filter, got %v", item)
		}
	}

	// Filter by date range - a window safely in the past should exclude
	// everything that just happened.
	past := time.Now().Add(-48 * time.Hour)
	dateFilterRec := doJSONRequest(t, app.router, http.MethodGet,
		"/api/v1/audit-logs?dateFrom="+url.QueryEscape(past.Add(-time.Hour).Format(time.RFC3339))+"&dateTo="+url.QueryEscape(past.Format(time.RFC3339)),
		adminAToken, nil)
	assertStatus(t, dateFilterRec, http.StatusOK)
	if len(decodeAuditItems(t, dateFilterRec)) != 0 {
		t.Fatalf("expected no entries in a date range before any activity occurred")
	}

	// Invalid filter values are rejected, not silently ignored.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/audit-logs?userId=not-a-number", adminAToken, nil), http.StatusBadRequest)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/audit-logs?dateFrom=not-a-date", adminAToken, nil), http.StatusBadRequest)
}

// TestAuditLogsAreImmutableIntegration confirms there is no route capable of
// mutating an audit log - immutability holds by construction (no
// PUT/PATCH/DELETE handler exists for the resource at all), not by an
// after-the-fact permission check.
func TestAuditLogsAreImmutableIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Immutable Admin", "immutable-admin@opspilot.dev", "password123")

	for _, method := range []string{http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodPost} {
		rec := doJSONRequest(t, app.router, method, "/api/v1/audit-logs", adminToken, map[string]any{"entity_type": "tamper"})
		if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected %s /audit-logs to be unroutable, got status %d", method, rec.Code)
		}
	}

	rec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/audit-logs/1", adminToken, map[string]any{"entity_type": "tamper"})
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected PUT /audit-logs/1 to be unroutable, got status %d", rec.Code)
	}
}

func decodeAuditItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode audit list payload: %v; data=%s", err, string(env.Data))
	}
	return payload.Items
}

func fetchAllAuditItems(t *testing.T, r *gin.Engine, token string) []map[string]any {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodGet, "/api/v1/audit-logs?limit=100", token, nil)
	assertStatus(t, rec, http.StatusOK)
	return decodeAuditItems(t, rec)
}
