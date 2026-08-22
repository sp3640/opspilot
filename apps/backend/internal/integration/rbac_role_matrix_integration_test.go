package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

// TestRoleMatrixAuthorizationIntegration exercises the four-role RBAC matrix
// end to end over HTTP: Platform Admin, DevOps Engineer, Developer, and
// Viewer, covering the capability checklist for each role plus cross
// organization isolation. Negative deployment/cluster/application checks use
// a syntactically valid but nonexistent UUID: RequirePermission runs before
// any lookup, so a role that lacks the permission gets 403 regardless of
// whether the resource exists, while a role that holds the permission gets
// a 404/400 from the service layer instead — proving authorization passed.
func TestRoleMatrixAuthorizationIntegration(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)

	adminToken := registerAndLogin(t, app.router, "Matrix Admin", "matrix-admin@opspilot.dev", "password123")
	admin := mustGetUserByEmail(t, app.userRepo, "matrix-admin@opspilot.dev")
	if admin.OrganizationID == nil {
		t.Fatalf("expected admin organization")
	}
	orgID := *admin.OrganizationID

	registerAndLogin(t, app.router, "Matrix DevOps", "matrix-devops@opspilot.dev", "password123")
	devops := mustGetUserByEmail(t, app.userRepo, "matrix-devops@opspilot.dev")
	mustSetOrgAndRole(t, app.userRepo, devops.ID, orgID, models.RoleDevOpsEngineer)
	// The JWT minted at registration still carries the pre-reassignment role
	// claim, so a fresh login is required to pick up the new role.
	devopsToken := loginOnly(t, app.router, "matrix-devops@opspilot.dev", "password123")

	registerAndLogin(t, app.router, "Matrix Developer", "matrix-developer@opspilot.dev", "password123")
	developer := mustGetUserByEmail(t, app.userRepo, "matrix-developer@opspilot.dev")
	mustSetOrgAndRole(t, app.userRepo, developer.ID, orgID, models.RoleDeveloper)
	developerToken := loginOnly(t, app.router, "matrix-developer@opspilot.dev", "password123")

	registerAndLogin(t, app.router, "Matrix Viewer", "matrix-viewer@opspilot.dev", "password123")
	viewer := mustGetUserByEmail(t, app.userRepo, "matrix-viewer@opspilot.dev")
	mustSetOrgAndRole(t, app.userRepo, viewer.ID, orgID, models.RoleViewer)
	viewerToken := loginOnly(t, app.router, "matrix-viewer@opspilot.dev", "password123")

	projectID := createProject(t, app.router, adminToken, "Matrix Project")
	clusterID := createCluster(t, app.router, adminToken, projectID, "matrix-cluster")
	resourceID := createResource(t, app.router, adminToken, projectID, "matrix-resource")
	incidentID := createIncident(t, app.router, adminToken, projectID, "Matrix Incident")
	alertID := createAlert(t, app.router, adminToken, projectID, "Matrix Alert", resourceID)
	applicationID := createApplicationForServices(t, app.router, adminToken, projectID, "Matrix App")

	bogusID := uuid.NewString()

	t.Run("Platform Admin can perform all admin actions", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/teams", adminToken, map[string]any{
			"name": "Matrix Admin Team", "description": "team",
		}), http.StatusCreated)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "matrix-invitee@opspilot.dev", "role": models.RoleViewer,
		}), http.StatusCreated)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/organizations/"+orgID.String(), adminToken, map[string]any{
			"name": "Matrix Org Renamed", "description": "renamed",
		}), http.StatusOK)
	})

	t.Run("DevOps Engineer can manage clusters, applications, deployments, incidents, and alerts", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters", devopsToken, map[string]any{
			"project_id": projectID.String(), "name": "devops-cluster", "provider": constants.ClusterProviderKubernetes,
			"status": constants.ClusterStatusConnected, "connection_type": constants.ClusterConnectionTypeKubeconfig,
			"kubeconfig_encrypted": "dummy", "api_endpoint": "https://api.example", "region": "us-east-1",
			"version": "1.29", "validation_error": "", "metadata": map[string]any{"env": "test"},
		}), http.StatusCreated)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters/"+clusterID.String()+"/default", devopsToken, nil), http.StatusOK)

		assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/applications/"+applicationID.String(), devopsToken, map[string]any{
			"name": "Matrix App Updated", "runtime": constants.ApplicationRuntimeGo, "port": 8080,
		}), http.StatusOK)

		// Not-forbidden proves the permission gate passed; the bogus id then
		// fails downstream in the service layer instead of at authorization.
		if status := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+bogusID+"/rollback", devopsToken, map[string]any{"revision": 1}).Code; status == http.StatusForbidden {
			t.Fatalf("expected DevOps Engineer to pass deployment:rollback authorization, got 403")
		}
		if status := doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+bogusID+"/cancel", devopsToken, nil).Code; status == http.StatusForbidden {
			t.Fatalf("expected DevOps Engineer to pass deployment:cancel authorization, got 403")
		}

		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents/"+itoa(incidentID)+"/comments", devopsToken, map[string]any{
			"content": "DevOps comment",
		}), http.StatusCreated)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/acknowledge", devopsToken, nil), http.StatusOK)
	})

	t.Run("DevOps Engineer cannot manage organization, invitations, teams, or projects", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/organizations/"+orgID.String(), devopsToken, map[string]any{
			"name": "nope", "description": "no",
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", devopsToken, map[string]any{
			"email": "devops-blocked@opspilot.dev", "role": models.RoleViewer,
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/teams", devopsToken, map[string]any{
			"name": "devops-team", "description": "team",
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects", devopsToken, map[string]any{
			"name": "devops-project", "description": "project",
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/projects/"+projectID.String(), devopsToken, map[string]any{
			"name": "devops-project-update", "description": "project",
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/teams", devopsToken, map[string]any{
			"teamId": uuid.NewString(),
		}), http.StatusForbidden)
	})

	t.Run("Developer can manage incidents and alerts and read applications and deployments", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/incidents/"+itoa(incidentID), developerToken, map[string]any{
			"title": "Matrix Incident Updated", "description": "updated", "severity": constants.SeverityP1,
			"status": constants.StatusOpen, "project_id": projectID.String(),
		}), http.StatusOK)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/resolve", developerToken, nil), http.StatusOK)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/reopen", developerToken, nil), http.StatusOK)

		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String(), developerToken, nil), http.StatusOK)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String()+"/deployments", developerToken, nil), http.StatusOK)
	})

	t.Run("Developer cannot manage clusters, mutate deployments, manage applications, projects, teams, or invite", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters", developerToken, map[string]any{
			"project_id": projectID.String(), "name": "developer-cluster", "provider": constants.ClusterProviderKubernetes,
			"status": constants.ClusterStatusConnected, "connection_type": constants.ClusterConnectionTypeKubeconfig,
			"kubeconfig_encrypted": "dummy", "api_endpoint": "https://api.example", "region": "us-east-1",
			"version": "1.29", "validation_error": "", "metadata": map[string]any{"env": "test"},
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/applications/"+applicationID.String(), developerToken, map[string]any{
			"name": "developer-app-update", "runtime": constants.ApplicationRuntimeGo, "port": 8080,
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+bogusID+"/rollback", developerToken, map[string]any{"revision": 1}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/deployments/"+bogusID+"/cancel", developerToken, nil), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", developerToken, map[string]any{
			"application_id": applicationID.String(), "target_cluster_id": clusterID.String(), "image": "example/app:latest",
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects", developerToken, map[string]any{
			"name": "developer-project", "description": "project",
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/teams", developerToken, map[string]any{
			"name": "developer-team", "description": "team",
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", developerToken, map[string]any{
			"email": "developer-blocked@opspilot.dev", "role": models.RoleViewer,
		}), http.StatusForbidden)
	})

	t.Run("Viewer can read but cannot mutate anything", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/projects", viewerToken, nil), http.StatusOK)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String(), viewerToken, nil), http.StatusOK)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID.String(), viewerToken, nil), http.StatusOK)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/incidents", viewerToken, nil), http.StatusOK)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/alerts", viewerToken, nil), http.StatusOK)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/incidents/"+itoa(incidentID)+"/comments", viewerToken, nil), http.StatusOK)

		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects", viewerToken, map[string]any{
			"name": "viewer-project", "description": "project",
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/clusters", viewerToken, map[string]any{
			"project_id": projectID.String(), "name": "viewer-cluster", "provider": constants.ClusterProviderKubernetes,
			"status": constants.ClusterStatusConnected, "connection_type": constants.ClusterConnectionTypeKubeconfig,
			"kubeconfig_encrypted": "dummy", "api_endpoint": "https://api.example", "region": "us-east-1",
			"version": "1.29", "validation_error": "", "metadata": map[string]any{"env": "test"},
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments/"+bogusID+"/rollback", viewerToken, map[string]any{"revision": 1}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents", viewerToken, map[string]any{
			"title": "viewer-incident", "description": "d", "severity": constants.SeverityP1,
			"status": constants.StatusOpen, "project_id": projectID.String(),
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/incidents/"+itoa(incidentID)+"/comments", viewerToken, map[string]any{
			"content": "viewer comment",
		}), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/alerts/"+itoa(alertID)+"/acknowledge", viewerToken, nil), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", viewerToken, map[string]any{
			"email": "viewer-blocked@opspilot.dev", "role": models.RoleViewer,
		}), http.StatusForbidden)
	})

	t.Run("Cross-organization isolation holds for the new roles", func(t *testing.T) {
		registerAndLogin(t, app.router, "Matrix Admin B", "matrix-admin-b@opspilot.dev", "password123")
		otherAdmin := mustGetUserByEmail(t, app.userRepo, "matrix-admin-b@opspilot.dev")

		// registerAndLogin lands non-first, non-invited users in the same
		// default organization as everyone else in this test's DB, so a
		// genuinely separate "org B" must be created explicitly.
		otherOrgID := mustCreateTeamTestOrganization(t, app.organizationRepo, "Matrix Org B", "matrix-org-b", otherAdmin.ID)
		if err := app.userRepo.AssignOrganizationAndRole(otherAdmin.ID, otherOrgID, models.RolePlatformAdmin); err != nil {
			t.Fatalf("assign other admin role: %v", err)
		}

		registerAndLogin(t, app.router, "Matrix DevOps B", "matrix-devops-b@opspilot.dev", "password123")
		otherDevOps := mustGetUserByEmail(t, app.userRepo, "matrix-devops-b@opspilot.dev")
		mustSetOrgAndRole(t, app.userRepo, otherDevOps.ID, otherOrgID, models.RoleDevOpsEngineer)
		otherDevOpsToken := loginOnly(t, app.router, "matrix-devops-b@opspilot.dev", "password123")

		// A DevOps Engineer in org B, despite holding cluster:manage, cannot
		// reach a cluster that belongs to org A.
		assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/clusters/"+clusterID.String(), otherDevOpsToken, nil), http.StatusForbidden)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/clusters/"+clusterID.String(), otherDevOpsToken, map[string]any{
			"project_id": projectID.String(), "name": "cross-org-update", "provider": constants.ClusterProviderKubernetes,
			"status": constants.ClusterStatusConnected, "connection_type": constants.ClusterConnectionTypeKubeconfig,
			"kubeconfig_encrypted": "dummy", "api_endpoint": "https://api.example", "region": "us-east-1",
			"version": "1.29", "validation_error": "", "metadata": map[string]any{"env": "test"},
		}), http.StatusForbidden)
	})

	t.Run("Invitation roles are validated and acceptance assigns the invited role", func(t *testing.T) {
		roleSlugs := map[string]string{
			models.RolePlatformAdmin:  "platform-admin",
			models.RoleDevOpsEngineer: "devops-engineer",
			models.RoleDeveloper:      "developer",
			models.RoleViewer:         "viewer",
		}

		for _, role := range []string{models.RolePlatformAdmin, models.RoleDevOpsEngineer, models.RoleDeveloper, models.RoleViewer} {
			email := "invite-" + roleSlugs[role] + "@opspilot.dev"
			inviteRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
				"email": email, "role": role,
			})
			assertStatus(t, inviteRec, http.StatusCreated)
			token, _ := decodeDataMap(t, inviteRec)["token"].(string)

			memberToken := registerAndLogin(t, app.router, "Invitee "+roleSlugs[role], email, "password123")
			assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations/accept", memberToken, map[string]any{
				"token": token,
			}), http.StatusOK)

			accepted := mustGetUserByEmail(t, app.userRepo, email)
			if accepted.Role != role {
				t.Fatalf("expected accepted invitation to assign role %q, got %q", role, accepted.Role)
			}
			if accepted.OrganizationID == nil || *accepted.OrganizationID != orgID {
				t.Fatalf("expected accepted invitation to assign organization %s", orgID)
			}
		}
	})

	t.Run("Invalid invitation role is rejected", func(t *testing.T) {
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "invalid-role@opspilot.dev", "role": "User",
		}), http.StatusBadRequest)
		assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/invitations", adminToken, map[string]any{
			"email": "invalid-role-2@opspilot.dev", "role": "Superuser",
		}), http.StatusBadRequest)
	})
}

func mustSetOrgAndRole(t *testing.T, userRepo *repository.UserRepository, userID uint, organizationID uuid.UUID, role string) {
	t.Helper()

	if err := userRepo.AssignOrganizationAndRole(userID, organizationID, role); err != nil {
		t.Fatalf("assign organization and role %q: %v", role, err)
	}
}
