package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/auth"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo             *repository.UserRepository
	organizationRepo *repository.OrganizationRepository
	invitationRepo   *repository.InvitationRepository
	cfg              *config.Config
}

func NewUserService(
	repo *repository.UserRepository,
	organizationRepo *repository.OrganizationRepository,
	invitationRepo *repository.InvitationRepository,
	cfg *config.Config,
) *UserService {
	return &UserService{
		repo:             repo,
		organizationRepo: organizationRepo,
		invitationRepo:   invitationRepo,
		cfg:              cfg,
	}
}

// Register creates a new user through exactly one of two explicit paths:
//   - a pending invitation matching the email exists: the user joins that
//     invitation's organization with its role (invitation-aware registration).
//   - otherwise: the user creates their own new workspace and becomes its
//     Platform Admin. This applies uniformly to the very first user ever
//     registered and to every later uninvited signup — there is no hidden
//     "attach to an existing organization" fallback (previously GetFirst()).
func (s *UserService) Register(name, email, password, organizationName string) error {

	// Normalize user input
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))

	_, err := s.repo.GetByEmail(email)

	if err == nil {
		return apperrors.ErrEmailAlreadyExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	invitation, err := s.findValidPendingInvitation(email)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}

	if invitation != nil {
		user.Role = invitation.Role
	} else {
		user.Role = models.RolePlatformAdmin
	}

	if err := s.repo.Create(user); err != nil {
		return err
	}

	if invitation != nil {
		return s.repo.AssignOrganization(user.ID, invitation.OrganizationID)
	}

	organization, err := s.createWorkspace(user, organizationName)
	if err != nil {
		return err
	}

	return s.repo.AssignOrganization(user.ID, organization.ID)
}

// createWorkspace creates the new organization an uninvited registrant
// becomes Platform Admin of. The slug is suffixed with the user's own id,
// which is only known after user creation, so that two registrants choosing
// (or defaulting to) the same workspace name never collide on the
// organization's unique slug.
func (s *UserService) createWorkspace(user *models.User, organizationName string) (*models.Organization, error) {
	name := strings.TrimSpace(organizationName)
	if name == "" {
		name = fmt.Sprintf("%s's Workspace", user.Name)
	}

	suffix := fmt.Sprintf("-%d", user.ID)
	base := utils.GenerateSlug(name)
	if maxBaseLen := 120 - len(suffix); len(base) > maxBaseLen {
		base = base[:maxBaseLen]
	}

	organization := &models.Organization{
		Name:        name,
		Slug:        base + suffix,
		Description: "",
		OwnerID:     user.ID,
	}

	if err := s.organizationRepo.Create(organization); err != nil {
		return nil, err
	}

	return organization, nil
}

// findValidPendingInvitation looks up a still-valid pending invitation for
// email, reusing the same expiry semantics InvitationService already applies
// (ExpireOldInvitations followed by an ExpiresAt check). It returns (nil, nil)
// when no usable invitation exists, so registration falls back to the
// existing first-organization behavior.
func (s *UserService) findValidPendingInvitation(email string) (*models.Invitation, error) {
	if _, err := s.invitationRepo.ExpireOldInvitations(time.Now().UTC()); err != nil {
		return nil, err
	}

	invitation, err := s.invitationRepo.GetPendingByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if invitation.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil
	}

	return invitation, nil
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(email, password string) (string, error) {

	// Normalize email before searching
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(
		user.ID,
		user.Email,
		resolveUserRole(user.Role),
		resolveOrganizationID(user.OrganizationID),
		s.cfg.JWTSecret,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}
func (s *UserService) GetCurrentUser(id uint) (*models.User, error) {

	user, err := s.repo.GetByIDWithOrganization(id)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	return user, nil
}

// ReissueToken mints a fresh access token for userID from their CURRENT
// database role and organization. It exists so a session started before a
// role/organization change — most commonly, accepting an invitation — can
// be brought up to date without a full re-login. It deliberately reuses the
// same claims/signing path as Login rather than introducing a second,
// longer-lived token type.
func (s *UserService) ReissueToken(userID uint) (string, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperrors.ErrUserNotFound
		}
		return "", err
	}

	return auth.GenerateToken(
		user.ID,
		user.Email,
		resolveUserRole(user.Role),
		resolveOrganizationID(user.OrganizationID),
		s.cfg.JWTSecret,
	)
}

func resolveUserRole(role string) string {
	if strings.TrimSpace(role) == "" {
		return models.RoleViewer
	}

	return role
}

func resolveOrganizationID(organizationID *uuid.UUID) string {
	if organizationID == nil {
		return ""
	}

	return organizationID.String()
}
