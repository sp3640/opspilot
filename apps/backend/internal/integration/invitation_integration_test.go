package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/services"
)

func TestInvitationSystemIntegration(t *testing.T) {
	db := setupSQLiteIntegrationDB(t)

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	invitationService := services.NewInvitationService(invitationRepo, userRepo, organizationRepo)

	admin := mustCreateInvitationUser(t, userRepo, "Admin", "admin@opspilot.dev", models.RolePlatformAdmin)
	member := mustCreateInvitationUser(t, userRepo, "Member", "member@opspilot.dev", models.RoleViewer)
	otherUser := mustCreateInvitationUser(t, userRepo, "Other", "other@opspilot.dev", models.RoleViewer)

	orgA := mustCreateTeamTestOrganization(t, organizationRepo, "Org A", "org-a-invites", admin.ID)
	orgB := mustCreateTeamTestOrganization(t, organizationRepo, "Org B", "org-b-invites", otherUser.ID)

	mustAssignUserOrganization(t, userRepo, admin.ID, orgA)
	mustAssignUserOrganization(t, userRepo, member.ID, orgB)
	mustAssignUserOrganization(t, userRepo, otherUser.ID, orgB)

	var createdToken string
	var createdInvitationID uuid.UUID

	t.Run("Invite user", func(t *testing.T) {
		created, err := invitationService.InviteUser(admin.ID, models.RolePlatformAdmin, orgA, dto.InviteRequest{
			Email: member.Email,
			Role:  models.RoleViewer,
		})
		if err != nil {
			t.Fatalf("invite user: %v", err)
		}

		if created.Status != models.InvitationStatusPending {
			t.Fatalf("expected pending status, got %s", created.Status)
		}
		if created.Token == "" {
			t.Fatalf("expected token")
		}

		parsedID, err := uuid.Parse(created.ID)
		if err != nil {
			t.Fatalf("parse invitation id: %v", err)
		}
		createdInvitationID = parsedID
		createdToken = created.Token
	})

	t.Run("Duplicate invite rejected", func(t *testing.T) {
		_, err := invitationService.InviteUser(admin.ID, models.RolePlatformAdmin, orgA, dto.InviteRequest{
			Email: member.Email,
			Role:  models.RoleViewer,
		})
		if !errors.Is(err, apperrors.ErrInvitationAlreadyExists) {
			t.Fatalf("expected ErrInvitationAlreadyExists, got %v", err)
		}
	})

	t.Run("Wrong token rejected", func(t *testing.T) {
		_, err := invitationService.AcceptInvitation(member.ID, member.Email, dto.AcceptInvitationRequest{Token: "not-a-token"})
		if !errors.Is(err, apperrors.ErrInvitationNotFound) {
			t.Fatalf("expected ErrInvitationNotFound, got %v", err)
		}
	})

	t.Run("Wrong email rejected", func(t *testing.T) {
		_, err := invitationService.AcceptInvitation(otherUser.ID, otherUser.Email, dto.AcceptInvitationRequest{Token: createdToken})
		if !errors.Is(err, apperrors.ErrInvitationEmailMismatch) {
			t.Fatalf("expected ErrInvitationEmailMismatch, got %v", err)
		}
	})

	t.Run("Accept invite", func(t *testing.T) {
		accepted, err := invitationService.AcceptInvitation(member.ID, member.Email, dto.AcceptInvitationRequest{Token: createdToken})
		if err != nil {
			t.Fatalf("accept invite: %v", err)
		}
		if accepted.Status != models.InvitationStatusAccepted {
			t.Fatalf("expected accepted status, got %s", accepted.Status)
		}
		if accepted.AcceptedAt == nil {
			t.Fatalf("expected accepted_at to be set")
		}

		updatedUser, err := userRepo.GetByID(member.ID)
		if err != nil {
			t.Fatalf("reload user after acceptance: %v", err)
		}
		if updatedUser.OrganizationID == nil || *updatedUser.OrganizationID != orgA {
			t.Fatalf("expected user to join invited organization")
		}
		if updatedUser.Role != models.RoleViewer {
			t.Fatalf("expected role to be updated from invitation, got %s", updatedUser.Role)
		}
	})

	t.Run("Invitation status updated", func(t *testing.T) {
		stored, err := invitationRepo.GetByID(createdInvitationID)
		if err != nil {
			t.Fatalf("load accepted invitation: %v", err)
		}
		if stored.Status != models.InvitationStatusAccepted {
			t.Fatalf("expected accepted invitation status, got %s", stored.Status)
		}
	})

	t.Run("Expired invite", func(t *testing.T) {
		expiredInvitation := &models.Invitation{
			OrganizationID: orgA,
			Email:          "expired@opspilot.dev",
			Role:           models.RoleViewer,
			Token:          "expired-token",
			Status:         models.InvitationStatusPending,
			InvitedBy:      admin.ID,
			ExpiresAt:      time.Now().UTC().Add(-1 * time.Hour),
		}
		if err := invitationRepo.Create(expiredInvitation); err != nil {
			t.Fatalf("create expired invitation: %v", err)
		}

		_, err := invitationService.AcceptInvitation(member.ID, member.Email, dto.AcceptInvitationRequest{Token: expiredInvitation.Token})
		if !errors.Is(err, apperrors.ErrInvitationExpired) {
			t.Fatalf("expected ErrInvitationExpired, got %v", err)
		}
	})

	t.Run("Revoke invite", func(t *testing.T) {
		revoked, err := invitationService.InviteUser(admin.ID, models.RolePlatformAdmin, orgA, dto.InviteRequest{
			Email: "revoke@opspilot.dev",
			Role:  models.RoleViewer,
		})
		if err != nil {
			t.Fatalf("create invitation to revoke: %v", err)
		}

		revokedID, err := uuid.Parse(revoked.ID)
		if err != nil {
			t.Fatalf("parse revoked invitation id: %v", err)
		}

		if err := invitationService.RevokeInvitation(revokedID, models.RolePlatformAdmin, orgA); err != nil {
			t.Fatalf("revoke invitation: %v", err)
		}

		stored, err := invitationRepo.GetByID(revokedID)
		if err != nil {
			t.Fatalf("load revoked invitation: %v", err)
		}
		if stored.Status != models.InvitationStatusRevoked {
			t.Fatalf("expected revoked status, got %s", stored.Status)
		}
	})

	t.Run("Organization scoped listing", func(t *testing.T) {
		_, err := invitationService.InviteUser(admin.ID, models.RolePlatformAdmin, orgA, dto.InviteRequest{
			Email: "org-a-list@opspilot.dev",
			Role:  models.RoleViewer,
		})
		if err != nil {
			t.Fatalf("create orgA listing invitation: %v", err)
		}

		_, err = invitationService.InviteUser(otherUser.ID, models.RolePlatformAdmin, orgB, dto.InviteRequest{
			Email: "org-b-list@opspilot.dev",
			Role:  models.RoleViewer,
		})
		if err != nil {
			t.Fatalf("create orgB listing invitation: %v", err)
		}

		result, err := invitationService.ListInvitations(models.RolePlatformAdmin, orgA, &models.PaginationRequest{
			Page:  1,
			Limit: 50,
			Sort:  "created_at",
			Order: "desc",
		})
		if err != nil {
			t.Fatalf("list invitations for orgA: %v", err)
		}

		if len(result.Items) == 0 {
			t.Fatalf("expected at least one invitation in orgA")
		}

		for _, item := range result.Items {
			if item.OrganizationID != orgA.String() {
				t.Fatalf("expected invitation organization %s, got %s", orgA.String(), item.OrganizationID)
			}
		}
	})
}

func mustCreateInvitationUser(
	t *testing.T,
	userRepo *repository.UserRepository,
	name, email, role string,
) *models.User {
	t.Helper()

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: "hashed-password",
		Role:         role,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("create invitation user %s: %v", email, err)
	}

	return user
}
