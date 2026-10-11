package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/auth"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/logger"
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
	verification     *EmailVerificationService
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

// WithEmailVerification enables account registration. Registration requires
// it: without a verification service Register fails closed rather than
// creating accounts that could never be verified (or, worse, silently
// skipping verification).
func (s *UserService) WithEmailVerification(verification *EmailVerificationService) *UserService {
	s.verification = verification
	return s
}

func (s *UserService) WithAuditService(auditService *AuditService) *UserService {
	s.auditService = auditService
	return s
}

// Register creates a new, UNVERIFIED user who owns a new workspace of their
// own (and is its Platform Admin), then emails a verification link. The
// account cannot log in until that link is followed.
//
// There is no "attach to an existing organization" fallback here, even when
// a pending invitation exists for the email: registration alone never
// proves the registrant controls that email address, so it must never be
// sufficient to grant membership in someone else's organization. Joining an
// invited organization is exclusively the job of the token-verified
// InvitationService.AcceptInvitation flow, which requires both the
// invitation's token and a matching authenticated email.
//
// The user, workspace and verification token are written in one
// transaction and the email is sent before it commits: if the email cannot
// be sent the whole registration rolls back (ErrEmailDeliveryUnavailable) and
// nothing half-created remains.
//
// The result does not depend on whether the email is already registered, so
// registration cannot be used to discover accounts:
//   - a verified account: nothing happens and no email is sent;
//   - an unverified account: its name/password are replaced by this
//     request's (so whoever registers last controls a not-yet-activated
//     account, and only the mailbox owner can activate it - this defeats
//     pre-registration of someone else's address) and a new link is sent,
//     subject to the resend cooldown.
//
// Residual, documented: response time differs slightly between a new
// address (SMTP send) and a known one (no send); the per-IP registration
// rate limit bounds how much that can be probed.
func (s *UserService) Register(name, email, password, organizationName string) error {
	if s.verification == nil || !s.verification.email.Available() {
		return apperrors.ErrEmailDeliveryUnavailable
	}

	// Normalize user input
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	ctx := context.Background()

	// Hash before looking the account up so both outcomes pay for bcrypt.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	existing, err := s.repo.GetByEmail(email)
	if err == nil {
		return s.registerExisting(ctx, existing, name, string(hash))
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         models.RolePlatformAdmin,
	}
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		users := s.repo.WithTx(tx)
		if err := users.Create(user); err != nil {
			return err
		}

		organization, err := s.createWorkspace(s.organizationRepo.WithTx(tx), user, organizationName)
		if err != nil {
			return err
		}
		if err := users.AssignOrganization(user.ID, organization.ID); err != nil {
			return err
		}
		user.OrganizationID = &organization.ID

		return s.verification.IssueAndSend(ctx, tx, user)
	})
	if err != nil {
		if isDuplicateUserEmailError(err) {
			// Lost a race with a concurrent registration of the same
			// address; indistinguishable from "already registered".
			return nil
		}
		if errors.Is(err, ErrVerificationEmailNotSent) {
			logger.Error(ctx, "registration rolled back: verification email could not be sent", slog.Any("error", err))
			return apperrors.ErrEmailDeliveryUnavailable
		}
		return err
	}

	s.verification.RecordSent(user, "sent")
	return nil
}

// registerExisting handles a registration for an address that already has
// an account, returning nil in every non-failure case (see Register).
func (s *UserService) registerExisting(ctx context.Context, existing *models.User, name, passwordHash string) error {
	if existing.IsEmailVerified() {
		return nil
	}

	var resent bool
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		users := s.repo.WithTx(tx)
		user, err := users.GetByIDForUpdate(existing.ID)
		if err != nil {
			return err
		}
		if user.IsEmailVerified() {
			return nil
		}
		if err := users.UpdateUnverifiedCredentials(user.ID, name, passwordHash); err != nil {
			return err
		}

		recent, err := s.verification.withinCooldown(tx, user.ID)
		if err != nil || recent {
			return err
		}
		if err := s.verification.IssueAndSend(ctx, tx, user); err != nil {
			return err
		}
		resent = true
		return nil
	})
	if err != nil {
		// Reporting a delivery failure here would reveal that this address
		// is registered and unverified; log it and answer like any other
		// registration. The transaction rolled back, so the user can retry
		// or use "resend".
		logger.Error(ctx, "re-registration of an unverified account failed", slog.Any("error", err))
		return nil
	}
	if resent {
		s.verification.RecordSent(existing, "resent")
	}
	return nil
}

// isDuplicateUserEmailError recognizes a unique-constraint violation on
// users.email from PostgreSQL (or SQLite in unit tests).
func isDuplicateUserEmailError(err error) bool {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError.Code == "23505" && pgError.ConstraintName == "idx_users_email"
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed: users.email")
}

// createWorkspace creates the new organization an uninvited registrant
// becomes Platform Admin of. The slug is suffixed with the user's own id,
// which is only known after user creation, so that two registrants choosing
// (or defaulting to) the same workspace name never collide on the
// organization's unique slug.
func (s *UserService) createWorkspace(organizationRepo *repository.OrganizationRepository, user *models.User, organizationName string) (*models.Organization, error) {
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

	if err := organizationRepo.Create(organization); err != nil {
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

	// Only after the password is proven: telling an unauthenticated caller
	// "this address is unverified" would reveal that it has an account.
	if !user.IsEmailVerified() {
		s.logLoginAttempt(user, models.AuditResultFailure, ipAddress, userAgent)
		return "", apperrors.ErrEmailNotVerified
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
	if !user.IsEmailVerified() {
		return "", apperrors.ErrEmailNotVerified
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
