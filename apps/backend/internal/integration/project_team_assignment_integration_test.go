package integration

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/services"
)

func TestProjectTeamAssignmentIntegration(t *testing.T) {
	db := setupSQLiteIntegrationDB(t)

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	projectTeamRepo := repository.NewProjectTeamRepository(db)
	service := services.NewProjectTeamService(projectTeamRepo, projectRepo, teamRepo)

	adminA := mustCreateTeamTestUser(t, userRepo, "Project Team Admin A", "pt-admin-a@opspilot.dev")
	adminB := mustCreateTeamTestUser(t, userRepo, "Project Team Admin B", "pt-admin-b@opspilot.dev")
	userA := mustCreateTeamTestUser(t, userRepo, "Project Team User A", "pt-user-a@opspilot.dev")

	orgA := mustCreateTeamTestOrganization(t, organizationRepo, "Project Team Org A", "project-team-org-a", adminA.ID)
	orgB := mustCreateTeamTestOrganization(t, organizationRepo, "Project Team Org B", "project-team-org-b", adminB.ID)

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

	projectA := mustCreateProjectForOrg(t, projectRepo, orgA, adminA.ID, "Project Team A")
	projectB := mustCreateProjectForOrg(t, projectRepo, orgB, adminB.ID, "Project Team B")
	teamA := mustCreateTeamForOrg(t, teamRepo, orgA, "Alpha Team")
	teamA2 := mustCreateTeamForOrg(t, teamRepo, orgA, "Beta Team")
	teamB := mustCreateTeamForOrg(t, teamRepo, orgB, "Gamma Team")

	t.Run("Assign Team", func(t *testing.T) {
		assigned, err := service.AssignTeam(models.RolePlatformAdmin, orgA, projectA.ID, teamA.ID)
		if err != nil {
			t.Fatalf("assign team: %v", err)
		}
		if assigned.ProjectID != projectA.ID.String() {
			t.Fatalf("unexpected project id in assignment")
		}
		if assigned.TeamID != teamA.ID.String() {
			t.Fatalf("unexpected team id in assignment")
		}
	})

	t.Run("Duplicate assignment rejected", func(t *testing.T) {
		_, err := service.AssignTeam(models.RolePlatformAdmin, orgA, projectA.ID, teamA.ID)
		if !errors.Is(err, apperrors.ErrProjectTeamAlreadyAssigned) {
			t.Fatalf("expected ErrProjectTeamAlreadyAssigned, got %v", err)
		}
	})

	t.Run("Cross-organization assignment rejected", func(t *testing.T) {
		_, err := service.AssignTeam(models.RolePlatformAdmin, orgA, projectA.ID, teamB.ID)
		if !errors.Is(err, apperrors.ErrProjectForbidden) {
			t.Fatalf("expected ErrProjectForbidden, got %v", err)
		}
	})

	t.Run("Viewer forbidden to assign", func(t *testing.T) {
		_, err := service.AssignTeam(models.RoleViewer, orgA, projectA.ID, teamA2.ID)
		if !errors.Is(err, apperrors.ErrProjectForbidden) {
			t.Fatalf("expected ErrProjectForbidden, got %v", err)
		}
	})

	t.Run("Admin allowed", func(t *testing.T) {
		_, err := service.AssignTeam(models.RolePlatformAdmin, orgA, projectA.ID, teamA2.ID)
		if err != nil {
			t.Fatalf("admin assign should succeed: %v", err)
		}
	})

	t.Run("List Project Teams", func(t *testing.T) {
		result, err := service.ListProjectTeams(models.RoleViewer, orgA, projectA.ID)
		if err != nil {
			t.Fatalf("list project teams: %v", err)
		}
		if result.Total != 2 {
			t.Fatalf("expected 2 project-team mappings, got %d", result.Total)
		}
	})

	t.Run("List Team Projects", func(t *testing.T) {
		result, err := service.ListTeamProjects(models.RoleViewer, orgA, teamA.ID)
		if err != nil {
			t.Fatalf("list team projects: %v", err)
		}
		if result.Total != 1 {
			t.Fatalf("expected 1 mapping for team, got %d", result.Total)
		}
	})

	t.Run("Organization isolation maintained", func(t *testing.T) {
		result, err := service.ListProjectTeams(models.RolePlatformAdmin, orgB, projectB.ID)
		if err != nil {
			t.Fatalf("list orgB project teams: %v", err)
		}
		if result.Total != 0 {
			t.Fatalf("expected 0 mappings in orgB project, got %d", result.Total)
		}
	})

	t.Run("Remove assignment", func(t *testing.T) {
		if err := service.RemoveTeam(models.RolePlatformAdmin, orgA, projectA.ID, teamA2.ID); err != nil {
			t.Fatalf("remove team assignment: %v", err)
		}

		isAssigned, err := projectTeamRepo.IsAssigned(projectA.ID, teamA2.ID, orgA)
		if err != nil {
			t.Fatalf("check assignment after removal: %v", err)
		}
		if isAssigned {
			t.Fatalf("expected assignment removed")
		}
	})
}

func mustCreateProjectForOrg(
	t *testing.T,
	projectRepo *repository.ProjectRepository,
	organizationID uuid.UUID,
	ownerID uint,
	name string,
) *models.Project {
	t.Helper()

	project := &models.Project{
		OrganizationID: organizationID,
		Name:           name,
		Slug:           name + "-slug",
		Description:    "",
		Environment:    "development",
		Health:         "healthy",
		OwnerID:        ownerID,
		Members:        0,
		Services:       0,
	}
	if err := projectRepo.Create(project); err != nil {
		t.Fatalf("create project %s: %v", name, err)
	}

	return project
}

func mustCreateTeamForOrg(
	t *testing.T,
	teamRepo *repository.TeamRepository,
	organizationID uuid.UUID,
	name string,
) *models.Team {
	t.Helper()

	team := &models.Team{
		OrganizationID: organizationID,
		Name:           name,
		Description:    "",
	}
	if err := teamRepo.Create(team); err != nil {
		t.Fatalf("create team %s: %v", name, err)
	}

	return team
}
