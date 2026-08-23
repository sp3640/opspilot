package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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
	auditService     *AuditService
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

func (s *UserService) WithAuditService(auditService *AuditService) *UserService {
	s.auditService = auditService
	return s
}

// Register creates a new user who always creates their own new workspace
// and becomes its Platform Admin - uniformly for the very first user ever
// registered and for every later signup. There is no "attach to an existing
// organization" fallback here, even when a pending invitation exists for the
// email: registration alone never proves the registrant controls that email
// address, so it must never be sufficient to grant membership in someone
// else's organization. Joining an invited organization is exclusively the
// job of the token-verified InvitationService.AcceptInvitation flow, which
// requires both the invitation's token and a matching authenticated email.
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

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RolePlatformAdmin,
	}

	if err := s.repo.Create(user); err != nil {
		return err
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

// Login authenticates a user and returns a JWT token. ipAddress/userAgent
// are captured on a best-effort basis from the originating request purely
// for the audit trail (Phase 23) - an empty value is recorded as-is, never
// fabricated.
func (s *UserService) Login(email, password, ipAddress, userAgent string) (string, error) {

	// Normalize email before searching
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		// No user record exists for this email, so there is no organization
		// to attach an audit entry to (AuditLog.OrganizationID is required,
		// and inventing one would violate organization isolation) - this
		// attempt is therefore not audited, unlike a failed attempt against
		// a real account below.
		return "", apperrors.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		s.logLoginAttempt(user, models.AuditResultFailure, ipAddress, userAgent)
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

	s.logLoginAttempt(user, models.AuditResultSuccess, ipAddress, userAgent)

	return token, nil
}

// logLoginAttempt records a login success/failure for a KNOWN user account
// (one that exists and therefore has a real organization to scope the entry
// to). Best-effort: a write failure never fails the login itself.
func (s *UserService) logLoginAttempt(user *models.User, result models.AuditResult, ipAddress, userAgent string) {
	if s.auditService == nil || user.OrganizationID == nil {
		return
	}

	_ = s.auditService.LogEvent(AuditEventInput{
		UserID:         user.ID,
		OrganizationID: *user.OrganizationID,
		EntityType:     "auth",
		EntityID:       strconv.FormatUint(uint64(user.ID), 10),
		Action:         models.AuditActionLogin,
		Result:         result,
		FieldName:      "email",
		NewValue:       user.Email,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	})
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
