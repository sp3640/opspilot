package integration

import (
	"context"
	"net/http"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/notification"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/services"
)

// Account registration now requires following an emailed link. The
// integration tests keep the real registration, real token generation and
// hashing, real database, and real /auth/verify-email handler; the only
// thing they replace is the SMTP transport, with an in-memory mailbox that
// records what would have been sent. A test "clicks the link" by reading
// the raw token out of that recorded email - the same way a person would.

type sentEmail struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// testMailbox is a notification.Provider that records messages instead of
// sending them. Each test app has its own, so parallel tests that reuse an
// address never see each other's mail.
type testMailbox struct {
	mu       sync.Mutex
	messages []sentEmail

	// sendErr, when set, makes every Send fail (an SMTP outage);
	// notConfigured makes Configured report false (no SMTP_HOST).
	sendErr       error
	notConfigured bool
}

func (m *testMailbox) Send(_ context.Context, target string, msg notification.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.messages = append(m.messages, sentEmail{To: target, Subject: msg.Title, Text: msg.Body, HTML: msg.HTML})
	return nil
}

func (m *testMailbox) Configured() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.notConfigured
}

func (m *testMailbox) setSendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendErr = err
}

var verificationLinkPattern = regexp.MustCompile(`/verify-email\?token=([A-Za-z0-9_-]{43})`)

// latestToken is the raw token in the most recent email sent to address.
func (m *testMailbox) latestToken(t testing.TB, address string) string {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].To != address {
			continue
		}
		match := verificationLinkPattern.FindStringSubmatch(m.messages[i].Text)
		if match == nil {
			t.Fatalf("email to %s has no verification link: %q", address, m.messages[i].Text)
		}
		return match[1]
	}
	t.Fatalf("no email was sent to %s", address)
	return ""
}

func (m *testMailbox) sentTo(address string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, message := range m.messages {
		if message.To == address {
			count++
		}
	}
	return count
}

func (m *testMailbox) total() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.messages)
}

// testMailboxes maps an app's router to its mailbox so the shared HTTP
// helpers (registerAndLogin) can find the right one from the router alone.
var testMailboxes sync.Map // *gin.Engine -> *testMailbox

func attachMailbox(r *gin.Engine, mailbox *testMailbox) {
	testMailboxes.Store(r, mailbox)
}

func mailboxFor(t testing.TB, r *gin.Engine) *testMailbox {
	t.Helper()
	mailbox, ok := testMailboxes.Load(r)
	if !ok {
		t.Fatalf("no test mailbox is attached to this router; call attachMailbox in the app setup")
	}
	return mailbox.(*testMailbox)
}

// newTestEmailVerification builds the real verification service on db with
// the recording mailbox as its transport.
func newTestEmailVerification(db *gorm.DB, userRepo *repository.UserRepository, auditService *services.AuditService) (*services.EmailVerificationService, *testMailbox) {
	mailbox := &testMailbox{}
	service := services.NewEmailVerificationService(
		repository.NewEmailVerificationTokenRepository(db),
		userRepo,
		services.NewEmailService(mailbox, "https://app.opspilot.test"),
		24*time.Hour,
	)
	if auditService != nil {
		service.WithAuditService(auditService)
	}
	return service, mailbox
}

// registerVerified registers through the service and follows the emailed
// link, leaving a verified account that can log in.
func registerVerified(t testing.TB, userService *services.UserService, verification *services.EmailVerificationService, mailbox *testMailbox, name, email, password, organizationName string) {
	t.Helper()
	if err := userService.Register(name, email, password, organizationName); err != nil {
		t.Fatalf("register %s: %v", email, err)
	}
	if _, err := verification.Verify(mailbox.latestToken(t, email), "", ""); err != nil {
		t.Fatalf("verify %s: %v", email, err)
	}
}

// verifyEmailViaHTTP follows the emailed link through the real
// /auth/verify-email endpoint for an account registered over HTTP.
func verifyEmailViaHTTP(t *testing.T, r *gin.Engine, email string) {
	t.Helper()
	rec := doJSONRequest(t, r, http.MethodPost, "/api/v1/auth/verify-email", "", map[string]any{
		"token": mailboxFor(t, r).latestToken(t, email),
	})
	assertStatus(t, rec, http.StatusOK)
}
