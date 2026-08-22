package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func TestDeploymentManagementIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "Deploy Admin", "deploy-admin@opspilot.dev", "password123")
	registerAndLogin(t, app.router, "Deploy Member", "deploy-member@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "deploy-admin@opspilot.dev")
	member := mustGetUserByEmail(t, app.userRepo, "deploy-member@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	organizationA := *admin.OrganizationID
	// Uninvited registration now creates its own organization, so the member
	// must be explicitly moved into organization A and re-authenticated to
	// pick up the updated organization/role claims.
	if err := app.userRepo.AssignOrganizationAndRole(member.ID, organizationA, models.RoleViewer); err != nil {
		t.Fatalf("assign member to organization: %v", err)
	}
	memberToken := loginOnly(t, app.router, "deploy-member@opspilot.dev", "password123")

	projectID := createProject(t, app.router, adminToken, "Deployment Project")
	clusterID := createCluster(t, app.router, adminToken, projectID, "deployment-cluster")

	applicationRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "Deployment App",
		"runtime": constants.ApplicationRuntimeGo,
		"port":    8080,
	})
	assertStatus(t, applicationRec, http.StatusCreated)
	applicationData := decodeDataMap(t, applicationRec)
	applicationID, err := uuid.Parse(applicationData["id"].(string))
	if err != nil {
		t.Fatalf("parse application id: %v", err)
	}

	createRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
		"applicationId":      applicationID.String(),
		"projectId":          projectID.String(),
		"targetClusterId":    clusterID.String(),
		"image":              "ghcr.io/opspilot/api",
		"imageTag":           "v1.0.0",
		"environment":        "production",
		"namespace":          "ops-api",
		"replicaCount":       2,
		"deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
	})
	assertStatus(t, createRec, http.StatusCreated)
	created := decodeDataMap(t, createRec)
	deploymentID, err := uuid.Parse(created["id"].(string))
	if err != nil {
		t.Fatalf("parse deployment id: %v", err)
	}
	if created["status"].(string) != constants.DeploymentStatusPending {
		t.Fatalf("expected deployment status pending, got %v", created["status"])
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments", memberToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/deployments", memberToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects/"+projectID.String()+"/deployments", memberToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String(), memberToken, nil), http.StatusOK)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", memberToken, map[string]any{
		"applicationId":      applicationID.String(),
		"projectId":          projectID.String(),
		"targetClusterId":    clusterID.String(),
		"image":              "ghcr.io/opspilot/api",
		"environment":        "production",
		"namespace":          "ops-api",
		"replicaCount":       1,
		"deploymentStrategy": constants.DeploymentStrategyRecreate,
	}), http.StatusForbidden)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+deploymentID.String(), memberToken, map[string]any{
		"imageTag": "v1.0.1",
	}), http.StatusForbidden)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+deploymentID.String()+"/cancel", memberToken, nil), http.StatusForbidden)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/deployments/"+deploymentID.String(), memberToken, nil), http.StatusForbidden)

	updateRec := doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+deploymentID.String(), adminToken, map[string]any{
		"imageTag":           "v1.0.2",
		"replicaCount":       3,
		"deploymentStrategy": constants.DeploymentStrategyRecreate,
	})
	assertStatus(t, updateRec, http.StatusOK)
	updated := decodeDataMap(t, updateRec)
	if updated["imageTag"].(string) != "v1.0.2" {
		t.Fatalf("expected updated image tag")
	}
	if int(updated["replicaCount"].(float64)) != 3 {
		t.Fatalf("expected updated replica count")
	}

	cancelRec := doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+deploymentID.String()+"/cancel", adminToken, nil)
	assertStatus(t, cancelRec, http.StatusOK)
	cancelled := decodeDataMap(t, cancelRec)
	if cancelled["status"].(string) != constants.DeploymentStatusCancelled {
		t.Fatalf("expected cancelled status, got %v", cancelled["status"])
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
		"applicationId":      uuid.New().String(),
		"projectId":          projectID.String(),
		"targetClusterId":    clusterID.String(),
		"image":              "ghcr.io/opspilot/api",
		"environment":        "production",
		"namespace":          "ops-api",
		"replicaCount":       1,
		"deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
	}), http.StatusBadRequest)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
		"applicationId":      applicationID.String(),
		"projectId":          projectID.String(),
		"targetClusterId":    uuid.New().String(),
		"image":              "ghcr.io/opspilot/api",
		"environment":        "production",
		"namespace":          "ops-api",
		"replicaCount":       1,
		"deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
	}), http.StatusBadRequest)

	secondRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
		"applicationId":      applicationID.String(),
		"projectId":          projectID.String(),
		"targetClusterId":    clusterID.String(),
		"image":              "ghcr.io/opspilot/api",
		"imageTag":           "v2.0.0",
		"environment":        "staging",
		"namespace":          "ops-api",
		"replicaCount":       1,
		"deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
	})
	assertStatus(t, secondRec, http.StatusCreated)
	second := decodeDataMap(t, secondRec)
	secondID, err := uuid.Parse(second["id"].(string))
	if err != nil {
		t.Fatalf("parse second deployment id: %v", err)
	}

	latest, err := app.deploymentRepo.GetLatestDeployment(applicationID, organizationA)
	if err != nil {
		t.Fatalf("get latest deployment: %v", err)
	}
	if latest.ID != secondID {
		t.Fatalf("expected latest deployment %s, got %s", secondID.String(), latest.ID.String())
	}

	// GET /applications/:id/deployments/latest (Phase 8) surfaces the same
	// "most recently created" deployment over HTTP — this is the
	// authoritative Application -> Cluster/Namespace/Environment mapping the
	// frontend Runtime summary consumes directly.
	latestRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/deployments/latest", memberToken, nil)
	assertStatus(t, latestRec, http.StatusOK)
	latestData := decodeDataMap(t, latestRec)
	if latestData["id"].(string) != secondID.String() {
		t.Fatalf("expected latest deployment endpoint to return %s, got %v", secondID.String(), latestData["id"])
	}
	if latestData["targetClusterId"].(string) != clusterID.String() {
		t.Fatalf("expected latest deployment to map to cluster %s, got %v", clusterID.String(), latestData["targetClusterId"])
	}
	if latestData["namespace"].(string) != "ops-api" {
		t.Fatalf("expected latest deployment namespace ops-api, got %v", latestData["namespace"])
	}
	if latestData["environment"].(string) != constants.DeploymentEnvironmentStaging {
		t.Fatalf("expected latest deployment environment %s, got %v", constants.DeploymentEnvironmentStaging, latestData["environment"])
	}

	// An application with no deployments yet has no runtime mapping: the
	// endpoint returns a clean 404 rather than a zero-value/empty record, so
	// the frontend can render a clear empty state.
	noDeploymentsAppRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name":    "No Deployments App",
		"runtime": constants.ApplicationRuntimeGo,
		"port":    8080,
	})
	assertStatus(t, noDeploymentsAppRec, http.StatusCreated)
	noDeploymentsAppID := decodeDataMap(t, noDeploymentsAppRec)["id"].(string)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+noDeploymentsAppID+"/deployments/latest", adminToken, nil), http.StatusNotFound)

	orgBRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/organizations", adminToken, map[string]any{
		"name":        "Deploy Org B",
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
	memberOrgBToken := loginOnly(t, app.router, "deploy-member@opspilot.dev", "password123")

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String(), memberOrgBToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+deploymentID.String(), memberOrgBToken, map[string]any{"imageTag": "forbidden"}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/deployments/"+deploymentID.String(), memberOrgBToken, nil), http.StatusForbidden)
	// Cross-org access to another org's application's runtime mapping never
	// succeeds — arbitrary cross-org resource mapping must be impossible.
	// This matches the existing application-scoped deployment endpoints'
	// convention of reporting a cross-org application as not found rather
	// than distinguishing "forbidden" from "doesn't exist."
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/deployments/latest", memberOrgBToken, nil), http.StatusNotFound)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/deployments/"+deploymentID.String(), adminToken, nil), http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+deploymentID.String(), adminToken, nil), http.StatusNotFound)
}
