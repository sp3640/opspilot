package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sp3640/opspilot/backend/internal/alerting"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func decodeAlertItems(t *testing.T, env mtAPIResponse) []map[string]any {
	t.Helper()
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode alert list payload: %v; data=%s", err, string(env.Data))
	}
	return payload.Items
}

// TestAlertLifecycleIntegration covers OPEN -> ACKNOWLEDGED -> RESOLVED ->
// (reopen) OPEN, and confirms organization isolation on every transition.
func TestAlertLifecycleIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminAToken := registerAndLogin(t, app.router, "Alert Lifecycle Admin A", "alert-lifecycle-admin-a@opspilot.dev", "password123")
	projectA := createProject(t, app.router, adminAToken, "Alert Lifecycle Project A")
	resourceA := createResource(t, app.router, adminAToken, projectA, "alert-lifecycle-resource-a")

	adminBToken := registerAndLogin(t, app.router, "Alert Lifecycle Admin B", "alert-lifecycle-admin-b@opspilot.dev", "password123")
	adminB := mustGetUserByEmail(t, app.userRepo, "alert-lifecycle-admin-b@opspilot.dev")
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, *adminB.OrganizationID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "alert-lifecycle-admin-b@opspilot.dev", "password123")

	alertID := createAlert(t, app.router, adminAToken, projectA, "Lifecycle test alert", resourceA)

	// Org B cannot see, acknowledge, resolve, or reopen org A's alert.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts/"+itoa(alertID), adminBToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/acknowledge", adminBToken, nil), http.StatusForbidden)

	// OPEN -> ACKNOWLEDGED
	ackRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/acknowledge", adminAToken, nil)
	assertStatus(t, ackRec, http.StatusOK)
	acked := decodeDataMap(t, ackRec)
	if acked["status"] != constants.AlertStatusAcknowledged {
		t.Fatalf("expected status ACKNOWLEDGED, got %v", acked["status"])
	}
	if acked["acknowledgedAt"] == nil {
		t.Fatalf("expected acknowledgedAt to be set")
	}

	// ACKNOWLEDGED -> RESOLVED
	resolveRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/resolve", adminAToken, nil)
	assertStatus(t, resolveRec, http.StatusOK)
	resolved := decodeDataMap(t, resolveRec)
	if resolved["status"] != constants.AlertStatusResolved {
		t.Fatalf("expected status RESOLVED, got %v", resolved["status"])
	}
	if resolved["resolvedAt"] == nil {
		t.Fatalf("expected resolvedAt to be set")
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/reopen", adminBToken, nil), http.StatusForbidden)

	// RESOLVED -> OPEN (reopen)
	reopenRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/reopen", adminAToken, nil)
	assertStatus(t, reopenRec, http.StatusOK)
	reopened := decodeDataMap(t, reopenRec)
	if reopened["status"] != constants.AlertStatusOpen {
		t.Fatalf("expected status OPEN after reopen, got %v", reopened["status"])
	}
	if reopened["resolvedAt"] != nil {
		t.Fatalf("expected resolvedAt to be cleared after reopen, got %v", reopened["resolvedAt"])
	}
}

// TestAlertAuditLogsIntegration covers Phase 15's alert-timeline endpoint:
// every lifecycle transition must be visible in the alert's own audit trail,
// and it must respect organization isolation like every other alert route.
func TestAlertAuditLogsIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminAToken := registerAndLogin(t, app.router, "Alert Audit Admin A", "alert-audit-admin-a@opspilot.dev", "password123")
	projectA := createProject(t, app.router, adminAToken, "Alert Audit Project A")
	resourceA := createResource(t, app.router, adminAToken, projectA, "alert-audit-resource-a")

	adminBToken := registerAndLogin(t, app.router, "Alert Audit Admin B", "alert-audit-admin-b@opspilot.dev", "password123")
	adminB := mustGetUserByEmail(t, app.userRepo, "alert-audit-admin-b@opspilot.dev")
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, *adminB.OrganizationID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "alert-audit-admin-b@opspilot.dev", "password123")

	alertID := createAlert(t, app.router, adminAToken, projectA, "Audit trail test alert", resourceA)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/acknowledge", adminAToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/resolve", adminAToken, nil), http.StatusOK)

	rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts/"+itoa(alertID)+"/audit-logs", adminAToken, nil)
	assertStatus(t, rec, http.StatusOK)
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode audit log payload: %v; data=%s", err, string(env.Data))
	}

	var sawAcknowledged, sawResolved, sawCreate bool
	for _, entry := range payload.Items {
		if entry["entity_type"] != "alert" {
			t.Fatalf("expected entity_type=alert on every row, got %+v", entry)
		}
		if entry["action"] == "CREATE" {
			sawCreate = true
		}
		if entry["field_name"] == "status" && entry["new_value"] == constants.AlertStatusAcknowledged {
			sawAcknowledged = true
		}
		if entry["field_name"] == "status" && entry["new_value"] == constants.AlertStatusResolved {
			sawResolved = true
		}
	}
	if !sawCreate || !sawAcknowledged || !sawResolved {
		t.Fatalf("expected create+acknowledge+resolve transitions in the timeline, got %+v", payload.Items)
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts/"+itoa(alertID)+"/audit-logs", adminBToken, nil), http.StatusForbidden)
}

// TestAlertDeduplication confirms the existing fingerprint-based dedup: an
// identical alert (same project/resource/severity/title) submitted twice
// never produces a second row.
func TestAlertDeduplication(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Alert Dedup Admin", "alert-dedup-admin@opspilot.dev", "password123")
	projectID := createProject(t, app.router, adminToken, "Alert Dedup Project")
	resourceID := createResource(t, app.router, adminToken, projectID, "alert-dedup-resource")

	firstID := createAlert(t, app.router, adminToken, projectID, "Duplicate condition", resourceID)
	secondID := createAlert(t, app.router, adminToken, projectID, "Duplicate condition", resourceID)

	if firstID != secondID {
		t.Fatalf("expected the same alert to be reused, got ids %d and %d", firstID, secondID)
	}

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts/"+itoa(firstID), adminToken, nil)
	assertStatus(t, getRec, http.StatusOK)
	data := decodeDataMap(t, getRec)
	if toUint(t, data["occurrenceCount"]) != 2 {
		t.Fatalf("expected occurrenceCount=2 after a duplicate create, got %v", data["occurrenceCount"])
	}
}

// TestAlertEngineReconciliation drives AlertService.ReconcileConditions -
// the alert-evaluation engine's entry point - directly through a fire /
// re-fire / escalate / clear / re-fire cycle, and asserts at every step that
// exactly one alert row exists for the condition (never a duplicate).
func TestAlertEngineReconciliation(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Alert Engine Admin", "alert-engine-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "alert-engine-admin@opspilot.dev")
	projectID := createProject(t, app.router, adminToken, "Alert Engine Project")

	condition := func(severity string) alerting.EvaluatedCondition {
		return alerting.EvaluatedCondition{
			ResourceType: constants.AlertResourceTypeCluster,
			ResourceID:   "cluster-under-test",
			ResourceName: "cluster-under-test",
			ConditionKey: alerting.ConditionCPUThreshold,
			Title:        "High CPU usage on cluster cluster-under-test",
			Description:  "CPU usage is high",
			Severity:     severity,
			CurrentValue: 85,
			Threshold:    80,
			Unit:         "%",
			ClusterID:    "cluster-under-test",
			ClusterName:  "cluster-under-test",
		}
	}

	listAlerts := func() []map[string]any {
		rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts?projectId="+projectID.String(), adminToken, nil)
		assertStatus(t, rec, http.StatusOK)
		return decodeAlertItems(t, decodeEnvelope(t, rec))
	}

	// First evaluation pass: condition fires -> one new OPEN alert.
	if err := app.alertService.ReconcileConditions(t.Context(), projectID, admin.ID, constants.AlertSourceKubernetes, []alerting.EvaluatedCondition{condition(constants.AlertSeverityHigh)}); err != nil {
		t.Fatalf("reconcile (create): %v", err)
	}
	alerts := listAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected exactly 1 alert after first reconciliation, got %d: %+v", len(alerts), alerts)
	}
	if alerts[0]["status"] != constants.AlertStatusOpen || alerts[0]["severity"] != constants.AlertSeverityHigh {
		t.Fatalf("expected OPEN/HIGH alert, got %+v", alerts[0])
	}
	alertID := toUint(t, alerts[0]["id"])

	// Second pass: same condition, same severity -> no duplicate, occurrence increments.
	if err := app.alertService.ReconcileConditions(t.Context(), projectID, admin.ID, constants.AlertSourceKubernetes, []alerting.EvaluatedCondition{condition(constants.AlertSeverityHigh)}); err != nil {
		t.Fatalf("reconcile (dedup): %v", err)
	}
	alerts = listAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected still exactly 1 alert after re-firing the same condition, got %d: %+v", len(alerts), alerts)
	}
	if toUint(t, alerts[0]["id"]) != alertID {
		t.Fatalf("expected the same alert id to be reused, got %v want %d", alerts[0]["id"], alertID)
	}
	if toUint(t, alerts[0]["occurrenceCount"]) != 2 {
		t.Fatalf("expected occurrenceCount=2, got %v", alerts[0]["occurrenceCount"])
	}

	// Third pass: severity escalates -> same alert row updated in place, not duplicated.
	if err := app.alertService.ReconcileConditions(t.Context(), projectID, admin.ID, constants.AlertSourceKubernetes, []alerting.EvaluatedCondition{condition(constants.AlertSeverityCritical)}); err != nil {
		t.Fatalf("reconcile (escalate): %v", err)
	}
	alerts = listAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected still exactly 1 alert after severity escalation, got %d: %+v", len(alerts), alerts)
	}
	if toUint(t, alerts[0]["id"]) != alertID {
		t.Fatalf("expected escalation to update the same alert id, got %v want %d", alerts[0]["id"], alertID)
	}
	if alerts[0]["severity"] != constants.AlertSeverityCritical {
		t.Fatalf("expected severity to escalate to CRITICAL, got %v", alerts[0]["severity"])
	}

	// Fourth pass: condition clears -> the alert auto-resolves.
	if err := app.alertService.ReconcileConditions(t.Context(), projectID, admin.ID, constants.AlertSourceKubernetes, []alerting.EvaluatedCondition{}); err != nil {
		t.Fatalf("reconcile (clear): %v", err)
	}
	alerts = listAlerts()
	if len(alerts) != 1 || alerts[0]["status"] != constants.AlertStatusResolved {
		t.Fatalf("expected the alert to auto-resolve once the condition clears, got %+v", alerts)
	}

	// Fifth pass: condition reappears -> the resolved alert reopens, no new row.
	if err := app.alertService.ReconcileConditions(t.Context(), projectID, admin.ID, constants.AlertSourceKubernetes, []alerting.EvaluatedCondition{condition(constants.AlertSeverityHigh)}); err != nil {
		t.Fatalf("reconcile (recur): %v", err)
	}
	alerts = listAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected exactly 1 alert after the condition recurs, got %d: %+v", len(alerts), alerts)
	}
	if toUint(t, alerts[0]["id"]) != alertID {
		t.Fatalf("expected the same alert id to be reopened, got %v want %d", alerts[0]["id"], alertID)
	}
	if alerts[0]["status"] != constants.AlertStatusOpen {
		t.Fatalf("expected the recurring condition to reopen the alert, got status %v", alerts[0]["status"])
	}
}

// TestAlertEngineReconciliationOrganizationIsolation confirms that
// reconciling conditions for one organization's project never creates,
// updates, or otherwise touches alerts visible to a different organization.
func TestAlertEngineReconciliationOrganizationIsolation(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminAToken := registerAndLogin(t, app.router, "Alert Isolation Admin A", "alert-isolation-admin-a@opspilot.dev", "password123")
	adminA := mustGetUserByEmail(t, app.userRepo, "alert-isolation-admin-a@opspilot.dev")
	projectA := createProject(t, app.router, adminAToken, "Alert Isolation Project A")

	adminBToken := registerAndLogin(t, app.router, "Alert Isolation Admin B", "alert-isolation-admin-b@opspilot.dev", "password123")
	adminB := mustGetUserByEmail(t, app.userRepo, "alert-isolation-admin-b@opspilot.dev")
	if err := app.userRepo.AssignOrganizationAndRole(adminB.ID, *adminB.OrganizationID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin B role: %v", err)
	}
	adminBToken = loginOnly(t, app.router, "alert-isolation-admin-b@opspilot.dev", "password123")
	projectB := createProject(t, app.router, adminBToken, "Alert Isolation Project B")

	sharedCondition := alerting.EvaluatedCondition{
		ResourceType: constants.AlertResourceTypeCluster,
		ResourceID:   "shared-cluster-id", // deliberately identical across orgs
		ResourceName: "shared-cluster-id",
		ConditionKey: alerting.ConditionCPUThreshold,
		Title:        "High CPU usage on cluster shared-cluster-id",
		Description:  "CPU usage is high",
		Severity:     constants.AlertSeverityHigh,
		CurrentValue: 85,
		Threshold:    80,
		Unit:         "%",
	}

	if err := app.alertService.ReconcileConditions(t.Context(), projectA, adminA.ID, constants.AlertSourceKubernetes, []alerting.EvaluatedCondition{sharedCondition}); err != nil {
		t.Fatalf("reconcile project A: %v", err)
	}

	// Org B must not see org A's alert, even though the underlying
	// condition (same ResourceType/ResourceID/title) is identical.
	listBRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts?projectId="+projectB.String(), adminBToken, nil)
	assertStatus(t, listBRec, http.StatusOK)
	itemsB := decodeAlertItems(t, decodeEnvelope(t, listBRec))
	if len(itemsB) != 0 {
		t.Fatalf("expected organization B to see no alerts, got %+v", itemsB)
	}

	// Reconciling the identical-looking condition for org B's project must
	// create its own independent alert, not touch org A's.
	if err := app.alertService.ReconcileConditions(t.Context(), projectB, adminB.ID, constants.AlertSourceKubernetes, []alerting.EvaluatedCondition{sharedCondition}); err != nil {
		t.Fatalf("reconcile project B: %v", err)
	}

	listARec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts?projectId="+projectA.String(), adminAToken, nil)
	assertStatus(t, listARec, http.StatusOK)
	itemsA := decodeAlertItems(t, decodeEnvelope(t, listARec))
	if len(itemsA) != 1 {
		t.Fatalf("expected organization A to still have exactly 1 alert, got %+v", itemsA)
	}
	if toUint(t, itemsA[0]["occurrenceCount"]) != 1 {
		t.Fatalf("expected org A's alert occurrence count to be untouched by org B's reconciliation, got %v", itemsA[0]["occurrenceCount"])
	}

	listB2Rec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts?projectId="+projectB.String(), adminBToken, nil)
	assertStatus(t, listB2Rec, http.StatusOK)
	itemsB2 := decodeAlertItems(t, decodeEnvelope(t, listB2Rec))
	if len(itemsB2) != 1 {
		t.Fatalf("expected organization B to now have exactly 1 alert of its own, got %+v", itemsB2)
	}
}
