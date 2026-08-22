package integration

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/services"
)

func TestApplicationTeamAssignmentIntegration(t *testing.T) {
	db := setupSQLiteIntegrationDB(t)

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	applicationRepo := repository.NewApplicationRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	applicationTeamRepo := repository.NewApplicationTeamRepository(db)
	service := services.NewApplicationTeamService(applicationTeamRepo, applicationRepo, teamRepo)

	adminA := mustCreateTeamTestUser(t, userRepo, "Application Team Admin A", "at-admin-a@opspilot.dev")
	adminB := mustCreateTeamTestUser(t, userRepo, "Application Team Admin B", "at-admin-b@opspilot.dev")
	userA := mustCreateTeamTestUser(t, userRepo, "Application Team User A", "at-user-a@opspilot.dev")

	orgA := mustCreateTeamTestOrganization(t, organizationRepo, "Application Team Org A", "application-team-org-a", adminA.ID)
	orgB := mustCreateTeamTestOrganization(t, organizationRepo, "Application Team Org B", "application-team-org-b", adminB.ID)

	mustAssignUserOrganization(t, userRepo, adminA.ID, orgA)
	mustAssignUserOrganization(t, userRepo, adminB.ID, orgB)
	mustAssignUserOrganization(t, userRepo, userA.ID, orgA)
	if err := userRepo.AssignOrganizationAndRole(adminA.ID, orgA, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin role for orgA admin: %v", err)
	}
	if err := userRepo.AssignOrganizationAndRole(adminB.ID, orgB, models.RolePlatformAdmin); err != nil {
		t.Fatalf("assign admin role for orgB admin: %v", err)
	}
	if err := userRepo.AssignOrganizationAndRole(userA.ID, orgA, models.RoleViewer); err != nil {
		t.Fatalf("assign user role for orgA member: %v", err)
	}

	projectA := mustCreateProjectForOrg(t, projectRepo, orgA, adminA.ID, "Application Team Project A")
	projectB := mustCreateProjectForOrg(t, projectRepo, orgB, adminB.ID, "Application Team Project B")
	appA := mustCreateApplicationForProject(t, applicationRepo, orgA, projectA.ID, "Application Team App A")
	appB := mustCreateApplicationForProject(t, applicationRepo, orgB, projectB.ID, "Application Team App B")
	teamA := mustCreateTeamForOrg(t, teamRepo, orgA, "App Alpha Team")
	teamA2 := mustCreateTeamForOrg(t, teamRepo, orgA, "App Beta Team")
	teamB := mustCreateTeamForOrg(t, teamRepo, orgB, "App Gamma Team")

	t.Run("Assign Team", func(t *testing.T) {
		assigned, err := service.AssignTeam(models.RolePlatformAdmin, orgA, appA.ID, teamA.ID)
		if err != nil {
			t.Fatalf("assign team: %v", err)
		}
		if assigned.ApplicationID != appA.ID.String() {
			t.Fatalf("unexpected application id in assignment")
		}
		if assigned.TeamID != teamA.ID.String() {
			t.Fatalf("unexpected team id in assignment")
		}
	})

	t.Run("Duplicate assignment rejected", func(t *testing.T) {
		_, err := service.AssignTeam(models.RolePlatformAdmin, orgA, appA.ID, teamA.ID)
		if !errors.Is(err, apperrors.ErrApplicationTeamAlreadyAssigned) {
			t.Fatalf("expected ErrApplicationTeamAlreadyAssigned, got %v", err)
		}
	})

	t.Run("Cross-organization assignment rejected", func(t *testing.T) {
		_, err := service.AssignTeam(models.RolePlatformAdmin, orgA, appA.ID, teamB.ID)
		if !errors.Is(err, apperrors.ErrApplicationForbidden) {
			t.Fatalf("expected ErrApplicationForbidden, got %v", err)
		}
	})

	t.Run("Cross-organization application id rejected", func(t *testing.T) {
		_, err := service.AssignTeam(models.RolePlatformAdmin, orgA, appB.ID, teamA.ID)
		if !errors.Is(err, apperrors.ErrApplicationNotFound) {
			t.Fatalf("expected ErrApplicationNotFound, got %v", err)
		}
	})

	t.Run("Viewer forbidden to assign", func(t *testing.T) {
		_, err := service.AssignTeam(models.RoleViewer, orgA, appA.ID, teamA2.ID)
		if !errors.Is(err, apperrors.ErrApplicationForbidden) {
			t.Fatalf("expected ErrApplicationForbidden, got %v", err)
		}
	})

	t.Run("Developer forbidden to assign", func(t *testing.T) {
		_, err := service.AssignTeam(models.RoleDeveloper, orgA, appA.ID, teamA2.ID)
		if !errors.Is(err, apperrors.ErrApplicationForbidden) {
			t.Fatalf("expected ErrApplicationForbidden, got %v", err)
		}
	})

	t.Run("DevOps Engineer forbidden to assign (read only)", func(t *testing.T) {
		_, err := service.AssignTeam(models.RoleDevOpsEngineer, orgA, appA.ID, teamA2.ID)
		if !errors.Is(err, apperrors.ErrApplicationForbidden) {
			t.Fatalf("expected ErrApplicationForbidden, got %v", err)
		}
	})

	t.Run("Admin allowed", func(t *testing.T) {
		_, err := service.AssignTeam(models.RolePlatformAdmin, orgA, appA.ID, teamA2.ID)
		if err != nil {
			t.Fatalf("admin assign should succeed: %v", err)
		}
	})

	t.Run("List Application Teams", func(t *testing.T) {
		result, err := service.ListApplicationTeams(models.RoleViewer, orgA, appA.ID)
		if err != nil {
			t.Fatalf("list application teams: %v", err)
		}
		if result.Total != 2 {
			t.Fatalf("expected 2 application-team mappings, got %d", result.Total)
		}
	})

	t.Run("DevOps Engineer can read", func(t *testing.T) {
		result, err := service.ListApplicationTeams(models.RoleDevOpsEngineer, orgA, appA.ID)
		if err != nil {
			t.Fatalf("devops engineer list application teams: %v", err)
		}
		if result.Total != 2 {
			t.Fatalf("expected 2 application-team mappings, got %d", result.Total)
		}
	})

	t.Run("List Team Applications", func(t *testing.T) {
		result, err := service.ListTeamApplications(models.RoleViewer, orgA, teamA.ID)
		if err != nil {
			t.Fatalf("list team applications: %v", err)
		}
		if result.Total != 1 {
			t.Fatalf("expected 1 mapping for team, got %d", result.Total)
		}
	})

	t.Run("Organization isolation maintained", func(t *testing.T) {
		result, err := service.ListApplicationTeams(models.RolePlatformAdmin, orgB, appB.ID)
		if err != nil {
			t.Fatalf("list orgB application teams: %v", err)
		}
		if result.Total != 0 {
			t.Fatalf("expected 0 mappings in orgB application, got %d", result.Total)
		}
	})

	t.Run("Remove assignment", func(t *testing.T) {
		if err := service.RemoveTeam(models.RolePlatformAdmin, orgA, appA.ID, teamA2.ID); err != nil {
			t.Fatalf("remove team assignment: %v", err)
		}

		isAssigned, err := applicationTeamRepo.IsAssigned(appA.ID, teamA2.ID, orgA)
		if err != nil {
			t.Fatalf("check assignment after removal: %v", err)
		}
		if isAssigned {
			t.Fatalf("expected assignment removed")
		}
	})

	t.Run("Viewer forbidden to remove", func(t *testing.T) {
		if err := service.RemoveTeam(models.RoleViewer, orgA, appA.ID, teamA.ID); !errors.Is(err, apperrors.ErrApplicationForbidden) {
			t.Fatalf("expected ErrApplicationForbidden, got %v", err)
		}
	})
}

func mustCreateApplicationForProject(
	t *testing.T,
	applicationRepo *repository.ApplicationRepository,
	organizationID uuid.UUID,
	projectID uuid.UUID,
	name string,
) *models.Application {
	t.Helper()

	application := &models.Application{
		OrganizationID: organizationID,
		ProjectID:      projectID,
		Name:           name,
		Slug:           name + "-slug",
		Runtime:        constants.ApplicationRuntimeGo,
		Port:           8080,
		Status:         constants.ApplicationStatusDraft,
	}
	if err := applicationRepo.CreateApplication(application); err != nil {
		t.Fatalf("create application %s: %v", name, err)
	}

	return application
}
