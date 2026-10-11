package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/logger"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
)

const (
	// verificationTokenBytes of CSPRNG output = 256 bits of entropy, far
	// beyond guessable; the token is base64url encoded (43 characters).
	verificationTokenBytes  = 32
	verificationTokenLength = 43

	// resendCooldown is the minimum gap between two verification emails for
	// the same account. It is enforced from the token table, so it holds
	// across server instances and cannot be bypassed by changing IP.
	resendCooldown = 60 * time.Second

	auditEntityEmailVerification = "email_verification"
)

// VerificationOutcome is the successful result of following a verify link.
type VerificationOutcome string

const (
	// VerificationVerified: this request activated the account.
	VerificationVerified VerificationOutcome = "verified"
	// VerificationAlreadyVerified: the account was already active (the link
	// was used before). Nothing was changed.
	VerificationAlreadyVerified VerificationOutcome = "already_verified"
)

// EmailVerificationService issues, sends and redeems single-use email
// verification tokens.
//
// Token life cycle: a raw token is 256 random bits, mailed to the account's
// address and never stored - only its SHA-256 is. Redeeming hashes what was
// presented and looks that hash up, so the database cannot yield a working
// link. Tokens expire, are consumed on first use, and issuing a new one
// deletes the previous unused one.
type EmailVerificationService struct {
	tokens       *repository.EmailVerificationTokenRepository
	users        *repository.UserRepository
	email        *EmailService
	auditService *AuditService
	ttl          time.Duration
	now          func() time.Time
}

func NewEmailVerificationService(
	tokens *repository.EmailVerificationTokenRepository,
	users *repository.UserRepository,
	email *EmailService,
	ttl time.Duration,
) *EmailVerificationService {
	return &EmailVerificationService{
		tokens: tokens,
		users:  users,
		email:  email,
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *EmailVerificationService) WithAuditService(auditService *AuditService) *EmailVerificationService {
	s.auditService = auditService
	return s
}

// generateVerificationToken returns a fresh raw token and its storage hash.
func generateVerificationToken() (raw, hash string, err error) {
	buf := make([]byte, verificationTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate verification token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, hashVerificationToken(raw), nil
}

// hashVerificationToken is what the database stores and compares. SHA-256
// is appropriate (rather than a slow password hash) because the input is
// 256 bits of randomness, not a human-chosen secret.
func hashVerificationToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// validVerificationTokenFormat rejects anything that cannot be a token this
// service issued, before it is hashed or touches the database.
func validVerificationTokenFormat(raw string) bool {
	if len(raw) != verificationTokenLength {
		return false
	}
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

// IssueAndSend makes a new token the user's only unused one and mails it.
// It runs inside the caller's transaction, and the email is sent before the
// caller commits: if sending fails the error propagates, the transaction
// rolls back, and no token that nobody received is left behind (the
// previous token, if any, stays valid). The reverse - email sent but
// commit failing - leaves a link that simply never validates.
func (s *EmailVerificationService) IssueAndSend(ctx context.Context, tx *gorm.DB, user *models.User) error {
	raw, hash, err := generateVerificationToken()
	if err != nil {
		return err
	}

	token := &models.EmailVerificationToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: s.now().Add(s.ttl),
	}
	if err := s.tokens.WithTx(tx).ReplaceForUser(token); err != nil {
		return fmt.Errorf("store verification token: %w", err)
	}

	return s.email.SendVerificationEmail(ctx, user.Email, raw, s.ttl)
}

// withinCooldown reports whether a verification email was already issued to
// this user recently enough that another must not be sent yet.
func (s *EmailVerificationService) withinCooldown(tx *gorm.DB, userID uint) (bool, error) {
	last, found, err := s.tokens.WithTx(tx).LatestCreatedAt(userID)
	if err != nil {
		return false, err
	}
	return found && s.now().Sub(last) < resendCooldown, nil
}

// Verify redeems a raw token from a verification link.
//
// Errors: ErrVerificationTokenInvalid (malformed, unknown, or superseded by
// a newer link) and ErrVerificationTokenExpired. Following a link that was
// already used succeeds with VerificationAlreadyVerified and changes
// nothing, so verification is safe to repeat - including by mail scanners
// that pre-fetch links.
func (s *EmailVerificationService) Verify(rawToken, ipAddress, userAgent string) (VerificationOutcome, error) {
	rawToken = strings.TrimSpace(rawToken)
	if !validVerificationTokenFormat(rawToken) {
		return "", apperrors.ErrVerificationTokenInvalid
	}
	hash := hashVerificationToken(rawToken)

	var (
		outcome  VerificationOutcome
		verified *models.User
	)
	err := s.users.Transaction(func(tx *gorm.DB) error {
		tokens := s.tokens.WithTx(tx)
		users := s.users.WithTx(tx)

		token, err := tokens.GetByHashForUpdate(hash)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrVerificationTokenInvalid
		}
		if err != nil {
			return err
		}

		user, err := users.GetByIDForUpdate(token.UserID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrVerificationTokenInvalid
		}
		if err != nil {
			return err
		}

		if user.IsEmailVerified() {
			outcome = VerificationAlreadyVerified
			return nil
		}
		if token.UsedAt != nil {
			// A spent token whose account is not verified cannot happen
			// through this service; refuse rather than guess.
			return apperrors.ErrVerificationTokenInvalid
		}
		now := s.now()
		if !now.Before(token.ExpiresAt) {
			return apperrors.ErrVerificationTokenExpired
		}

		if err := users.MarkEmailVerified(user.ID, now); err != nil {
			return err
		}
		if err := tokens.MarkUsed(token.ID, now); err != nil {
			return err
		}
		if err := tokens.DeleteUnusedForUser(user.ID); err != nil {
			return err
		}

		user.EmailVerifiedAt = &now
		verified = user
		outcome = VerificationVerified
		return nil
	})
	if err != nil {
		return "", err
	}

	if verified != nil {
		s.audit(verified, "email_verified_at", "", verified.EmailVerifiedAt.UTC().Format(time.RFC3339), models.AuditActionUpdate, ipAddress, userAgent)
	}
	return outcome, nil
}

// Resend sends a fresh verification link to an unverified account.
//
// It is deliberately silent about the account: unknown emails, verified
// accounts, and requests inside the cooldown all return nil without doing
// anything, so the response cannot be used to learn who is registered. Only
// ErrEmailDeliveryUnavailable (a server-wide condition, independent of the
// address) is reported. A failure to deliver to one specific address is
// logged and swallowed for the same reason; the token stays unchanged so the
// user can simply try again.
func (s *EmailVerificationService) Resend(ctx context.Context, email, ipAddress, userAgent string) error {
	if !s.email.Available() {
		return apperrors.ErrEmailDeliveryUnavailable
	}
	email = strings.TrimSpace(strings.ToLower(email))

	var resent *models.User
	err := s.users.Transaction(func(tx *gorm.DB) error {
		user, err := s.users.WithTx(tx).GetByEmailForUpdate(email)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if user.IsEmailVerified() {
			return nil
		}

		recent, err := s.withinCooldown(tx, user.ID)
		if err != nil || recent {
			return err
		}
		if err := s.IssueAndSend(ctx, tx, user); err != nil {
			return err
		}
		resent = user
		return nil
	})
	if err != nil {
		logger.Error(ctx, "resend verification email failed", slog.Any("error", err))
		return nil
	}

	if resent != nil {
		s.audit(resent, "verification_email", "", "resent", models.AuditActionCreate, ipAddress, userAgent)
	}
	return nil
}

// RecordSent audits a verification email issued during registration.
func (s *EmailVerificationService) RecordSent(user *models.User, note string) {
	s.audit(user, "verification_email", "", note, models.AuditActionCreate, "", "")
}

// audit writes best-effort trail entries. Only the fact of the event is
// recorded - never the token, its hash, or the link.
func (s *EmailVerificationService) audit(user *models.User, field, oldValue, newValue string, action models.AuditAction, ipAddress, userAgent string) {
	if s.auditService == nil || user == nil || user.OrganizationID == nil {
		return
	}
	_ = s.auditService.LogEvent(AuditEventInput{
		UserID:         user.ID,
		OrganizationID: *user.OrganizationID,
		EntityType:     auditEntityEmailVerification,
		EntityID:       strconv.FormatUint(uint64(user.ID), 10),
		Action:         action,
		FieldName:      field,
		OldValue:       oldValue,
		NewValue:       newValue,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	})
}
