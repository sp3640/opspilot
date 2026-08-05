package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func TestDeploymentHistoryIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "History Admin", "history-admin@opspilot.dev", "password123")
	memberToken := registerAndLogin(t, app.router, "History Member", "history-member@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "history-admin@opspilot.dev")
	member := mustGetUserByEmail(t, app.userRepo, "history-member@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	organizationID := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(member.ID, organizationID, models.RoleUser); err != nil {
		t.Fatalf("assign member to organization: %v", err)
	}

	projectID := createProject(t, app.router, adminToken, "History Project")
	clusterID := createCluster(t, app.router, adminToken, projectID, "history-cluster")

	applicationRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "History App",
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
		"image":              "ghcr.io/opspilot/history",
		"imageTag":           "v1.0.0",
		"environment":        "production",
		"namespace":          "history-app",
		"replicaCount":       2,
		"deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
	})
	assertStatus(t, createRec, http.StatusCreated)
	deploymentID, err := uuid.Parse(decodeDataMap(t, createRec)["id"].(string))
	if err != nil {
		t.Fatalf("parse deployment id: %v", err)
	}

	initialHistoryRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String()+"/history?sort=revision&order=asc", memberToken, nil)
	assertStatus(t, initialHistoryRec, http.StatusOK)
	initialHistory := decodeHistoryListItems(t, initialHistoryRec)
	if len(initialHistory) != 1 {
		t.Fatalf("expected one history revision after create, got %d", len(initialHistory))
	}
	if int(initialHistory[0]["revision"].(float64)) != 1 {
		t.Fatalf("expected first revision to be 1")
	}
	if initialHistory[0]["changeSummary"].(string) != "Deployment created" {
		t.Fatalf("expected creation summary")
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+deploymentID.String(), adminToken, map[string]any{
		"imageTag":     "v1.0.1",
		"replicaCount": 4,
	}), http.StatusOK)

	if _, err := app.deploymentService.UpdateDeploymentStatus(context.Background(), deploymentID, organizationID, admin.ID, constants.DeploymentStatusSucceeded); err != nil {
		t.Fatalf("set deployment succeeded: %v", err)
	}
	if _, err := app.deploymentService.UpdateDeploymentStatus(context.Background(), deploymentID, organizationID, admin.ID, constants.DeploymentStatusFailed); err != nil {
		t.Fatalf("set deployment failed: %v", err)
	}
	if _, err := app.deploymentService.UpdateDeploymentStatus(context.Background(), deploymentID, organizationID, admin.ID, constants.DeploymentStatusRolledBack); err != nil {
		t.Fatalf("set deployment rolled back: %v", err)
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+deploymentID.String()+"/cancel", adminToken, nil), http.StatusOK)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+deploymentID.String(), memberToken, map[string]any{
		"imageTag": "forbidden",
	}), http.StatusForbidden)

	historyRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String()+"/history?sort=revision&order=asc", memberToken, nil)
	assertStatus(t, historyRec, http.StatusOK)
	historyItems := decodeHistoryListItems(t, historyRec)
	if len(historyItems) != 6 {
		t.Fatalf("expected six revisions, got %d", len(historyItems))
	}

	for i := range historyItems {
		revision := int(historyItems[i]["revision"].(float64))
		if revision != i+1 {
			t.Fatalf("expected revision %d at index %d, got %d", i+1, i, revision)
		}
	}

	if historyItems[2]["status"].(string) != constants.DeploymentStatusSucceeded {
		t.Fatalf("expected succeeded status at revision 3")
	}
	if historyItems[3]["status"].(string) != constants.DeploymentStatusFailed {
		t.Fatalf("expected failed status at revision 4")
	}
	if historyItems[4]["status"].(string) != constants.DeploymentStatusRolledBack {
		t.Fatalf("expected rolled back status at revision 5")
	}
	if historyItems[5]["status"].(string) != constants.DeploymentStatusCancelled {
		t.Fatalf("expected cancelled status at revision 6")
	}

	revisionTwoRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String()+"/history/2", memberToken, nil)
	assertStatus(t, revisionTwoRec, http.StatusOK)
	revisionTwo := decodeDataMap(t, revisionTwoRec)
	if revisionTwo["imageTag"].(string) != "v1.0.1" {
		t.Fatalf("expected revision 2 image tag v1.0.1")
	}
	if revisionTwo["changeSummary"].(string) != "Deployment updated" {
		t.Fatalf("expected revision 2 summary to reflect update")
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String()+"/history/999", adminToken, nil), http.StatusNotFound)

	orgBRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/organizations", adminToken, map[string]any{
		"name":        "History Org B",
		"description": "secondary org",
	})
	assertStatus(t, orgBRec, http.StatusCreated)
	orgBData := decodeDataMap(t, orgBRec)
	orgBID, err := uuid.Parse(orgBData["id"].(string))
	if err != nil {
		t.Fatalf("parse org b id: %v", err)
	}
	if err := app.userRepo.AssignOrganizationAndRole(member.ID, orgBID, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign member to org b: %v", err)
	}
	memberOrgBToken := loginOnly(t, app.router, "history-member@opspilot.dev", "password123")

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String()+"/history", memberOrgBToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String()+"/history/1", memberOrgBToken, nil), http.StatusForbidden)
}

func decodeHistoryListItems(t *testing.T, rec *httptest.ResponseRecorder) []map[string]any {
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
