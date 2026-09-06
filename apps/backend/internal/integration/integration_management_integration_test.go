package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

const integrationSuperSecretTestToken = "SUPER_SECRET_TEST_TOKEN"

func TestIntegrationCRUDIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Integration Admin", "integration-admin@opspilot.dev", "password123")

	createRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", adminToken, map[string]any{
		// Slack (not GitHub) here deliberately: Sprint 28 registered a real
		// GitHub connector, so this generic Sprint 27 CRUD test uses a type
		// that is still genuinely unimplemented to exercise that path.
		"type":     constants.IntegrationTypeSlack,
		"name":     "Primary Slack",
		"metadata": map[string]any{"organization": "opspilot"},
		"credentials": map[string]any{
			"token": integrationSuperSecretTestToken,
		},
	})
	assertStatus(t, createRec, http.StatusCreated)
	created := decodeDataMap(t, createRec)

	integrationID, _ := created["id"].(string)
	if integrationID == "" {
		t.Fatalf("expected integration id in response, got %v", created)
	}
	if created["status"] != constants.IntegrationStatusPending {
		t.Fatalf("expected new integration status PENDING, got %v", created["status"])
	}
	if created["hasCredentials"] != true {
		t.Fatalf("expected hasCredentials=true, got %v", created["hasCredentials"])
	}
	if created["connectorImplemented"] != false {
		t.Fatalf("expected connectorImplemented=false (no connector implemented for slack), got %v", created["connectorImplemented"])
	}
	assertNoSecretLeak(t, createRec.Body.String())

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID, adminToken, nil)
	assertStatus(t, getRec, http.StatusOK)
	assertNoSecretLeak(t, getRec.Body.String())

	listRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations", adminToken, nil)
	assertStatus(t, listRec, http.StatusOK)
	assertNoSecretLeak(t, listRec.Body.String())
	listItems := decodeIntegrationItems(t, listRec)
	found := false
	for _, item := range listItems {
		if item["id"] == integrationID {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected created integration to appear in list, got %v", listItems)
	}

	updateRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/integrations/"+integrationID, adminToken, map[string]any{
		"name":     "Renamed GitHub",
		"metadata": map[string]any{"organization": "opspilot-renamed"},
		"credentials": map[string]any{
			"token": "ROTATED_" + integrationSuperSecretTestToken,
		},
	})
	assertStatus(t, updateRec, http.StatusOK)
	updated := decodeDataMap(t, updateRec)
	if updated["name"] != "Renamed GitHub" {
		t.Fatalf("expected renamed integration, got %v", updated["name"])
	}
	assertNoSecretLeak(t, updateRec.Body.String())

	deleteRec := doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/integrations/"+integrationID, adminToken, nil)
	assertStatus(t, deleteRec, http.StatusOK)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID, adminToken, nil), http.StatusNotFound)
}

func TestIntegrationRejectsInvalidTypeAndDuplicate(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Integration Validate Admin", "integration-validate-admin@opspilot.dev", "password123")

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", adminToken, map[string]any{
		"type": "not-a-real-type",
		"name": "Bad Type",
	}), http.StatusBadRequest)

	firstRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", adminToken, map[string]any{
		"type": constants.IntegrationTypeSlack,
		"name": "Ops Slack",
	})
	assertStatus(t, firstRec, http.StatusCreated)

	duplicateRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", adminToken, map[string]any{
		"type": constants.IntegrationTypeSlack,
		"name": "Ops Slack",
	})
	assertStatus(t, duplicateRec, http.StatusConflict)
}

func TestIntegrationTestAndCheckReportUnsupportedConnector(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Integration Check Admin", "integration-check-admin@opspilot.dev", "password123")

	createRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", adminToken, map[string]any{
		"type": constants.IntegrationTypePrometheus,
		"name": "Prod Prometheus",
		"credentials": map[string]any{
			"apiKey": integrationSuperSecretTestToken,
		},
	})
	assertStatus(t, createRec, http.StatusCreated)
	integrationID, _ := decodeDataMap(t, createRec)["id"].(string)

	// No connector is implemented for any type this sprint - test/check
	// must fail honestly, never fabricate a successful connection.
	testRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/test", adminToken, nil)
	assertStatus(t, testRec, http.StatusOK)
	testResult := decodeDataMap(t, testRec)
	if testResult["success"] != false {
		t.Fatalf("expected test connection to report success=false for an unimplemented connector, got %v", testResult)
	}
	assertNoSecretLeak(t, testRec.Body.String())

	checkRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/check", adminToken, nil)
	assertStatus(t, checkRec, http.StatusOK)
	checkResult := decodeDataMap(t, checkRec)
	if checkResult["success"] != false {
		t.Fatalf("expected health check to report success=false for an unimplemented connector, got %v", checkResult)
	}

	integrationField, ok := checkResult["integration"].(map[string]any)
	if !ok {
		t.Fatalf("expected check response to embed the updated integration, got %v", checkResult)
	}
	if integrationField["status"] != constants.IntegrationStatusError {
		t.Fatalf("expected status ERROR after a failed check, got %v", integrationField["status"])
	}
	if integrationField["lastError"] == "" || integrationField["lastError"] == nil {
		t.Fatalf("expected a non-empty lastError after a failed check")
	}
	assertNoSecretLeak(t, checkRec.Body.String())

	// GET must reflect the same persisted status/lastError.
	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID, adminToken, nil)
	assertStatus(t, getRec, http.StatusOK)
	getData := decodeDataMap(t, getRec)
	if getData["status"] != constants.IntegrationStatusError {
		t.Fatalf("expected persisted status ERROR, got %v", getData["status"])
	}
}

func TestIntegrationCrossOrganizationAccessDenied(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminAToken := registerAndLogin(t, app.router, "Integration Org A Admin", "integration-org-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "Integration Org B Admin", "integration-org-b@opspilot.dev", "password123")

	createRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", adminAToken, map[string]any{
		"type": constants.IntegrationTypeWebhook,
		"name": "Org A Webhook",
	})
	assertStatus(t, createRec, http.StatusCreated)
	integrationID, _ := decodeDataMap(t, createRec)["id"].(string)

	// Org B must never see, modify, or delete org A's integration.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID, adminBToken, nil), http.StatusNotFound)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/integrations/"+integrationID, adminBToken, map[string]any{
		"name": "Hijacked",
	}), http.StatusNotFound)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/integrations/"+integrationID, adminBToken, nil), http.StatusNotFound)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/test", adminBToken, nil), http.StatusNotFound)

	listRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations", adminBToken, nil)
	assertStatus(t, listRec, http.StatusOK)
	for _, item := range decodeIntegrationItems(t, listRec) {
		if item["id"] == integrationID {
			t.Fatalf("org B must not see org A's integration in its own list")
		}
	}
}

func TestIntegrationRBACDeniesNonAdminMutations(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Integration RBAC Admin", "integration-rbac-admin@opspilot.dev", "password123")
	registerAndLogin(t, app.router, "Integration RBAC Dev", "integration-rbac-dev@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "integration-rbac-admin@opspilot.dev")
	dev := mustGetUserByEmail(t, app.userRepo, "integration-rbac-dev@opspilot.dev")
	orgID := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(dev.ID, orgID, models.RoleDevOpsEngineer); err != nil {
		t.Fatalf("assign devops engineer role: %v", err)
	}
	devToken := loginOnly(t, app.router, "integration-rbac-dev@opspilot.dev", "password123")

	// DevOps Engineer can read (organization:read) but must not manage
	// integrations (organization:manage is Platform-Admin-only) - reusing
	// the existing organization permission family, not a new one.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations", devToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", devToken, map[string]any{
		"type": constants.IntegrationTypeWebhook,
		"name": "Dev Attempt",
	}), http.StatusForbidden)

	createRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", adminToken, map[string]any{
		"type": constants.IntegrationTypeWebhook,
		"name": "Admin Webhook",
	})
	assertStatus(t, createRec, http.StatusCreated)
	integrationID, _ := decodeDataMap(t, createRec)["id"].(string)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/integrations/"+integrationID, devToken, map[string]any{"name": "Nope"}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/integrations/"+integrationID, devToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/test", devToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/check", devToken, nil), http.StatusForbidden)
}

// TestIntegrationSecretNeverLeaks is the sprint's headline security test: a
// fake secret must never surface through any read path or audit record.
func TestIntegrationSecretNeverLeaks(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Integration Secret Admin", "integration-secret-admin@opspilot.dev", "password123")

	createRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", adminToken, map[string]any{
		"type": constants.IntegrationTypeGitHub,
		"name": "Secret Test Integration",
		"credentials": map[string]any{
			"token": integrationSuperSecretTestToken,
		},
	})
	assertStatus(t, createRec, http.StatusCreated)
	assertNoSecretLeak(t, createRec.Body.String())
	integrationID, _ := decodeDataMap(t, createRec)["id"].(string)

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID, adminToken, nil)
	assertNoSecretLeak(t, getRec.Body.String())

	listRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations", adminToken, nil)
	assertNoSecretLeak(t, listRec.Body.String())

	updateRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/integrations/"+integrationID, adminToken, map[string]any{
		"name":        "Secret Test Integration Renamed",
		"credentials": map[string]any{"token": "ROTATED_" + integrationSuperSecretTestToken},
	})
	assertNoSecretLeak(t, updateRec.Body.String())

	testRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/test", adminToken, nil)
	assertNoSecretLeak(t, testRec.Body.String())

	// The audit trail: neither the original nor the rotated secret, in any
	// field (before_state/after_state/old_value/new_value), anywhere.
	auditItems := fetchAllAuditItems(t, app.router, adminToken)
	for _, item := range auditItems {
		encoded, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("marshal audit item: %v", err)
		}
		assertNoSecretLeak(t, string(encoded))
	}

	// Direct DB inspection: EncryptedCredentials must not be plaintext and
	// must not contain the secret either (it's ciphertext, but this also
	// catches an accidental no-op "encryption").
	parsedID, err := uuid.Parse(integrationID)
	if err != nil {
		t.Fatalf("parse integration id: %v", err)
	}
	admin := mustGetUserByEmail(t, app.userRepo, "integration-secret-admin@opspilot.dev")
	stored, err := app.integrationRepo.GetByID(parsedID, *admin.OrganizationID)
	if err != nil {
		t.Fatalf("load stored integration: %v", err)
	}
	if strings.Contains(stored.EncryptedCredentials, integrationSuperSecretTestToken) {
		t.Fatalf("expected EncryptedCredentials to never contain the plaintext secret")
	}
}

func assertNoSecretLeak(t *testing.T, body string) {
	t.Helper()
	if strings.Contains(body, integrationSuperSecretTestToken) {
		t.Fatalf("secret leaked into response/audit body: %s", body)
	}
}

func decodeIntegrationItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode integration list payload: %v; data=%s", err, string(env.Data))
	}
	return payload.Items
}
