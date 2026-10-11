package integration

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/config"
	"github.com/sp3640/opspilot/backend/internal/database"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/services"
)

// These tests exercise the real email verification behaviour end to end at
// the service layer: real UserService/EmailVerificationService, real
// tables, real token hashing and expiry. Only the SMTP transport is a
// recording mailbox. Every test runs twice - against in-memory SQLite, and
// against a genuine PostgreSQL schema when OPSPILOT_TEST_POSTGRES_DSN is
// set (row locking, constraints and timestamp handling differ there).

type verificationHarness struct {
	db           *gorm.DB
	users        *repository.UserRepository
	userService  *services.UserService
	verification *services.EmailVerificationService
	mailbox      *testMailbox
	auditRepo    *repository.AuditRepository
}

func newVerificationHarness(t *testing.T, db *gorm.DB) *verificationHarness {
	t.Helper()

	users := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)
	verification, mailbox := newTestEmailVerification(db, users, auditService)
	userService := services.NewUserService(
		users,
		repository.NewOrganizationRepository(db),
		repository.NewInvitationRepository(db),
		&config.Config{JWTSecret: "this-is-a-very-long-test-jwt-secret-1234567890"},
	).WithAuditService(auditService).WithEmailVerification(verification)

	return &verificationHarness{db: db, users: users, userService: userService, verification: verification, mailbox: mailbox, auditRepo: auditRepo}
}

func (h *verificationHarness) count(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := h.db.Raw(query, args...).Scan(&n).Error; err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return n
}

func (h *verificationHarness) user(t *testing.T, email string) *models.User {
	t.Helper()
	user, err := h.users.GetByEmail(email)
	if err != nil {
		t.Fatalf("load %s: %v", email, err)
	}
	return user
}

// ageTokens pretends the user's tokens were issued `by` ago, moving them
// past the resend cooldown (and, for large values, past expiry).
func (h *verificationHarness) ageTokens(t *testing.T, userID uint, by time.Duration) {
	t.Helper()
	if err := h.db.Exec(`UPDATE email_verification_tokens SET created_at = ? WHERE user_id = ?`, time.Now().Add(-by), userID).Error; err != nil {
		t.Fatalf("age tokens: %v", err)
	}
}

func (h *verificationHarness) expireTokens(t *testing.T, userID uint) {
	t.Helper()
	if err := h.db.Exec(`UPDATE email_verification_tokens SET expires_at = ? WHERE user_id = ?`, time.Now().Add(-time.Minute), userID).Error; err != nil {
		t.Fatalf("expire tokens: %v", err)
	}
}

func TestEmailVerificationSQLite(t *testing.T) {
	t.Parallel()
	runEmailVerificationSuite(t, func(t *testing.T) *gorm.DB {
		dsn := fmt.Sprintf("file:email_verification_%s?mode=memory&cache=private", uuid.NewString())
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("sqlite handle: %v", err)
		}
		sqlDB.SetMaxOpenConns(1)
		t.Cleanup(func() { _ = sqlDB.Close() })
		migrateIntegrationSchema(t, db)
		return db
	})
}

func TestEmailVerificationPostgres(t *testing.T) {
	t.Parallel()
	dsn := strings.TrimSpace(os.Getenv(postgresIntegrationDSNEnv))
	if dsn == "" {
		t.Skipf("skipping postgres email verification test: %s is not set", postgresIntegrationDSNEnv)
	}
	runEmailVerificationSuite(t, func(t *testing.T) *gorm.DB { return newIsolatedPostgresDB(t, dsn) })
}

// newIsolatedPostgresDB creates a uniquely named, initially empty schema,
// migrates the application tables into it and drops it afterwards, so
// concurrent tests (and the other Postgres integration test, which owns the
// public schema) never interfere.
func newIsolatedPostgresDB(t *testing.T, baseDSN string) *gorm.DB {
	t.Helper()

	admin, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}
	adminSQL, _ := admin.DB()

	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatalf("random schema name: %v", err)
	}
	name := "opspilot_ev_" + hex.EncodeToString(suffix)
	if err := admin.Exec(`CREATE SCHEMA ` + name).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	dsn := baseDSN + " search_path=" + name
	if strings.Contains(baseDSN, "://") {
		separator := "?"
		if strings.Contains(baseDSN, "?") {
			separator = "&"
		}
		dsn = baseDSN + separator + "search_path=" + name
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connect to schema %s: %v", name, err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() {
		_ = sqlDB.Close()
		_ = admin.Exec(`DROP SCHEMA ` + name + ` CASCADE`).Error
		_ = adminSQL.Close()
	})

	if err := database.MigrateUserOrganizationFoundation(db); err != nil {
		t.Fatalf("migrate user/organization foundation: %v", err)
	}
	migrateIntegrationSchema(t, db)
	return db
}

func runEmailVerificationSuite(t *testing.T, newDB func(t *testing.T) *gorm.DB) {
	t.Helper()

	const (
		email    = "new.user@opspilot.dev"
		password = "password123"
	)
	fresh := func(t *testing.T) *verificationHarness {
		t.Helper()
		return newVerificationHarness(t, newDB(t))
	}

	t.Run("registration creates an unverified account and stores only a hash of the token", func(t *testing.T) {
		h := fresh(t)
		if err := h.userService.Register("New User", email, password, ""); err != nil {
			t.Fatalf("register: %v", err)
		}

		user := h.user(t, email)
		if user.IsEmailVerified() {
			t.Fatalf("a newly registered account must be unverified")
		}
		if user.OrganizationID == nil {
			t.Fatalf("registration must still create the user's workspace")
		}

		raw := h.mailbox.latestToken(t, email)
		if len(raw) != 43 {
			t.Fatalf("raw token is %d chars, want 43 (256 bits, base64url)", len(raw))
		}
		if got := h.count(t, `SELECT count(*) FROM email_verification_tokens WHERE user_id = ?`, user.ID); got != 1 {
			t.Fatalf("expected exactly one stored token, have %d", got)
		}
		if got := h.count(t, `SELECT count(*) FROM email_verification_tokens WHERE token_hash = ?`, raw); got != 0 {
			t.Fatalf("the raw token is stored in plaintext")
		}
		var stored string
		if err := h.db.Raw(`SELECT token_hash FROM email_verification_tokens WHERE user_id = ?`, user.ID).Scan(&stored).Error; err != nil {
			t.Fatalf("read token hash: %v", err)
		}
		if len(stored) != 64 || strings.Contains(stored, raw) {
			t.Fatalf("stored value %q is not a SHA-256 hex digest distinct from the token", stored)
		}

		if got := h.mailbox.sentTo(email); got != 1 {
			t.Fatalf("expected one verification email, sent %d", got)
		}
		message := h.mailbox.messages[0]
		if message.Subject != "Verify your OpsPilot email" {
			t.Fatalf("unexpected subject %q", message.Subject)
		}
		if !strings.Contains(message.Text, "https://app.opspilot.test/verify-email?token="+raw) {
			t.Fatalf("email link is not built from the configured frontend URL: %q", message.Text)
		}
		for _, secret := range []string{password, "localhost"} {
			if strings.Contains(message.Text, secret) || strings.Contains(message.HTML, secret) {
				t.Fatalf("email contains %q", secret)
			}
		}
	})

	t.Run("an unverified account cannot log in, a verified one can", func(t *testing.T) {
		h := fresh(t)
		if err := h.userService.Register("New User", email, password, ""); err != nil {
			t.Fatalf("register: %v", err)
		}

		token, err := h.userService.Login(email, password, "", "")
		if !errors.Is(err, apperrors.ErrEmailNotVerified) || token != "" {
			t.Fatalf("unverified login = (%q, %v), want no token and ErrEmailNotVerified", token, err)
		}
		// A wrong password must still look like any other bad login.
		if _, err := h.userService.Login(email, "wrong-password", "", ""); !errors.Is(err, apperrors.ErrInvalidCredentials) {
			t.Fatalf("wrong password on an unverified account = %v, want ErrInvalidCredentials (no verification hint without the password)", err)
		}
		user := h.user(t, email)
		if _, err := h.userService.ReissueToken(user.ID); !errors.Is(err, apperrors.ErrEmailNotVerified) {
			t.Fatalf("an unverified account must not be able to reissue a token, got %v", err)
		}

		outcome, err := h.verification.Verify(h.mailbox.latestToken(t, email), "203.0.113.9", "test-agent")
		if err != nil || outcome != services.VerificationVerified {
			t.Fatalf("verify = (%q, %v), want verified", outcome, err)
		}
		token, err = h.userService.Login(email, password, "", "")
		if err != nil || token == "" {
			t.Fatalf("verified login failed: %v", err)
		}
		if !h.user(t, email).IsEmailVerified() {
			t.Fatalf("email_verified_at was not set")
		}
	})

	t.Run("a link is single use and repeating it is safe", func(t *testing.T) {
		h := fresh(t)
		if err := h.userService.Register("New User", email, password, ""); err != nil {
			t.Fatalf("register: %v", err)
		}
		raw := h.mailbox.latestToken(t, email)

		if _, err := h.verification.Verify(raw, "", ""); err != nil {
			t.Fatalf("first verify: %v", err)
		}
		stampedAt := *h.user(t, email).EmailVerifiedAt

		outcome, err := h.verification.Verify(raw, "", "")
		if err != nil || outcome != services.VerificationAlreadyVerified {
			t.Fatalf("second verify = (%q, %v), want already_verified", outcome, err)
		}
		if !h.user(t, email).EmailVerifiedAt.Equal(stampedAt) {
			t.Fatalf("re-using a link changed the verification timestamp")
		}
		if got := h.count(t, `SELECT count(*) FROM email_verification_tokens WHERE used_at IS NULL`); got != 0 {
			t.Fatalf("%d unused token(s) remain after verification", got)
		}
	})

	t.Run("expired and invalid tokens fail", func(t *testing.T) {
		h := fresh(t)
		if err := h.userService.Register("New User", email, password, ""); err != nil {
			t.Fatalf("register: %v", err)
		}
		user := h.user(t, email)
		raw := h.mailbox.latestToken(t, email)

		h.expireTokens(t, user.ID)
		if _, err := h.verification.Verify(raw, "", ""); !errors.Is(err, apperrors.ErrVerificationTokenExpired) {
			t.Fatalf("expired token = %v, want ErrVerificationTokenExpired", err)
		}
		if h.user(t, email).IsEmailVerified() {
			t.Fatalf("an expired token verified the account")
		}

		for name, bad := range map[string]string{
			"empty":                   "",
			"too short":               "abc",
			"outside the alphabet":    strings.Repeat("!", 43),
			"well formed but unknown": strings.Repeat("A", 43),
		} {
			if _, err := h.verification.Verify(bad, "", ""); !errors.Is(err, apperrors.ErrVerificationTokenInvalid) {
				t.Fatalf("%s token = %v, want ErrVerificationTokenInvalid", name, err)
			}
		}
		var storedHash string
		if err := h.db.Raw(`SELECT token_hash FROM email_verification_tokens WHERE user_id = ?`, user.ID).Scan(&storedHash).Error; err != nil {
			t.Fatalf("read hash: %v", err)
		}
		if _, err := h.verification.Verify(storedHash[:43], "", ""); !errors.Is(err, apperrors.ErrVerificationTokenInvalid) {
			t.Fatalf("presenting the stored hash as a token must not work, got %v", err)
		}
		if h.user(t, email).IsEmailVerified() {
			t.Fatalf("an invalid token verified the account")
		}
	})

	t.Run("resend replaces the previous token", func(t *testing.T) {
		h := fresh(t)
		if err := h.userService.Register("New User", email, password, ""); err != nil {
			t.Fatalf("register: %v", err)
		}
		user := h.user(t, email)
		first := h.mailbox.latestToken(t, email)

		// Inside the cooldown: silently nothing.
		if err := h.verification.Resend(t.Context(), email, "", ""); err != nil {
			t.Fatalf("resend inside cooldown: %v", err)
		}
		if got := h.mailbox.sentTo(email); got != 1 {
			t.Fatalf("resend inside the cooldown sent an email (total %d)", got)
		}

		h.ageTokens(t, user.ID, 5*time.Minute)
		if err := h.verification.Resend(t.Context(), email, "", ""); err != nil {
			t.Fatalf("resend: %v", err)
		}
		if got := h.mailbox.sentTo(email); got != 2 {
			t.Fatalf("expected a second email after the cooldown, total %d", got)
		}
		second := h.mailbox.latestToken(t, email)
		if second == first {
			t.Fatalf("resend reused the previous token")
		}
		if got := h.count(t, `SELECT count(*) FROM email_verification_tokens WHERE user_id = ? AND used_at IS NULL`, user.ID); got != 1 {
			t.Fatalf("resend left %d live tokens, want exactly 1", got)
		}

		if _, err := h.verification.Verify(first, "", ""); !errors.Is(err, apperrors.ErrVerificationTokenInvalid) {
			t.Fatalf("the replaced token still works: %v", err)
		}
		if _, err := h.verification.Verify(second, "", ""); err != nil {
			t.Fatalf("the new token failed: %v", err)
		}
	})

	t.Run("resend reveals nothing about unknown or verified addresses", func(t *testing.T) {
		h := fresh(t)
		registerVerified(t, h.userService, h.verification, h.mailbox, "Done User", "done@opspilot.dev", password, "")
		before := h.mailbox.total()

		for _, address := range []string{"nobody@opspilot.dev", "done@opspilot.dev"} {
			if err := h.verification.Resend(t.Context(), address, "", ""); err != nil {
				t.Fatalf("resend for %s returned %v; the outcome must not depend on the account", address, err)
			}
		}
		if got := h.mailbox.total(); got != before {
			t.Fatalf("resend sent %d email(s) for unknown/verified addresses", got-before)
		}
	})

	t.Run("a failed send rolls the whole registration back", func(t *testing.T) {
		h := fresh(t)
		h.mailbox.setSendError(errors.New("smtp: connection refused"))

		err := h.userService.Register("New User", email, password, "")
		if !errors.Is(err, apperrors.ErrEmailDeliveryUnavailable) {
			t.Fatalf("register with SMTP down = %v, want ErrEmailDeliveryUnavailable", err)
		}
		if got := h.count(t, `SELECT count(*) FROM users`); got != 0 {
			t.Fatalf("a failed send left %d user(s) behind", got)
		}
		if got := h.count(t, `SELECT count(*) FROM organizations`); got != 0 {
			t.Fatalf("a failed send left %d organization(s) behind", got)
		}
		if got := h.count(t, `SELECT count(*) FROM email_verification_tokens`); got != 0 {
			t.Fatalf("a failed send left %d token(s) behind", got)
		}

		// Once email works again the same address registers normally.
		h.mailbox.setSendError(nil)
		if err := h.userService.Register("New User", email, password, ""); err != nil {
			t.Fatalf("register after recovery: %v", err)
		}
	})

	t.Run("registration fails closed when email is not configured", func(t *testing.T) {
		h := fresh(t)
		h.mailbox.notConfigured = true

		if err := h.userService.Register("New User", email, password, ""); !errors.Is(err, apperrors.ErrEmailDeliveryUnavailable) {
			t.Fatalf("register without email = %v, want ErrEmailDeliveryUnavailable", err)
		}
		if got := h.count(t, `SELECT count(*) FROM users`); got != 0 {
			t.Fatalf("an account was created although no verification email can be sent")
		}
		if err := h.verification.Resend(t.Context(), email, "", ""); !errors.Is(err, apperrors.ErrEmailDeliveryUnavailable) {
			t.Fatalf("resend without email = %v, want ErrEmailDeliveryUnavailable", err)
		}
	})

	t.Run("registering someone else's unverified address cannot pre-set their password", func(t *testing.T) {
		h := fresh(t)
		// An attacker registers the victim's address first ...
		if err := h.userService.Register("Attacker", email, "attacker-password", ""); err != nil {
			t.Fatalf("attacker register: %v", err)
		}
		victim := h.user(t, email)
		h.ageTokens(t, victim.ID, 5*time.Minute)

		// ... then the real owner registers it, and verifies from their mailbox.
		if err := h.userService.Register("Real Owner", email, password, ""); err != nil {
			t.Fatalf("owner register: %v", err)
		}
		if got := h.count(t, `SELECT count(*) FROM users WHERE email = ?`, email); got != 1 {
			t.Fatalf("re-registration created a second account")
		}
		if _, err := h.verification.Verify(h.mailbox.latestToken(t, email), "", ""); err != nil {
			t.Fatalf("verify: %v", err)
		}

		if _, err := h.userService.Login(email, "attacker-password", "", ""); !errors.Is(err, apperrors.ErrInvalidCredentials) {
			t.Fatalf("the attacker's password works on the account: %v", err)
		}
		if _, err := h.userService.Login(email, password, "", ""); err != nil {
			t.Fatalf("the real owner cannot log in: %v", err)
		}
		if got := h.user(t, email).Name; got != "Real Owner" {
			t.Fatalf("account name = %q, want the latest registrant's", got)
		}
	})

	t.Run("registering a verified address alters nothing", func(t *testing.T) {
		h := fresh(t)
		registerVerified(t, h.userService, h.verification, h.mailbox, "Done User", email, password, "")
		before := h.user(t, email)
		emails := h.mailbox.total()

		if err := h.userService.Register("Mallory", email, "attacker-password", ""); err != nil {
			t.Fatalf("register existing: %v", err)
		}

		after := h.user(t, email)
		if after.PasswordHash != before.PasswordHash || after.Name != before.Name || !after.IsEmailVerified() {
			t.Fatalf("registering a verified address modified the account")
		}
		if h.mailbox.total() != emails {
			t.Fatalf("registering a verified address sent an email")
		}
	})

	t.Run("audit trail records the events and never the token", func(t *testing.T) {
		h := fresh(t)
		if err := h.userService.Register("New User", email, password, ""); err != nil {
			t.Fatalf("register: %v", err)
		}
		user := h.user(t, email)
		raw := h.mailbox.latestToken(t, email)
		var storedHash string
		if err := h.db.Raw(`SELECT token_hash FROM email_verification_tokens WHERE user_id = ?`, user.ID).Scan(&storedHash).Error; err != nil {
			t.Fatalf("read hash: %v", err)
		}

		h.ageTokens(t, user.ID, 5*time.Minute)
		if err := h.verification.Resend(t.Context(), email, "203.0.113.7", "agent"); err != nil {
			t.Fatalf("resend: %v", err)
		}
		if _, err := h.verification.Verify(h.mailbox.latestToken(t, email), "203.0.113.7", "agent"); err != nil {
			t.Fatalf("verify: %v", err)
		}

		var rows []models.AuditLog
		if err := h.db.Where("entity_type = ? AND organization_id = ?", "email_verification", *user.OrganizationID).Order("id ASC").Find(&rows).Error; err != nil {
			t.Fatalf("load audit rows: %v", err)
		}
		if len(rows) != 3 {
			t.Fatalf("expected sent, resent and verified audit rows, got %d", len(rows))
		}
		want := []struct {
			action       models.AuditAction
			field, value string
		}{
			{models.AuditActionCreate, "verification_email", "sent"},
			{models.AuditActionCreate, "verification_email", "resent"},
			{models.AuditActionUpdate, "email_verified_at", ""},
		}
		for i, row := range rows {
			if row.Action != want[i].action || row.FieldName != want[i].field || (want[i].value != "" && row.NewValue != want[i].value) {
				t.Errorf("audit row %d = %s/%s/%q, want %s/%s/%q", i, row.Action, row.FieldName, row.NewValue, want[i].action, want[i].field, want[i].value)
			}
			if row.UserID != user.ID || row.EntityID != fmt.Sprint(user.ID) {
				t.Errorf("audit row %d is attributed to the wrong user/entity", i)
			}
		}
		if rows[2].IPAddress != "203.0.113.7" {
			t.Errorf("verification audit row lost the request IP")
		}

		var everything []string
		if err := h.db.Raw(`SELECT COALESCE(entity_id,'') || COALESCE(field_name,'') || COALESCE(old_value,'') || COALESCE(new_value,'') || COALESCE(before_state,'') || COALESCE(after_state,'') || COALESCE(user_agent,'') FROM audit_logs`).Scan(&everything).Error; err != nil {
			t.Fatalf("scan audit text: %v", err)
		}
		for _, text := range everything {
			if strings.Contains(text, raw) || strings.Contains(text, storedHash) || strings.Contains(text, "verify-email") {
				t.Fatalf("an audit row contains token material: %q", text)
			}
		}
	})
}
