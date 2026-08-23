package integration

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/services"
)

func TestTeamManagementFoundationIntegration(t *testing.T) {
	db := setupSQLiteIntegrationDB(t)

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	teamMemberRepo := repository.NewTeamMemberRepository(db)
	teamService := services.NewTeamService(teamRepo, teamMemberRepo, userRepo)

	ownerA := mustCreateTeamTestUser(t, userRepo, "Owner A", "owner-a@opspilot.dev")
	ownerB := mustCreateTeamTestUser(t, userRepo, "Owner B", "owner-b@opspilot.dev")

	orgA := mustCreateTeamTestOrganization(t, organizationRepo, "Org A", "org-a", ownerA.ID)
	orgB := mustCreateTeamTestOrganization(t, organizationRepo, "Org B", "org-b", ownerB.ID)

	mustAssignUserOrganization(t, userRepo, ownerA.ID, orgA)
	mustAssignUserOrganization(t, userRepo, ownerB.ID, orgB)

	memberA := mustCreateTeamTestUser(t, userRepo, "Member A", "member-a@opspilot.dev")
	memberA2 := mustCreateTeamTestUser(t, userRepo, "Member A2", "member-a2@opspilot.dev")
	memberB := mustCreateTeamTestUser(t, userRepo, "Member B", "member-b@opspilot.dev")

	mustAssignUserOrganization(t, userRepo, memberA.ID, orgA)
	mustAssignUserOrganization(t, userRepo, memberA2.ID, orgA)
	mustAssignUserOrganization(t, userRepo, memberB.ID, orgB)

	var coreTeamID uuid.UUID

	t.Run("Create Team", func(t *testing.T) {
		created, err := teamService.CreateTeam(ownerA.ID, orgA, dto.CreateTeamRequest{
			Name:        "Platform",
			Description: "Core platform team",
		})
		if err != nil {
			t.Fatalf("create team: %v", err)
		}
		if created.ID == "" {
			t.Fatalf("expected team id")
		}
		if created.OrganizationID != orgA.String() {
			t.Fatalf("unexpected organization id: got %s want %s", created.OrganizationID, orgA.String())
		}

		parsed, parseErr := uuid.Parse(created.ID)
		if parseErr != nil {
			t.Fatalf("parse team id: %v", parseErr)
		}
		coreTeamID = parsed
	})

	t.Run("Duplicate Team Name Same Organization", func(t *testing.T) {
		_, err := teamService.CreateTeam(ownerA.ID, orgA, dto.CreateTeamRequest{
			Name:        "Platform",
			Description: "Duplicate name",
		})
		if !errors.Is(err, apperrors.ErrTeamAlreadyExists) {
			t.Fatalf("expected ErrTeamAlreadyExists, got %v", err)
		}
	})

	t.Run("Same Team Name Different Organization", func(t *testing.T) {
		created, err := teamService.CreateTeam(ownerB.ID, orgB, dto.CreateTeamRequest{
			Name:        "Platform",
			Description: "Allowed in different org",
		})
		if err != nil {
			t.Fatalf("create team in other org: %v", err)
		}
		if created.OrganizationID != orgB.String() {
			t.Fatalf("unexpected organization id in orgB team")
		}
	})

	t.Run("Update Team", func(t *testing.T) {
		updated, err := teamService.UpdateTeam(ownerA.ID, coreTeamID, orgA, dto.UpdateTeamRequest{
			Name:        "Platform Engineering",
			Description: "Updated description",
		})
		if err != nil {
			t.Fatalf("update team: %v", err)
		}
		if updated.Name != "Platform Engineering" {
			t.Fatalf("unexpected updated name: %s", updated.Name)
		}
	})

	t.Run("Add Member", func(t *testing.T) {
		member, err := teamService.AddMember(ownerA.ID, coreTeamID, orgA, memberA.ID)
		if err != nil {
			t.Fatalf("add member: %v", err)
		}
		if member.UserID != memberA.ID {
			t.Fatalf("unexpected member user id")
		}
	})

	t.Run("Prevent Duplicate Membership", func(t *testing.T) {
		_, err := teamService.AddMember(ownerA.ID, coreTeamID, orgA, memberA.ID)
		if !errors.Is(err, apperrors.ErrTeamMemberAlreadyExists) {
			t.Fatalf("expected ErrTeamMemberAlreadyExists, got %v", err)
		}
	})

	t.Run("Prevent Cross-Organization Membership", func(t *testing.T) {
		_, err := teamService.AddMember(ownerA.ID, coreTeamID, orgA, memberB.ID)
		if !errors.Is(err, apperrors.ErrTeamForbidden) {
			t.Fatalf("expected ErrTeamForbidden, got %v", err)
		}
	})

	t.Run("List Members", func(t *testing.T) {
		if _, err := teamService.AddMember(ownerA.ID, coreTeamID, orgA, memberA2.ID); err != nil {
			t.Fatalf("add second member: %v", err)
		}

		members, err := teamService.ListMembers(coreTeamID, orgA)
		if err != nil {
			t.Fatalf("list members: %v", err)
		}
		if members.Total != 2 {
			t.Fatalf("expected 2 members, got %d", members.Total)
		}
	})

	t.Run("Remove Member", func(t *testing.T) {
		if err := teamService.RemoveMember(ownerA.ID, coreTeamID, orgA, memberA.ID); err != nil {
			t.Fatalf("remove member: %v", err)
		}
		isMember, err := teamMemberRepo.IsMember(coreTeamID, memberA.ID)
		if err != nil {
			t.Fatalf("check membership after remove: %v", err)
		}
		if isMember {
			t.Fatalf("expected member to be removed")
		}
	})

	t.Run("Delete Team", func(t *testing.T) {
		if err := teamService.DeleteTeam(ownerA.ID, coreTeamID, orgA); err != nil {
			t.Fatalf("delete team: %v", err)
		}

		_, err := teamService.GetTeamByID(coreTeamID, orgA)
		if !errors.Is(err, apperrors.ErrTeamNotFound) {
			t.Fatalf("expected ErrTeamNotFound after delete, got %v", err)
		}

		remainingMembers, err := teamMemberRepo.CountMembers(coreTeamID)
		if err != nil {
			t.Fatalf("count members after team delete: %v", err)
		}
		if remainingMembers != 0 {
			t.Fatalf("expected 0 members after team delete, got %d", remainingMembers)
		}
	})

}

func mustCreateTeamTestUser(
	t *testing.T,
	userRepo *repository.UserRepository,
	name, email string,
) *models.User {
	t.Helper()

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: "hashed-password",
		Role:         models.RoleUser,
	}

	if err := userRepo.Create(user); err != nil {
		t.Fatalf("create team test user %s: %v", email, err)
	}

	return user
}

func mustCreateTeamTestOrganization(
	t *testing.T,
	organizationRepo *repository.OrganizationRepository,
	name, slug string,
	ownerID uint,
) uuid.UUID {
	t.Helper()

	organization := &models.Organization{
		Name:        name,
		Slug:        slug,
		Description: "",
		OwnerID:     ownerID,
	}

	if err := organizationRepo.Create(organization); err != nil {
		t.Fatalf("create team test organization %s: %v", name, err)
	}

	return organization.ID
}

func mustAssignUserOrganization(t *testing.T, userRepo *repository.UserRepository, userID uint, organizationID uuid.UUID) {
	t.Helper()

	if err := userRepo.AssignOrganization(userID, organizationID); err != nil {
		t.Fatalf("assign organization %s to user %d: %v", organizationID.String(), userID, err)
	}
}
