package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	githubconnector "github.com/sp3640/opspilot/backend/internal/connector/github"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

const githubSuperSecretTestToken = "SUPER_SECRET_GITHUB_TOKEN"

// fakeGitHubServer stands in for GitHub's REST API - no test in this file
// depends on live GitHub. It serves the handful of endpoints
// internal/connector/github.Client calls, keyed by path, and requires the
// exact expected Bearer token.
type fakeGitHubServer struct {
	server *httptest.Server
}

func newFakeGitHubServer(t *testing.T, expectedToken string) *fakeGitHubServer {
	t.Helper()
	f := &fakeGitHubServer{}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+expectedToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch {
		case r.URL.Path == "/user":
			_ = json.NewEncoder(w).Encode(githubconnector.User{ID: 583231, Login: "octocat", Name: "The Octocat", AvatarURL: "https://avatars.githubusercontent.com/u/583231", HTMLURL: "https://github.com/octocat"})
		case r.URL.Path == "/user/repos":
			if r.URL.Query().Get("page") == "1" {
				repo := githubconnector.Repository{ID: 42, Name: "widgets", FullName: "acme/widgets", HTMLURL: "https://github.com/acme/widgets", DefaultBranch: "main"}
				repo.Owner.Login = "acme"
				_ = json.NewEncoder(w).Encode([]githubconnector.Repository{repo})
			} else {
				_ = json.NewEncoder(w).Encode([]githubconnector.Repository{})
			}
		case strings.HasSuffix(r.URL.Path, "/commits/deadbeef"):
			commit := githubconnector.Commit{SHA: "deadbeef", HTMLURL: "https://github.com/acme/widgets/commit/deadbeef"}
			commit.Commit.Message = "Fix payment timeout"
			commit.Commit.Author.Name = "Siddharth"
			commit.Commit.Author.Email = "sid@example.com"
			_ = json.NewEncoder(w).Encode(commit)
		case strings.HasSuffix(r.URL.Path, "/commits/deadbeef/pulls"):
			pr := githubconnector.PullRequest{Number: 421, Title: "Fix payment timeout", State: "closed"}
			pr.User.Login = "octocat"
			_ = json.NewEncoder(w).Encode([]githubconnector.PullRequest{pr})
		case strings.HasSuffix(r.URL.Path, "/commits/missing"):
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(r.URL.Path, "/commits"):
			commit := githubconnector.Commit{SHA: "abc123", HTMLURL: "https://github.com/acme/widgets/commit/abc123"}
			commit.Commit.Message = "Initial commit"
			commit.Commit.Author.Name = "Siddharth"
			_ = json.NewEncoder(w).Encode([]githubconnector.Commit{commit})
		case strings.HasSuffix(r.URL.Path, "/pulls"):
			pr := githubconnector.PullRequest{Number: 1, Title: "Add widgets", State: "open"}
			pr.User.Login = "octocat"
			_ = json.NewEncoder(w).Encode([]githubconnector.PullRequest{pr})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(f.server.Close)
	return f
}

func createGitHubIntegration(t *testing.T, app *rbacTestApp, token, accessToken string) string {
	t.Helper()
	rec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations", token, map[string]any{
		"type": constants.IntegrationTypeGitHub,
		"name": "Primary GitHub",
		"credentials": map[string]any{
			"access_token": accessToken,
		},
	})
	assertStatus(t, rec, http.StatusCreated)
	id, _ := decodeDataMap(t, rec)["id"].(string)
	return id
}

func TestGitHubDiscoverRepositoriesAndTestConnection(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "GitHub Admin", "github-admin@opspilot.dev", "password123")

	fake := newFakeGitHubServer(t, githubSuperSecretTestToken)
	app.githubClient.BaseURL = fake.server.URL

	integrationID := createGitHubIntegration(t, app, adminToken, githubSuperSecretTestToken)

	// The generic Sprint 27 /test endpoint now works for real against the
	// registered GitHub connector - no Sprint 27 file needed to change.
	testRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/test", adminToken, nil)
	assertStatus(t, testRec, http.StatusOK)
	testResult := decodeDataMap(t, testRec)
	if testResult["success"] != true {
		t.Fatalf("expected test connection to succeed against the fake GitHub server, got %v", testResult)
	}

	syncRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/github/repositories/sync", adminToken, nil)
	assertStatus(t, syncRec, http.StatusOK)
	syncData := decodeDataMap(t, syncRec)
	items, _ := syncData["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected exactly 1 discovered repository, got %v", syncData)
	}
	repo := items[0].(map[string]any)
	if repo["fullName"] != "acme/widgets" {
		t.Fatalf("expected repository acme/widgets, got %v", repo)
	}

	// A pure DB read (no GitHub call) must return the same repository.
	listRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/repositories", adminToken, nil)
	assertStatus(t, listRec, http.StatusOK)
	listData := decodeDataMap(t, listRec)
	listItems, _ := listData["items"].([]any)
	if len(listItems) != 1 {
		t.Fatalf("expected the stored repository list to contain 1 item, got %v", listData)
	}
	if listData["page"] != float64(1) || listData["limit"] != float64(20) || listData["total"] != float64(1) || listData["totalPages"] != float64(1) {
		t.Fatalf("expected stored repository pagination metadata, got %v", listData)
	}

	// Re-syncing must not duplicate the repository row.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/github/repositories/sync", adminToken, nil), http.StatusOK)
	againRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/repositories", adminToken, nil)
	againItems, _ := decodeDataMap(t, againRec)["items"].([]any)
	if len(againItems) != 1 {
		t.Fatalf("expected repository sync to be idempotent (no duplicates), got %d items", len(againItems))
	}
}

func TestGitHubConnectionTestFailsWithInvalidToken(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "GitHub Bad Token Admin", "github-badtoken-admin@opspilot.dev", "password123")

	fake := newFakeGitHubServer(t, githubSuperSecretTestToken)
	app.githubClient.BaseURL = fake.server.URL

	integrationID := createGitHubIntegration(t, app, adminToken, "not-the-right-token")

	testRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/test", adminToken, nil)
	assertStatus(t, testRec, http.StatusOK)
	result := decodeDataMap(t, testRec)
	if result["success"] != false {
		t.Fatalf("expected test connection to fail for an invalid token, got %v", result)
	}

	getRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID, adminToken, nil)
	data := decodeDataMap(t, getRec)
	if data["status"] != constants.IntegrationStatusError {
		t.Fatalf("expected integration status ERROR after a failed test, got %v", data["status"])
	}
}

func TestGitHubUnavailableReturnsSafeFailure(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "GitHub Unavailable Admin", "github-unavailable-admin@opspilot.dev", "password123")

	unreachable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	unreachable.Close() // closed immediately - connections to it now fail outright
	app.githubClient.BaseURL = unreachable.URL

	integrationID := createGitHubIntegration(t, app, adminToken, githubSuperSecretTestToken)

	checkRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/check", adminToken, nil)
	assertStatus(t, checkRec, http.StatusOK)
	result := decodeDataMap(t, checkRec)
	if result["success"] != false {
		t.Fatalf("expected the health check to fail cleanly against an unreachable GitHub, got %v", result)
	}
}

func TestGitHubRepositoryMappingAndCommitsAndPullRequests(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "GitHub Mapping Admin", "github-mapping-admin@opspilot.dev", "password123")

	fake := newFakeGitHubServer(t, githubSuperSecretTestToken)
	app.githubClient.BaseURL = fake.server.URL

	integrationID := createGitHubIntegration(t, app, adminToken, githubSuperSecretTestToken)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/github/repositories/sync", adminToken, nil), http.StatusOK)

	listRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/repositories", adminToken, nil)
	items, _ := decodeDataMap(t, listRec)["items"].([]any)
	repositoryID := items[0].(map[string]any)["id"].(string)

	projectID := createProject(t, app.router, adminToken, "GitHub Mapping Project")
	appRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name": "GitHub Mapped App", "runtime": constants.ApplicationRuntimeGo, "port": 8080,
	})
	assertStatus(t, appRec, http.StatusCreated)
	applicationID, _ := decodeDataMap(t, appRec)["id"].(string)

	// Not mapped yet.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID+"/github/repository", adminToken, nil), http.StatusNotFound)

	mapRec := doJSONRequest(t, app.router, http.MethodPut, "/api/v1/applications/"+applicationID+"/github/repository", adminToken, map[string]any{
		"repository_id": repositoryID,
	})
	assertStatus(t, mapRec, http.StatusOK)
	mapped := decodeDataMap(t, mapRec)
	repoField := mapped["repository"].(map[string]any)
	if repoField["fullName"] != "acme/widgets" {
		t.Fatalf("expected mapping to reference acme/widgets, got %v", mapped)
	}

	getMappingRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID+"/github/repository", adminToken, nil)
	assertStatus(t, getMappingRec, http.StatusOK)

	commitsRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/github/repositories/"+repositoryID+"/commits?page=2&limit=1", adminToken, nil)
	assertStatus(t, commitsRec, http.StatusOK)
	commitItems, _ := decodeDataMap(t, commitsRec)["items"].([]any)
	if len(commitItems) != 1 || commitItems[0].(map[string]any)["sha"] != "abc123" {
		t.Fatalf("expected 1 commit abc123, got %v", commitItems)
	}
	if commits := decodeDataMap(t, commitsRec); commits["page"] != float64(2) || commits["limit"] != float64(1) {
		t.Fatalf("expected commit pagination metadata, got %v", commits)
	}

	pullsRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/github/repositories/"+repositoryID+"/pulls?page=2&limit=1&state=open", adminToken, nil)
	assertStatus(t, pullsRec, http.StatusOK)
	pullItems, _ := decodeDataMap(t, pullsRec)["items"].([]any)
	if len(pullItems) != 1 || pullItems[0].(map[string]any)["number"] != float64(1) {
		t.Fatalf("expected 1 pull request #1, got %v", pullItems)
	}
	if pulls := decodeDataMap(t, pullsRec); pulls["page"] != float64(2) || pulls["limit"] != float64(1) || pulls["state"] != "open" {
		t.Fatalf("expected pull-request pagination and state metadata, got %v", pulls)
	}
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/github/repositories/"+repositoryID+"/pulls?state=invalid", adminToken, nil), http.StatusBadRequest)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/github/repositories/"+repositoryID+"/commits?limit=51", adminToken, nil), http.StatusBadRequest)

	unmapRec := doJSONRequest(t, app.router, http.MethodDelete, "/api/v1/applications/"+applicationID+"/github/repository", adminToken, nil)
	assertStatus(t, unmapRec, http.StatusOK)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/applications/"+applicationID+"/github/repository", adminToken, nil), http.StatusNotFound)
}

func TestGitHubIdentityEndpointIsSafeAndOrganizationScoped(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminAToken := registerAndLogin(t, app.router, "GitHub Identity Admin A", "github-identity-a@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "GitHub Identity Admin B", "github-identity-b@opspilot.dev", "password123")
	fake := newFakeGitHubServer(t, githubSuperSecretTestToken)
	app.githubClient.BaseURL = fake.server.URL
	integrationID := createGitHubIntegration(t, app, adminAToken, githubSuperSecretTestToken)

	identityRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/identity", adminAToken, nil)
	assertStatus(t, identityRec, http.StatusOK)
	identity := decodeDataMap(t, identityRec)
	if identity["id"] != float64(583231) || identity["login"] != "octocat" {
		t.Fatalf("expected safe GitHub identity, got %v", identity)
	}
	assertNoSecretLeak(t, identityRec.Body.String())
	if _, found := identity["access_token"]; found {
		t.Fatalf("identity response must not include an access token: %v", identity)
	}

	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/identity", adminBToken, nil), http.StatusNotFound)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/identity", "", nil), http.StatusUnauthorized)
}

func TestGitHubRepositoryMappingDeniesCrossOrganization(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminAToken := registerAndLogin(t, app.router, "GitHub Org A Admin", "github-org-a-admin@opspilot.dev", "password123")
	adminBToken := registerAndLogin(t, app.router, "GitHub Org B Admin", "github-org-b-admin@opspilot.dev", "password123")

	fake := newFakeGitHubServer(t, githubSuperSecretTestToken)
	app.githubClient.BaseURL = fake.server.URL

	integrationID := createGitHubIntegration(t, app, adminAToken, githubSuperSecretTestToken)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/github/repositories/sync", adminAToken, nil), http.StatusOK)
	items, _ := decodeDataMap(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/repositories", adminAToken, nil))["items"].([]any)
	repositoryID := items[0].(map[string]any)["id"].(string)

	projectBID := createProject(t, app.router, adminBToken, "Org B Project")
	appBRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectBID.String()+"/applications", adminBToken, map[string]any{
		"name": "Org B App", "runtime": constants.ApplicationRuntimeGo, "port": 8080,
	})
	assertStatus(t, appBRec, http.StatusCreated)
	applicationBID, _ := decodeDataMap(t, appBRec)["id"].(string)

	// Org B must not be able to map its application to org A's repository.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/applications/"+applicationBID+"/github/repository", adminBToken, map[string]any{
		"repository_id": repositoryID,
	}), http.StatusNotFound)

	// Org B must not see org A's repository via the commits/pulls endpoints either.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/github/repositories/"+repositoryID+"/commits", adminBToken, nil), http.StatusNotFound)
}

func TestGitHubRBACDeniesNonAdminMutations(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "GitHub RBAC Admin", "github-rbac-admin@opspilot.dev", "password123")
	registerAndLogin(t, app.router, "GitHub RBAC Viewer", "github-rbac-viewer@opspilot.dev", "password123")

	admin := mustGetUserByEmail(t, app.userRepo, "github-rbac-admin@opspilot.dev")
	viewer := mustGetUserByEmail(t, app.userRepo, "github-rbac-viewer@opspilot.dev")
	orgID := *admin.OrganizationID
	if err := app.userRepo.AssignOrganizationAndRole(viewer.ID, orgID, models.RoleViewer); err != nil {
		t.Fatalf("assign viewer role: %v", err)
	}
	viewerToken := loginOnly(t, app.router, "github-rbac-viewer@opspilot.dev", "password123")

	fake := newFakeGitHubServer(t, githubSuperSecretTestToken)
	app.githubClient.BaseURL = fake.server.URL

	integrationID := createGitHubIntegration(t, app, adminToken, githubSuperSecretTestToken)

	// Viewer cannot sync repositories or start OAuth, but can read.
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/github/repositories/sync", viewerToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/github/oauth/start", viewerToken, nil), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/repositories", viewerToken, nil), http.StatusOK)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/github/repositories/sync", adminToken, nil), http.StatusOK)
	items, _ := decodeDataMap(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/repositories", adminToken, nil))["items"].([]any)
	repositoryID := items[0].(map[string]any)["id"].(string)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/github/repositories/"+repositoryID+"/select", viewerToken, map[string]any{"selected": true}), http.StatusForbidden)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPatch, "/api/v1/github/repositories/"+repositoryID+"/select", adminToken, map[string]any{"selected": true}), http.StatusOK)
}

func TestGitHubDeploymentCorrelation(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "GitHub Correlation Admin", "github-correlation-admin@opspilot.dev", "password123")

	fake := newFakeGitHubServer(t, githubSuperSecretTestToken)
	app.githubClient.BaseURL = fake.server.URL

	integrationID := createGitHubIntegration(t, app, adminToken, githubSuperSecretTestToken)
	assertStatus(t, doJSONRequest(t, app.router, http.MethodPost, "/api/v1/integrations/"+integrationID+"/github/repositories/sync", adminToken, nil), http.StatusOK)
	items, _ := decodeDataMap(t, doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations/"+integrationID+"/github/repositories", adminToken, nil))["items"].([]any)
	repositoryID := items[0].(map[string]any)["id"].(string)

	projectID := createProject(t, app.router, adminToken, "Correlation Project")
	clusterID := createCluster(t, app.router, adminToken, projectID, "correlation-cluster")
	appRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/projects/"+projectID.String()+"/applications", adminToken, map[string]any{
		"name": "Correlation App", "runtime": constants.ApplicationRuntimeGo, "port": 8080,
	})
	assertStatus(t, appRec, http.StatusCreated)
	applicationID, _ := decodeDataMap(t, appRec)["id"].(string)

	assertStatus(t, doJSONRequest(t, app.router, http.MethodPut, "/api/v1/applications/"+applicationID+"/github/repository", adminToken, map[string]any{
		"repository_id": repositoryID,
	}), http.StatusOK)

	// A deployment with no commit SHA - correlation must honestly report unavailable.
	noShaRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
		"applicationId": applicationID, "projectId": projectID.String(), "targetClusterId": clusterID.String(),
		"image": "ghcr.io/acme/widgets", "imageTag": "v1", "environment": "production", "namespace": "widgets",
		"replicaCount": 1, "deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
	})
	assertStatus(t, noShaRec, http.StatusCreated)
	noShaDeploymentID, _ := decodeDataMap(t, noShaRec)["id"].(string)

	noShaCorrelationRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+noShaDeploymentID+"/github/correlation", adminToken, nil)
	assertStatus(t, noShaCorrelationRec, http.StatusOK)
	noShaCorrelation := decodeDataMap(t, noShaCorrelationRec)
	if noShaCorrelation["available"] != false {
		t.Fatalf("expected correlation to be unavailable with no commit SHA, got %v", noShaCorrelation)
	}

	// A deployment with a real, resolvable commit SHA.
	withShaRec := doJSONRequest(t, app.router, http.MethodPost, "/api/v1/deployments", adminToken, map[string]any{
		"applicationId": applicationID, "projectId": projectID.String(), "targetClusterId": clusterID.String(),
		"image": "ghcr.io/acme/widgets", "imageTag": "v2", "environment": "production", "namespace": "widgets",
		"replicaCount": 1, "deploymentStrategy": constants.DeploymentStrategyRollingUpdate,
		"commitSha": "deadbeef",
	})
	assertStatus(t, withShaRec, http.StatusCreated)
	withShaDeploymentID, _ := decodeDataMap(t, withShaRec)["id"].(string)

	correlationRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/deployments/"+withShaDeploymentID+"/github/correlation", adminToken, nil)
	assertStatus(t, correlationRec, http.StatusOK)
	correlation := decodeDataMap(t, correlationRec)
	if correlation["available"] != true {
		t.Fatalf("expected correlation to be available for a real commit SHA, got %v", correlation)
	}
	commit := correlation["commit"].(map[string]any)
	if commit["sha"] != "deadbeef" || commit["message"] != "Fix payment timeout" {
		t.Fatalf("expected the resolved commit to match the fake GitHub server, got %v", commit)
	}
	pullRequests, _ := correlation["pullRequests"].([]any)
	if len(pullRequests) != 1 || pullRequests[0].(map[string]any)["number"] != float64(421) {
		t.Fatalf("expected 1 associated pull request #421, got %v", pullRequests)
	}
}

func TestGitHubOAuthFullRoundTrip(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	adminToken := registerAndLogin(t, app.router, "GitHub OAuth Admin", "github-oauth-admin@opspilot.dev", "password123")

	fakeGitHub := newFakeGitHubServer(t, githubSuperSecretTestToken)
	app.githubClient.BaseURL = fakeGitHub.server.URL

	fakeTokenExchange := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": githubSuperSecretTestToken, "token_type": "bearer"})
	}))
	defer fakeTokenExchange.Close()
	app.githubOAuthConfig.AccessTokenURL = fakeTokenExchange.URL

	startRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/github/oauth/start", adminToken, nil)
	assertStatus(t, startRec, http.StatusOK)
	authorizeURL, _ := decodeDataMap(t, startRec)["authorizeUrl"].(string)
	if authorizeURL == "" {
		t.Fatalf("expected a non-empty authorize URL")
	}
	assertNoSecretLeak(t, startRec.Body.String())

	parsed, err := url.Parse(authorizeURL)
	if err != nil {
		t.Fatalf("parse authorize url: %v", err)
	}
	state := parsed.Query().Get("state")
	if state == "" {
		t.Fatalf("expected a state parameter in the authorize URL")
	}

	callbackURL := "/api/v1/github/oauth/callback?code=one-time-code&state=" + url.QueryEscape(state)
	callbackRec := doJSONRequest(t, app.router, http.MethodGet, callbackURL, "", nil)
	if callbackRec.Code != http.StatusFound {
		t.Fatalf("expected a redirect after a successful OAuth callback, got %d body=%s", callbackRec.Code, callbackRec.Body.String())
	}
	location := callbackRec.Header().Get("Location")
	if !strings.Contains(location, "github=connected") {
		t.Fatalf("expected redirect to indicate success, got %q", location)
	}
	assertNoSecretLeak(t, callbackRec.Body.String())

	// The resulting integration must exist, be CONNECTED, and never expose the token.
	listRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations", adminToken, nil)
	items := decodeIntegrationItems(t, listRec)
	var found map[string]any
	for _, item := range items {
		if item["type"] == constants.IntegrationTypeGitHub {
			found = item
		}
	}
	if found == nil {
		t.Fatalf("expected an OAuth-created GitHub integration to exist, got %v", items)
	}
	if found["status"] != constants.IntegrationStatusConnected {
		t.Fatalf("expected the OAuth-created integration to be CONNECTED, got %v", found)
	}
	assertNoSecretLeak(t, listRec.Body.String())

	// Reconnecting must update the same integration, not create a duplicate.
	callbackRec2 := doJSONRequest(t, app.router, http.MethodGet, callbackURL, "", nil)
	if callbackRec2.Code != http.StatusFound {
		t.Fatalf("expected the second callback to also redirect, got %d", callbackRec2.Code)
	}
	listRec2 := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/integrations", adminToken, nil)
	githubCount := 0
	for _, item := range decodeIntegrationItems(t, listRec2) {
		if item["type"] == constants.IntegrationTypeGitHub {
			githubCount++
		}
	}
	if githubCount != 1 {
		t.Fatalf("expected reconnecting via OAuth to update the same integration, got %d github integrations", githubCount)
	}

	// The audit trail for this organization must never contain the token.
	auditItems := fetchAllAuditItems(t, app.router, adminToken)
	for _, item := range auditItems {
		encoded, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("marshal audit item: %v", err)
		}
		assertNoSecretLeak(t, string(encoded))
	}
}

func TestGitHubOAuthCallbackRejectsInvalidState(t *testing.T) {
	t.Parallel()

	app := setupRBACApp(t)
	registerAndLogin(t, app.router, "GitHub OAuth Invalid State Admin", "github-oauth-invalid-admin@opspilot.dev", "password123")

	callbackRec := doJSONRequest(t, app.router, http.MethodGet, "/api/v1/github/oauth/callback?code=x&state=tampered.state", "", nil)
	if callbackRec.Code != http.StatusFound {
		t.Fatalf("expected the callback to redirect even on failure, got %d", callbackRec.Code)
	}
	location := callbackRec.Header().Get("Location")
	if !strings.Contains(location, "github=error") {
		t.Fatalf("expected redirect to indicate an error for a tampered state, got %q", location)
	}
}
