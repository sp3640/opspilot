package integration

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestInvitationAwareRegistrationIntegration covers UserService.Register:
// every registrant, invited or not, first or not, always creates their own
// new workspace and becomes its Platform Admin. Registration alone never
// proves the registrant controls the email address, so a pending invitation
// must never be enough to grant membership in someone else's organization -
// that requires the token-verified InvitationService.AcceptInvitation flow
// instead (see registration_workspace_integration_test.go for the full
// register-then-accept round trip over HTTP). Subtests run in sequence
// against a single isolated database because the first subtest depends on
// being the very first user ever registered.
func TestInvitationAwareRegistrationIntegration(t *testing.T) {
	db := setupSQLiteInvitationAwareRegistrationDB(t)

	userRepo := repository.NewUserRepository(db)
	organizationRepo := repository.NewOrganizationRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	cfg := &config.Config{JWTSecret: "this-is-a-very-long-test-jwt-secret-1234567890"}
	userService := services.NewUserService(userRepo, organizationRepo, invitationRepo, cfg)
	invitationService := services.NewInvitationService(invitationRepo, userRepo, organizationRepo)

	var firstOrgID uuid.UUID

	t.Run("First-ever registration creates a new workspace and becomes Platform Admin", func(t *testing.T) {
		if err := userService.Register("Alice Admin", "alice-first@opspilot.dev", "password123", ""); err != nil {
			t.Fatalf("register first user: %v", err)
		}

		user, err := userRepo.GetByEmail("alice-first@opspilot.dev")
		if err != nil {
			t.Fatalf("load first user: %v", err)
		}
		if user.Role != models.RolePlatformAdmin {
			t.Fatalf("expected Platform Admin role, got %s", user.Role)
		}
		if user.OrganizationID == nil {
			t.Fatalf("expected first user to be assigned an organization")
		}

		organization, err := organizationRepo.GetByID(*user.OrganizationID)
		if err != nil {
			t.Fatalf("load first user organization: %v", err)
		}
		if organization.OwnerID != user.ID {
			t.Fatalf("expected first user to own the new workspace")
		}

		firstOrgID = organization.ID
	})

	t.Run("Registration without an invitation creates its own new workspace and becomes Platform Admin", func(t *testing.T) {
		if err := userService.Register("Bob Member", "bob-normal@opspilot.dev", "password123", "Bob's Workspace"); err != nil {
			t.Fatalf("register normal user: %v", err)
		}

		user, err := userRepo.GetByEmail("bob-normal@opspilot.dev")
		if err != nil {
			t.Fatalf("load normal user: %v", err)
		}
		if user.Role != models.RolePlatformAdmin {
			t.Fatalf("expected Platform Admin role, got %s", user.Role)
		}
		if user.OrganizationID == nil {
			t.Fatalf("expected normal registration to be assigned a new organization")
		}
		if *user.OrganizationID == firstOrgID {
			t.Fatalf("expected normal registration to create its own organization, not join the first user's")
		}

		organization, err := organizationRepo.GetByID(*user.OrganizationID)
		if err != nil {
			t.Fatalf("load normal user's organization: %v", err)
		}
		if organization.OwnerID != user.ID {
			t.Fatalf("expected normal registration to own its new workspace")
		}
		if organization.Name != "Bob's Workspace" {
			t.Fatalf("expected requested organization name to be honored, got %q", organization.Name)
		}
	})

	t.Run("Registration with a valid pending invitation still creates its own workspace, not the invited org", func(t *testing.T) {
		admin, err := userRepo.GetByEmail("alice-first@opspilot.dev")
		if err != nil {
			t.Fatalf("load admin: %v", err)
		}

		invitedOrgID := mustCreateTeamTestOrganization(t, organizationRepo, "Invited Org", "invited-org-registration", admin.ID)
		invitation, err := invitationService.InviteUser(admin.ID, models.RolePlatformAdmin, invitedOrgID, dto.InviteRequest{
			Email: "carol-invited@opspilot.dev",
			Role:  models.RolePlatformAdmin,
		})
		if err != nil {
			t.Fatalf("create invitation: %v", err)
		}

		if err := userService.Register("Carol Invited", "carol-invited@opspilot.dev", "password123", "Carol's Workspace"); err != nil {
			t.Fatalf("register invited user: %v", err)
		}

		user, err := userRepo.GetByEmail("carol-invited@opspilot.dev")
		if err != nil {
			t.Fatalf("load invited user: %v", err)
		}
		if user.Role != models.RolePlatformAdmin {
			t.Fatalf("expected fallback to Platform Admin role, got %s", user.Role)
		}
		// A pending invitation must never be enough, by itself, to join its
		// organization at registration time - registration alone doesn't
		// prove the registrant controls the invited email address. Only the
		// token-verified accept-invitation flow may grant that membership.
		if user.OrganizationID == nil || *user.OrganizationID == invitedOrgID || *user.OrganizationID == firstOrgID {
			t.Fatalf("expected registration to create its own workspace, not join the invited organization via email match alone")
		}

		organization, err := organizationRepo.GetByID(*user.OrganizationID)
		if err != nil {
			t.Fatalf("load invited user's new workspace: %v", err)
		}
		if organization.Name != "Carol's Workspace" {
			t.Fatalf("expected requested organization name to be honored, got %q", organization.Name)
		}

		// Registration must not itself consume the invitation — acceptance
		// remains the responsibility of the existing, token-verified accept
		// flow (InvitationService.AcceptInvitation).
		invitationID, err := uuid.Parse(invitation.ID)
		if err != nil {
			t.Fatalf("parse invitation id: %v", err)
		}
		stored, err := invitationRepo.GetByID(invitationID)
		if err != nil {
			t.Fatalf("reload invitation: %v", err)
		}
		if stored.Status != models.InvitationStatusPending {
			t.Fatalf("expected invitation to remain pending after registration, got %s", stored.Status)
		}
	})

	t.Run("Registration with an expired invitation falls back to creating a new workspace", func(t *testing.T) {
		admin, err := userRepo.GetByEmail("alice-first@opspilot.dev")
		if err != nil {
			t.Fatalf("load admin: %v", err)
		}

		staleOrgID := mustCreateTeamTestOrganization(t, organizationRepo, "Stale Org", "stale-org-registration", admin.ID)
		expiredInvitation := &models.Invitation{
			OrganizationID: staleOrgID,
			Email:          "dave-expired@opspilot.dev",
			Role:           models.RolePlatformAdmin,
			Token:          "expired-registration-token",
			Status:         models.InvitationStatusPending,
			InvitedBy:      admin.ID,
			ExpiresAt:      time.Now().UTC().Add(-1 * time.Hour),
		}
		if err := invitationRepo.Create(expiredInvitation); err != nil {
			t.Fatalf("create expired invitation: %v", err)
		}

		if err := userService.Register("Dave Expired", "dave-expired@opspilot.dev", "password123", ""); err != nil {
			t.Fatalf("register user with expired invitation: %v", err)
		}

		user, err := userRepo.GetByEmail("dave-expired@opspilot.dev")
		if err != nil {
			t.Fatalf("load user: %v", err)
		}
		if user.Role != models.RolePlatformAdmin {
			t.Fatalf("expected fallback to Platform Admin role, got %s", user.Role)
		}
		if user.OrganizationID == nil || *user.OrganizationID == staleOrgID || *user.OrganizationID == firstOrgID {
			t.Fatalf("expected fallback to a brand new workspace, not the expired invitation's organization or the first user's")
		}
	})

	t.Run("Registration with a revoked invitation falls back to creating a new workspace", func(t *testing.T) {
		admin, err := userRepo.GetByEmail("alice-first@opspilot.dev")
		if err != nil {
			t.Fatalf("load admin: %v", err)
		}

		revokedOrgID := mustCreateTeamTestOrganization(t, organizationRepo, "Revoked Org", "revoked-org-registration", admin.ID)
		invitation, err := invitationService.InviteUser(admin.ID, models.RolePlatformAdmin, revokedOrgID, dto.InviteRequest{
			Email: "erin-revoked@opspilot.dev",
			Role:  models.RolePlatformAdmin,
		})
		if err != nil {
			t.Fatalf("create invitation to revoke: %v", err)
		}

		invitationID, err := uuid.Parse(invitation.ID)
		if err != nil {
			t.Fatalf("parse invitation id: %v", err)
		}
		if err := invitationService.RevokeInvitation(invitationID, models.RolePlatformAdmin, admin.ID, revokedOrgID); err != nil {
			t.Fatalf("revoke invitation: %v", err)
		}

		if err := userService.Register("Erin Revoked", "erin-revoked@opspilot.dev", "password123", ""); err != nil {
			t.Fatalf("register user with revoked invitation: %v", err)
		}

		user, err := userRepo.GetByEmail("erin-revoked@opspilot.dev")
		if err != nil {
			t.Fatalf("load user: %v", err)
		}
		if user.Role != models.RolePlatformAdmin {
			t.Fatalf("expected fallback to Platform Admin role, got %s", user.Role)
		}
		if user.OrganizationID == nil || *user.OrganizationID == revokedOrgID || *user.OrganizationID == firstOrgID {
			t.Fatalf("expected fallback to a brand new workspace, not the revoked invitation's organization or the first user's")
		}
	})

	t.Run("Registration unrelated to a pending invitation for a different email creates its own workspace", func(t *testing.T) {
		admin, err := userRepo.GetByEmail("alice-first@opspilot.dev")
		if err != nil {
			t.Fatalf("load admin: %v", err)
		}

		otherOrgID := mustCreateTeamTestOrganization(t, organizationRepo, "Other Org", "other-org-registration", admin.ID)
		if _, err := invitationService.InviteUser(admin.ID, models.RolePlatformAdmin, otherOrgID, dto.InviteRequest{
			Email: "frank-invited-someone-else@opspilot.dev",
			Role:  models.RolePlatformAdmin,
		}); err != nil {
			t.Fatalf("create unrelated invitation: %v", err)
		}

		if err := userService.Register("Grace Unrelated", "grace-unrelated@opspilot.dev", "password123", ""); err != nil {
			t.Fatalf("register unrelated user: %v", err)
		}

		user, err := userRepo.GetByEmail("grace-unrelated@opspilot.dev")
		if err != nil {
			t.Fatalf("load user: %v", err)
		}
		if user.Role != models.RolePlatformAdmin {
			t.Fatalf("expected normal Platform Admin role, got %s", user.Role)
		}
		if user.OrganizationID == nil || *user.OrganizationID == otherOrgID || *user.OrganizationID == firstOrgID {
			t.Fatalf("expected registration to create its own workspace, unaffected by an unrelated invitation")
		}
	})

	t.Run("Two uninvited registrants with the same default workspace name do not collide", func(t *testing.T) {
		if err := userService.Register("Henry Duplicate Name", "henry-one@opspilot.dev", "password123", "Shared Name"); err != nil {
			t.Fatalf("register first Henry: %v", err)
		}
		if err := userService.Register("Henry Duplicate Name", "henry-two@opspilot.dev", "password123", "Shared Name"); err != nil {
			t.Fatalf("expected duplicate workspace name to be handled without error: %v", err)
		}

		userOne, err := userRepo.GetByEmail("henry-one@opspilot.dev")
		if err != nil {
			t.Fatalf("load first Henry: %v", err)
		}
		userTwo, err := userRepo.GetByEmail("henry-two@opspilot.dev")
		if err != nil {
			t.Fatalf("load second Henry: %v", err)
		}
		if *userOne.OrganizationID == *userTwo.OrganizationID {
			t.Fatalf("expected two distinct organizations even with the same requested name")
		}

		orgOne, err := organizationRepo.GetByID(*userOne.OrganizationID)
		if err != nil {
			t.Fatalf("load first Henry's organization: %v", err)
		}
		orgTwo, err := organizationRepo.GetByID(*userTwo.OrganizationID)
		if err != nil {
			t.Fatalf("load second Henry's organization: %v", err)
		}
		if orgOne.Slug == orgTwo.Slug {
			t.Fatalf("expected distinct slugs, got %q for both", orgOne.Slug)
		}
		if orgOne.Name != "Shared Name" || orgTwo.Name != "Shared Name" {
			t.Fatalf("expected both organizations to keep the requested display name")
		}
	})

	t.Run("Email uniqueness behavior is unchanged", func(t *testing.T) {
		err := userService.Register("Alice Duplicate", "alice-first@opspilot.dev", "password123", "")
		if !errors.Is(err, apperrors.ErrEmailAlreadyExists) {
			t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})
}

func setupSQLiteInvitationAwareRegistrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:invitation_aware_registration_%s?mode=memory&cache=private", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}

	migrateIntegrationSchema(t, db)

	return db
}
