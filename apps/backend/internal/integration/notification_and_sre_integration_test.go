package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// capturingWebhookServer records every JSON POST body it receives, for
// asserting a notification was actually dispatched to a channel's target.
type capturingWebhookServer struct {
	server *httptest.Server
	mu     sync.Mutex
	bodies []map[string]any
}

func newCapturingWebhookServer(t *testing.T) *capturingWebhookServer {
	t.Helper()
	c := &capturingWebhookServer{}
	c.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		c.mu.Lock()
		c.bodies = append(c.bodies, body)
		c.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(c.server.Close)
	return c
}

func (c *capturingWebhookServer) received(event string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, body := range c.bodies {
		if body["event"] == event {
			return true
		}
	}
	return false
}

func (c *capturingWebhookServer) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.bodies)
}

func createNotificationChannel(t *testing.T, r *gin.Engine, token, name, channelType, target string, events []string, teamID *uuid.UUID) map[string]any {
	t.Helper()
	body := map[string]any{
		"name":   name,
		"type":   channelType,
		"target": target,
		"events": events,
	}
	if teamID != nil {
		body["team_id"] = teamID.String()
	}
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/notification-channels", token, body)
	assertStatus(t, rec, http.StatusCreated)
	return decodeDataMap(t, rec)
}

func TestNotificationChannelCRUDIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Notif Admin", "notif-admin@opspilot.dev", "password123")

	webhook := newCapturingWebhookServer(t)

	created := createNotificationChannel(t, app.router, adminToken, "Ops Webhook", constants.NotificationChannelWebhook, webhook.server.URL, []string{constants.NotificationEventCriticalAlert}, nil)
	channelID, _ := created["id"].(string)
	if channelID == "" {
		t.Fatalf("expected channel id in response, got %v", created)
	}
	maskedTarget, _ := created["maskedTarget"].(string)
	if maskedTarget == "" || maskedTarget == webhook.server.URL {
		t.Fatalf("expected a masked target, got %q", maskedTarget)
	}
	if created["enabled"] != true {
		t.Fatalf("expected new channel to be enabled by default")
	}

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/notification-channels/"+channelID, adminToken, nil)
	assertStatus(t, getRec, http.StatusOK)

	listRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/notification-channels", adminToken, nil)
	assertStatus(t, listRec, http.StatusOK)

	updateRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/notification-channels/"+channelID, adminToken, map[string]any{
		"name":    "Ops Webhook Renamed",
		"events":  []string{constants.NotificationEventCriticalAlert, constants.NotificationEventDeploymentFailed},
		"enabled": false,
	})
	assertStatus(t, updateRec, http.StatusOK)
	updated := decodeDataMap(t, updateRec)
	if updated["name"] != "Ops Webhook Renamed" {
		t.Fatalf("expected renamed channel, got %v", updated["name"])
	}
	if updated["enabled"] != false {
		t.Fatalf("expected channel to be disabled after update")
	}

	deleteRec := doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/notification-channels/"+channelID, adminToken, nil)
	assertStatus(t, deleteRec, http.StatusOK)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/notification-channels/"+channelID, adminToken, nil), http.StatusNotFound)
}

func TestNotificationChannelRejectsInvalidTargetAndEvents(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Notif Validate Admin", "notif-validate-admin@opspilot.dev", "password123")

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/notification-channels", adminToken, map[string]any{
		"name": "Bad Slack", "type": constants.NotificationChannelSlack, "target": "not-a-url", "events": []string{constants.NotificationEventCriticalAlert},
	}), http.StatusBadRequest)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/notification-channels", adminToken, map[string]any{
		"name": "Bad Email", "type": constants.NotificationChannelEmail, "target": "not-an-email", "events": []string{constants.NotificationEventCriticalAlert},
	}), http.StatusBadRequest)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/notification-channels", adminToken, map[string]any{
		"name": "Bad Events", "type": constants.NotificationChannelWebhook, "target": "https://example.com/hook", "events": []string{"NOT_A_REAL_EVENT"},
	}), http.StatusBadRequest)
}

func TestNotificationChannelManagePermissionIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Notif Perm Admin", "notif-perm-admin@opspilot.dev", "password123")
	registerAndLogin(t, app.router, "Notif Perm Dev", "notif-perm-dev@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "notif-perm-admin@opspilot.dev")
	dev := mustGetUserByEmail(t, app.userRepo, "notif-perm-dev@opspilot.dev")
	orgID := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(dev.ID, orgID, models.RoleDeveloper); err != nil {
		t.Fatalf("assign developer role: %v", err)
	}
	devToken := loginOnly(t, app.router, "notif-perm-dev@opspilot.dev", "password123")

	// Developer has notification:read but not notification:manage.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/notification-channels", devToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/notification-channels", devToken, map[string]any{
		"name": "Dev Attempt", "type": constants.NotificationChannelWebhook, "target": "https://example.com/hook", "events": []string{constants.NotificationEventCriticalAlert},
	}), http.StatusForbidden)

	// Platform Admin holds notification:manage and can create the same channel.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/notification-channels", adminToken, map[string]any{
		"name": "Admin Channel", "type": constants.NotificationChannelWebhook, "target": "https://example.com/hook", "events": []string{constants.NotificationEventCriticalAlert},
	}), http.StatusCreated)
}

func TestNotificationDispatchOnCriticalAlertIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Dispatch Admin", "dispatch-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "dispatch-admin@opspilot.dev")

	webhook := newCapturingWebhookServer(t)
	createNotificationChannel(t, app.router, adminToken, "Critical Alerts", constants.NotificationChannelWebhook, webhook.server.URL, []string{constants.NotificationEventCriticalAlert}, nil)

	projectID := createProject(t, app.router, adminToken, "Dispatch Project")
	resourceID := createResource(t, app.router, adminToken, projectID, "dispatch-resource")

	now := time.Now().UTC().Format(time.RFC3339)
	rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts", adminToken, map[string]any{
		"project_id":       projectID.String(),
		"title":            "CPU exhausted",
		"description":      "critical",
		"severity":         constants.AlertSeverityCritical,
		"status":           constants.AlertStatusOpen,
		"source":           constants.AlertSourceKubernetes,
		"resource_type":    constants.AlertResourceTypeDeployment,
		"resource_id":      resourceID.String(),
		"fingerprint":      "unused-by-service",
		"occurrence_count": 1,
		"labels":           map[string]any{"app": "dispatch"},
		"metadata":         map[string]any{"env": "test"},
		"first_seen_at":    now,
		"last_seen_at":     now,
	})
	assertStatus(t, rec, http.StatusCreated)

	if !webhook.received(constants.NotificationEventCriticalAlert) {
		t.Fatalf("expected webhook to receive a CRITICAL_ALERT dispatch")
	}

	items := fetchAllAuditItems(t, app.router, adminToken)
	var foundDispatchAudit bool
	for _, item := range items {
		if item["entity_type"] == "notification_dispatch" && item["new_value"] == constants.NotificationEventCriticalAlert {
			foundDispatchAudit = true
			if item["result"] != "SUCCESS" {
				t.Fatalf("expected dispatch audit result SUCCESS, got %v", item["result"])
			}
		}
	}
	if !foundDispatchAudit {
		t.Fatalf("expected an audit entry recording the notification dispatch attempt")
	}
	_ = admin
}

func TestNotificationDispatchOnSev1IncidentIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Sev1 Admin", "sev1-admin@opspilot.dev", "password123")

	webhook := newCapturingWebhookServer(t)
	createNotificationChannel(t, app.router, adminToken, "Sev1 Pager", constants.NotificationChannelWebhook, webhook.server.URL, []string{constants.NotificationEventSev1Incident}, nil)

	projectID := createProject(t, app.router, adminToken, "Sev1 Project")

	rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents", adminToken, map[string]any{
		"title":       "Total outage",
		"description": "sev1",
		"severity":    constants.SeverityP0,
		"status":      constants.StatusOpen,
		"project_id":  projectID.String(),
	})
	assertStatus(t, rec, http.StatusCreated)

	if !webhook.received(constants.NotificationEventSev1Incident) {
		t.Fatalf("expected webhook to receive a SEV1_INCIDENT dispatch")
	}
}

func TestIncidentAssignAndAcknowledgeIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Assign Admin", "assign-admin@opspilot.dev", "password123")
	registerAndLogin(t, app.router, "Assign Target", "assign-target@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "assign-admin@opspilot.dev")
	target := mustGetUserByEmail(t, app.userRepo, "assign-target@opspilot.dev")
	orgID := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(target.ID, orgID, models.RoleViewer); err != nil {
		t.Fatalf("assign target to org: %v", err)
	}

	projectID := createProject(t, app.router, adminToken, "Assign Project")
	incidentID := createIncident(t, app.router, adminToken, projectID, "Needs an owner")

	assignRec := doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/incidents/"+strconv.FormatUint(uint64(incidentID), 10)+"/assign", adminToken, map[string]any{
		"assignee_user_id": target.ID,
	})
	assertStatus(t, assignRec, http.StatusOK)
	assigned := decodeDataMap(t, assignRec)
	if uint64(assigned["assigneeId"].(float64)) != uint64(target.ID) {
		t.Fatalf("expected assigneeId %d, got %v", target.ID, assigned["assigneeId"])
	}

	// Assigning to a user outside the organization must fail.
	registerAndLogin(t, app.router, "Outsider", "assign-outsider@opspilot.dev", "password123")
	outsider := mustGetUserByEmail(t, app.userRepo, "assign-outsider@opspilot.dev")
	badAssignRec := doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/incidents/"+strconv.FormatUint(uint64(incidentID), 10)+"/assign", adminToken, map[string]any{
		"assignee_user_id": outsider.ID,
	})
	assertStatus(t, badAssignRec, http.StatusBadRequest)

	ackRec := doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/incidents/"+strconv.FormatUint(uint64(incidentID), 10)+"/acknowledge", adminToken, nil)
	assertStatus(t, ackRec, http.StatusOK)
	acked := decodeDataMap(t, ackRec)
	firstAckAt, _ := acked["acknowledgedAt"].(string)
	if firstAckAt == "" {
		t.Fatalf("expected acknowledgedAt to be set")
	}

	// Acknowledging again is idempotent - the timestamp must not move.
	ackAgainRec := doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/incidents/"+strconv.FormatUint(uint64(incidentID), 10)+"/acknowledge", adminToken, nil)
	assertStatus(t, ackAgainRec, http.StatusOK)
	ackedAgain := decodeDataMap(t, ackAgainRec)
	if ackedAgain["acknowledgedAt"] != firstAckAt {
		t.Fatalf("expected acknowledgedAt to stay stable across repeated acknowledgement, got %v then %v", firstAckAt, ackedAgain["acknowledgedAt"])
	}
}

func TestApplicationSLOIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "SLO Admin", "slo-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "slo-admin@opspilot.dev")
	orgID := *admin.OrganizationID

	projectID := createProject(t, app.router, adminToken, "SLO Project")

	appRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "SLO App",
		"runtime": constants.ApplicationRuntimeGo,
		"port":    8080,
	})
	assertStatus(t, appRec, http.StatusCreated)
	applicationData := decodeDataMap(t, appRec)
	applicationIDStr, _ := applicationData["id"].(string)
	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		t.Fatalf("parse application id: %v", err)
	}

	// SLO metrics/config require real history - before any is configured,
	// GetSLO reports "not configured" rather than a fabricated default.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationIDStr+"/slo", adminToken, nil), http.StatusNotFound)

	// Backdate the application by 10 days so a 10-day SLO window has real,
	// non-trivial observed history to measure against (an application
	// created moments ago would otherwise clip the window to near-zero).
	backdatedCreatedAt := time.Now().UTC().Add(-10 * 24 * time.Hour)
	if err := app.db.Model(&struct{}{}).Table("applications").Where("id = ?", applicationID).Update("created_at", backdatedCreatedAt).Error; err != nil {
		t.Fatalf("backdate application created_at: %v", err)
	}

	// One P0 (downtime-counting) incident lasting exactly 1 hour, resolved
	// and acknowledged with known offsets, entirely inside the window.
	incidentStart := backdatedCreatedAt.Add(24 * time.Hour)
	incidentAck := incidentStart.Add(10 * time.Minute)
	incidentResolved := incidentStart.Add(70 * time.Minute)
	incident := &models.Incident{
		OrganizationID: orgID,
		Title:          "Backdated outage",
		Description:    "fixture",
		Severity:       constants.SeverityP0,
		Status:         constants.StatusResolved,
		ProjectID:      projectID,
		ApplicationID:  &applicationID,
		UserID:         admin.ID,
		AcknowledgedAt: &incidentAck,
		ResolvedAt:     &incidentResolved,
	}
	if err := app.incidentRepo.Create(incident); err != nil {
		t.Fatalf("create fixture incident: %v", err)
	}
	if err := app.db.Model(&models.Incident{}).Where("id = ?", incident.ID).Update("created_at", incidentStart).Error; err != nil {
		t.Fatalf("backdate incident created_at: %v", err)
	}

	configureRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/applications/"+applicationIDStr+"/slo", adminToken, map[string]any{
		"target_percentage": 99.0,
		"window_days":       10,
	})
	assertStatus(t, configureRec, http.StatusOK)
	config := decodeDataMap(t, configureRec)
	if config["targetPercentage"] != 99.0 {
		t.Fatalf("expected target percentage to round-trip, got %v", config["targetPercentage"])
	}

	getConfigRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationIDStr+"/slo", adminToken, nil)
	assertStatus(t, getConfigRec, http.StatusOK)

	metricsRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationIDStr+"/slo/metrics", adminToken, nil)
	assertStatus(t, metricsRec, http.StatusOK)
	metrics := decodeDataMap(t, metricsRec)

	if metrics["incidentCount"].(float64) != 1 {
		t.Fatalf("expected exactly 1 incident counted in the window, got %v", metrics["incidentCount"])
	}

	availability := metrics["availability"].(map[string]any)
	if availability["available"] != true {
		t.Fatalf("expected availability to be computable, got %v", availability)
	}
	// windowSeconds = 10*86400 = 864000, downtime = 4200s (70m P0 incident)
	// availability = (864000-4200)/864000*100 = 99.5138...%
	availabilityValue := availability["value"].(float64)
	if availabilityValue < 99.45 || availabilityValue > 99.55 {
		t.Fatalf("expected availability ~99.51%%, got %v", availabilityValue)
	}

	mttr := metrics["mttr"].(map[string]any)
	if mttr["available"] != true || mttr["value"].(float64) != 70 {
		t.Fatalf("expected MTTR of 70 minutes, got %v", mttr)
	}

	mtta := metrics["mtta"].(map[string]any)
	if mtta["available"] != true || mtta["value"].(float64) != 10 {
		t.Fatalf("expected MTTA of 10 minutes, got %v", mtta)
	}

	compliance := metrics["sloCompliance"].(map[string]any)
	if compliance["formatted"] != "Compliant" {
		t.Fatalf("expected SLO to be compliant (99.51%% >= 99%% target), got %v", compliance)
	}

	errorBudget := metrics["errorBudget"].(map[string]any)
	if errorBudget["available"] != true {
		t.Fatalf("expected error budget to be computable once SLO is configured, got %v", errorBudget)
	}
}

func TestNotificationDeploymentRecoveryDetectionIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "Recovery Admin", "recovery-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "recovery-admin@opspilot.dev")

	webhook := newCapturingWebhookServer(t)
	createNotificationChannel(t, app.router, adminToken, "Deploy Watch", constants.NotificationChannelWebhook, webhook.server.URL,
		[]string{constants.NotificationEventDeploymentFailed, constants.NotificationEventDeploymentRecovered}, nil)

	projectID := createProject(t, app.router, adminToken, "Recovery Project")
	clusterID := createCluster(t, app.router, adminToken, projectID, "recovery-cluster")

	appRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name": "Recovery App", "runtime": constants.ApplicationRuntimeGo, "port": 8080,
	})
	assertStatus(t, appRec, http.StatusCreated)
	applicationID, err := uuid.Parse(decodeDataMap(t, appRec)["id"].(string))
	if err != nil {
		t.Fatalf("parse application id: %v", err)
	}

	deploy := func() uuid.UUID {
		rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
			"applicationId":      applicationID.String(),
			"projectId":          projectID.String(),
			"targetClusterId":    clusterID.String(),
			"image":              "ghcr.io/opspilot/api",
			"imageTag":           "v1.0.0",
			"environment":        "production",
			"namespace":          "ops-api",
			"replicaCount":       1,
			"deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
		})
		assertStatus(t, rec, http.StatusCreated)
		id, err := uuid.Parse(decodeDataMap(t, rec)["id"].(string))
		if err != nil {
			t.Fatalf("parse deployment id: %v", err)
		}
		return id
	}

	orgID := *admin.OrganizationID
	ctx := t.Context()

	// First deployment attempt fails.
	firstID := deploy()
	if _, err := app.deploymentService.UpdateDeploymentStatus(ctx, firstID, orgID, admin.ID, constants.DeploymentStatusFailed); err != nil {
		t.Fatalf("mark first deployment failed: %v", err)
	}
	firstDeployment, err := app.deploymentRepo.GetByID(firstID, orgID)
	if err != nil {
		t.Fatalf("load first deployment: %v", err)
	}
	app.notificationService.NotifyDeploymentFailed(ctx, firstDeployment, admin.ID, "image pull error")
	if !webhook.received(constants.NotificationEventDeploymentFailed) {
		t.Fatalf("expected DEPLOYMENT_FAILED to be dispatched")
	}

	// Second deployment attempt succeeds - this is a recovery.
	secondID := deploy()
	if _, err := app.deploymentService.UpdateDeploymentStatus(ctx, secondID, orgID, admin.ID, constants.DeploymentStatusSucceeded); err != nil {
		t.Fatalf("mark second deployment succeeded: %v", err)
	}
	secondDeployment, err := app.deploymentRepo.GetByID(secondID, orgID)
	if err != nil {
		t.Fatalf("load second deployment: %v", err)
	}
	app.notificationService.NotifyDeploymentOutcomeSucceeded(ctx, secondDeployment, admin.ID)
	if !webhook.received(constants.NotificationEventDeploymentRecovered) {
		t.Fatalf("expected DEPLOYMENT_RECOVERED to be dispatched after a failed deployment is followed by a successful one")
	}

	// A third, unrelated success (previous was already Succeeded) must NOT
	// fire another recovery notification.
	countBeforeThird := webhook.count()
	thirdID := deploy()
	if _, err := app.deploymentService.UpdateDeploymentStatus(ctx, thirdID, orgID, admin.ID, constants.DeploymentStatusSucceeded); err != nil {
		t.Fatalf("mark third deployment succeeded: %v", err)
	}
	thirdDeployment, err := app.deploymentRepo.GetByID(thirdID, orgID)
	if err != nil {
		t.Fatalf("load third deployment: %v", err)
	}
	app.notificationService.NotifyDeploymentOutcomeSucceeded(ctx, thirdDeployment, admin.ID)
	if webhook.count() != countBeforeThird {
		t.Fatalf("expected no additional dispatch for a normal Succeeded-to-Succeeded transition")
	}
}

func TestApplicationSLOInsufficientDataForBrandNewApplication(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "SLO New Admin", "slo-new-admin@opspilot.dev", "password123")

	projectID := createProject(t, app.router, adminToken, "SLO New Project")
	appRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "Brand New App",
		"runtime": constants.ApplicationRuntimeGo,
		"port":    8080,
	})
	assertStatus(t, appRec, http.StatusCreated)
	applicationData := decodeDataMap(t, appRec)
	applicationIDStr, _ := applicationData["id"].(string)

	metricsRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationIDStr+"/slo/metrics", adminToken, nil)
	assertStatus(t, metricsRec, http.StatusOK)
	metrics := decodeDataMap(t, metricsRec)

	mttr := metrics["mttr"].(map[string]any)
	if mttr["available"] != false || mttr["formatted"] != "Insufficient data" {
		t.Fatalf("expected MTTR to report insufficient data for a brand new application with no incidents, got %v", mttr)
	}
	errorBudget := metrics["errorBudget"].(map[string]any)
	if errorBudget["available"] != false {
		t.Fatalf("expected error budget to report insufficient data when SLO is not configured, got %v", errorBudget)
	}
}
