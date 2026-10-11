package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/sp3640/opspilot/backend/internal/notification"
)

type recordingProvider struct {
	target string
	msg    notification.Message
	err    error
	calls  int
}

func (p *recordingProvider) Send(_ context.Context, target string, msg notification.Message) error {
	p.calls++
	p.target, p.msg = target, msg
	return p.err
}

func TestVerificationEmailContent(t *testing.T) {
	provider := &recordingProvider{}
	service := NewEmailService(provider, "https://app.example.com/")

	raw, _, err := generateVerificationToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if err := service.SendVerificationEmail(t.Context(), "user@example.com", raw, 24*time.Hour); err != nil {
		t.Fatalf("send: %v", err)
	}

	if provider.target != "user@example.com" {
		t.Fatalf("sent to %q", provider.target)
	}
	if provider.msg.Title != "Verify your OpsPilot email" {
		t.Fatalf("subject = %q", provider.msg.Title)
	}
	wantLink := "https://app.example.com/verify-email?token=" + raw
	for name, body := range map[string]string{"text": provider.msg.Body, "html": provider.msg.HTML} {
		if !strings.Contains(body, wantLink) {
			t.Errorf("%s body does not contain the link built from the frontend base URL: %q", name, body)
		}
		if !strings.Contains(body, "24 hours") {
			t.Errorf("%s body does not state the 24 hour expiry", name)
		}
		if !strings.Contains(body, "safely ignore") {
			t.Errorf("%s body lacks the 'if you did not create this account' line", name)
		}
		if strings.Contains(strings.ToLower(body), "localhost") || strings.Contains(strings.ToLower(body), "password") {
			t.Errorf("%s body mentions localhost or a password: %q", name, body)
		}
	}
	if !strings.Contains(provider.msg.HTML, ">Verify Email</a>") {
		t.Errorf("html body has no Verify Email button")
	}
}

func TestVerificationURLDependsOnConfiguredBaseOnly(t *testing.T) {
	cases := map[string]string{
		"https://app.example.com":       "https://app.example.com/verify-email?token=abc",
		"https://app.example.com/":      "https://app.example.com/verify-email?token=abc",
		"http://localhost:3000":         "http://localhost:3000/verify-email?token=abc",
		"https://example.com/opspilot/": "https://example.com/opspilot/verify-email?token=abc",
	}
	for base, want := range cases {
		if got := NewEmailService(&recordingProvider{}, base).VerificationURL("abc"); got != want {
			t.Errorf("base %q -> %q, want %q", base, got, want)
		}
	}
}

func TestVerificationEmailHTMLEscapesTheLink(t *testing.T) {
	provider := &recordingProvider{}
	service := NewEmailService(provider, `https://app.example.com"><script>alert(1)</script>`)
	if err := service.SendVerificationEmail(t.Context(), "u@example.com", "tok", time.Hour); err != nil {
		t.Fatalf("send: %v", err)
	}
	if strings.Contains(provider.msg.HTML, "<script>") {
		t.Fatalf("html contains an unescaped tag: %q", provider.msg.HTML)
	}
}

func TestSendVerificationEmailFailureNeverContainsTheToken(t *testing.T) {
	provider := &recordingProvider{err: errors.New("smtp: 550 mailbox unavailable")}
	service := NewEmailService(provider, "https://app.example.com")

	raw, _, _ := generateVerificationToken()
	err := service.SendVerificationEmail(t.Context(), "u@example.com", raw, time.Hour)
	if !errors.Is(err, ErrVerificationEmailNotSent) {
		t.Fatalf("error = %v, want ErrVerificationEmailNotSent", err)
	}
	if strings.Contains(err.Error(), raw) {
		t.Fatalf("the returned error contains the raw token: %v", err)
	}
}

func TestEmailServiceAvailability(t *testing.T) {
	if (*EmailService)(nil).Available() {
		t.Errorf("a nil service must not be available")
	}
	if !NewEmailService(&recordingProvider{}, "https://x").Available() {
		t.Errorf("a provider with no Configured() method is assumed available")
	}
	if NewEmailService(notification.NewEmailProvider(notification.EmailConfig{}), "https://x").Available() {
		t.Errorf("an SMTP provider without a host must not be available")
	}
	if !NewEmailService(notification.NewEmailProvider(notification.EmailConfig{Host: "smtp.example.com"}), "https://x").Available() {
		t.Errorf("an SMTP provider with a host must be available")
	}
	unconfigured := NewEmailService(notification.NewEmailProvider(notification.EmailConfig{}), "https://x")
	err := unconfigured.SendVerificationEmail(t.Context(), "u@example.com", "tok", time.Hour)
	if !errors.Is(err, ErrVerificationEmailNotSent) || !errors.Is(err, notification.ErrEmailNotConfigured) {
		t.Fatalf("sending without SMTP = %v, want both ErrVerificationEmailNotSent and ErrEmailNotConfigured", err)
	}
}

func TestDescribeDuration(t *testing.T) {
	cases := map[time.Duration]string{
		24 * time.Hour:   "24 hours",
		time.Hour:        "1 hour",
		90 * time.Minute: "90 minutes",
		time.Minute:      "1 minute",
		10 * time.Second: "1 minute",
	}
	for d, want := range cases {
		if got := describeDuration(d); got != want {
			t.Errorf("describeDuration(%s) = %q, want %q", d, got, want)
		}
	}
}

func TestVerificationTokens(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		raw, hash, err := generateVerificationToken()
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if !validVerificationTokenFormat(raw) {
			t.Fatalf("a freshly issued token fails its own format check: %q", raw)
		}
		if len(raw) != 43 {
			t.Fatalf("token length = %d, want 43 (32 bytes base64url)", len(raw))
		}
		if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(hash) || hash == raw {
			t.Fatalf("stored value is not a distinct SHA-256 hex digest: %q", hash)
		}
		if hashVerificationToken(raw) != hash {
			t.Fatalf("hashing is not deterministic")
		}
		if seen[raw] {
			t.Fatalf("token collision after %d tokens", i)
		}
		seen[raw] = true
	}

	for _, bad := range []string{"", "short", strings.Repeat("a", 42), strings.Repeat("a", 44), strings.Repeat("+", 43), strings.Repeat("a", 42) + "=", " " + strings.Repeat("a", 42)} {
		if validVerificationTokenFormat(bad) {
			t.Errorf("format check accepted %q", bad)
		}
	}
}
