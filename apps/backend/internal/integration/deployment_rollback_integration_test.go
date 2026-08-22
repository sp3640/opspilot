package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func TestDeploymentRollbackIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "Rollback Admin", "rollback-admin@opspilot.dev", "password123")
	memberToken := registerAndLogin(t, app.router, "Rollback Member", "rollback-member@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "rollback-admin@opspilot.dev")
	member := mustGetUserByEmail(t, app.userRepo, "rollback-member@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	organizationID := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(member.ID, organizationID, models.RoleUser); err != nil {
		t.Fatalf("assign member to organization: %v", err)
	}

	projectID := createProject(t, app.router, adminToken, "Rollback Project")
	clusterID := createCluster(t, app.router, adminToken, projectID, "rollback-cluster")

	applicationRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "Rollback App",
		"runtime": constants.ApplicationRuntimeGo,
		"port":    8080,
	})
	assertStatus(t, applicationRec, http.StatusCreated)
	applicationID, err := uuid.Parse(decodeDataMap(t, applicationRec)["id"].(string))
	if err != nil {
		t.Fatalf("parse application id: %v", err)
	}

	createRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
		"applicationId":      applicationID.String(),
		"projectId":          projectID.String(),
		"targetClusterId":    clusterID.String(),
		"image":              "ghcr.io/opspilot/rollback",
		"imageTag":           "v1.0.0",
		"environment":        "production",
		"namespace":          "rollback-app",
		"replicaCount":       2,
		"deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
	})
	assertStatus(t, createRec, http.StatusCreated)
	deploymentID, err := uuid.Parse(decodeDataMap(t, createRec)["id"].(string))
	if err != nil {
		t.Fatalf("parse deployment id: %v", err)
	}

	updateRec := doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+deploymentID.String(), adminToken, map[string]any{
		"image":              "ghcr.io/opspilot/rollback-updated",
		"imageTag":           "v2.0.0",
		"replicaCount":       5,
		"namespace":          "rollback-updated",
		"environment":        "staging",
		"deploymentStrategy": constants.DeploymentStrategyRecreate,
	})
	assertStatus(t, updateRec, http.StatusOK)

	historyBeforeRollbackRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String()+"/history?sort=revision&order=asc", adminToken, nil)
	assertStatus(t, historyBeforeRollbackRec, http.StatusOK)
	historyBefore := decodeHistoryItemsForRollback(t, historyBeforeRollbackRec)
	if len(historyBefore) != 2 {
		t.Fatalf("expected two revisions before rollback, got %d", len(historyBefore))
	}
	baseRevisionOne := cloneMap(historyBefore[0])
	baseRevisionTwo := cloneMap(historyBefore[1])

	rollbackAsMemberRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+deploymentID.String()+"/rollback", memberToken, map[string]any{"revision": 1})
	assertStatus(t, rollbackAsMemberRec, http.StatusForbidden)

	rollbackLatestRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+deploymentID.String()+"/rollback", adminToken, map[string]any{"revision": 2})
	assertStatus(t, rollbackLatestRec, http.StatusBadRequest)

	rollbackInvalidRevisionRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+deploymentID.String()+"/rollback", adminToken, map[string]any{"revision": 999})
	assertStatus(t, rollbackInvalidRevisionRec, http.StatusNotFound)

	secondDeploymentRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
		"applicationId":      applicationID.String(),
		"projectId":          projectID.String(),
		"targetClusterId":    clusterID.String(),
		"image":              "ghcr.io/opspilot/rollback-second",
		"imageTag":           "v1.0.0",
		"environment":        "production",
		"namespace":          "rollback-second",
		"replicaCount":       1,
		"deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
	})
	assertStatus(t, secondDeploymentRec, http.StatusCreated)
	secondDeploymentID, err := uuid.Parse(decodeDataMap(t, secondDeploymentRec)["id"].(string))
	if err != nil {
		t.Fatalf("parse second deployment id: %v", err)
	}

	// Make second deployment reach revision 3 so using that revision against first deployment proves wrong-deployment rejection.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+secondDeploymentID.String(), adminToken, map[string]any{"imageTag": "v1.0.1"}), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+secondDeploymentID.String(), adminToken, map[string]any{"imageTag": "v1.0.2"}), http.StatusOK)

	wrongDeploymentRevisionRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+deploymentID.String()+"/rollback", adminToken, map[string]any{"revision": 3})
	assertStatus(t, wrongDeploymentRevisionRec, http.StatusNotFound)

	orgBRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/organizations", adminToken, map[string]any{
		"name":        "Rollback Org B",
		"description": "secondary org",
	})
	assertStatus(t, orgBRec, http.StatusCreated)
	orgBID, err := uuid.Parse(decodeDataMap(t, orgBRec)["id"].(string))
	if err != nil {
		t.Fatalf("parse org b id: %v", err)
	}
	if err := app.userRepo.AssignOrganizationAndRole(member.ID, orgBID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign member to org b: %v", err)
	}
	memberOrgBToken := loginOnly(t, app.router, "rollback-member@opspilot.dev", "password123")
	wrongOrgRollbackRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+deploymentID.String()+"/rollback", memberOrgBToken, map[string]any{"revision": 1})
	assertStatus(t, wrongOrgRollbackRec, http.StatusForbidden)

	rollbackRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+deploymentID.String()+"/rollback", adminToken, map[string]any{"revision": 1})
	assertStatus(t, rollbackRec, http.StatusOK)
	rollbackData := decodeDataMap(t, rollbackRec)
	currentRevision := int(rollbackData["currentRevision"].(float64))
	if currentRevision != 3 {
		t.Fatalf("expected current revision 3 after rollback, got %d", currentRevision)
	}
	if int(rollbackData["rollbackSourceRevision"].(float64)) != 1 {
		t.Fatalf("expected rollback source revision 1")
	}

	deploymentData, ok := rollbackData["deployment"].(map[string]any)
	if !ok {
		t.Fatalf("expected deployment object in rollback response")
	}
	if deploymentData["image"].(string) != baseRevisionOne["image"].(string) {
		t.Fatalf("expected deployment image restored from revision 1")
	}
	if deploymentData["imageTag"].(string) != baseRevisionOne["imageTag"].(string) {
		t.Fatalf("expected deployment imageTag restored from revision 1")
	}
	if deploymentData["namespace"].(string) != baseRevisionOne["namespace"].(string) {
		t.Fatalf("expected deployment namespace restored from revision 1")
	}
	if deploymentData["environment"].(string) != baseRevisionOne["environment"].(string) {
		t.Fatalf("expected deployment environment restored from revision 1")
	}
	if deploymentData["deploymentStrategy"].(string) != baseRevisionOne["deploymentStrategy"].(string) {
		t.Fatalf("expected deployment strategy restored from revision 1")
	}
	if int(deploymentData["replicaCount"].(float64)) != int(baseRevisionOne["replicaCount"].(float64)) {
		t.Fatalf("expected deployment replica count restored from revision 1")
	}
	if deploymentData["status"].(string) != constants.DeploymentStatusPending {
		t.Fatalf("expected deployment status pending after rollback")
	}

	historyAfterRollbackRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String()+"/history?sort=revision&order=asc", adminToken, nil)
	assertStatus(t, historyAfterRollbackRec, http.StatusOK)
	historyAfter := decodeHistoryItemsForRollback(t, historyAfterRollbackRec)
	if len(historyAfter) != 3 {
		t.Fatalf("expected three revisions after rollback, got %d", len(historyAfter))
	}

	if !equalHistorySnapshot(baseRevisionOne, historyAfter[0]) {
		t.Fatalf("revision 1 mutated after rollback")
	}
	if !equalHistorySnapshot(baseRevisionTwo, historyAfter[1]) {
		t.Fatalf("revision 2 mutated after rollback")
	}
	if int(historyAfter[2]["revision"].(float64)) != 3 {
		t.Fatalf("expected new revision number 3")
	}
	if historyAfter[2]["status"].(string) != constants.DeploymentStatusPending {
		t.Fatalf("expected rollback snapshot status pending")
	}
	if historyAfter[2]["changeSummary"].(string) != "Deployment rolled back to revision 1" {
		t.Fatalf("expected rollback change summary")
	}
	if historyAfter[2]["image"].(string) != baseRevisionOne["image"].(string) {
		t.Fatalf("expected rollback revision config copied from source revision")
	}

	// Rollback must leave a real audit trail (Phase 20: every remediation
	// action creates an audit event), independent of whether an executor is
	// wired to actually re-apply the change to a cluster.
	auditRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String()+"/audit-logs", adminToken, nil)
	assertStatus(t, auditRec, http.StatusOK)
	auditItems := decodeHistoryItemsForRollback(t, auditRec)
	foundRollbackAudit := false
	for _, item := range auditItems {
		if item["entity_type"] == "deployment" && item["field_name"] == "revision" {
			foundRollbackAudit = true
			break
		}
	}
	if !foundRollbackAudit {
		t.Fatalf("expected a deployment audit log entry recording the rollback, got %v", auditItems)
	}

	deleteRec := doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/deployments/"+deploymentID.String(), adminToken, nil)
	assertStatus(t, deleteRec, http.StatusOK)
	deletedRollbackRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+deploymentID.String()+"/rollback", adminToken, map[string]any{"revision": 1})
	assertStatus(t, deletedRollbackRec, http.StatusNotFound)
}

func decodeHistoryItemsForRollback(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	env := decodeEnvelope(t, rec)
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode history list payload: %v; data=%s", err, string(env.Data))
	}

	return payload.Items
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func equalHistorySnapshot(expected map[string]any, actual map[string]any) bool {
	keys := []string{"revision", "image", "imageTag", "replicaCount", "namespace", "environment", "deploymentStrategy", "status", "changeSummary"}
	for _, key := range keys {
		if expected[key] != actual[key] {
			return false
		}
	}
	return true
}
