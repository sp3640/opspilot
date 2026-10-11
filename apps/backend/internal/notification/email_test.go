package notification

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestEmailProviderReportsUnconfiguredHost(t *testing.T) {
	provider := NewEmailProvider(EmailConfig{})
	err := provider.Send(context.Background(), "user@example.com", Message{Title: "x"})
	if err == nil {
		t.Fatalf("expected error when SMTP host is unset")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected a clear 'not configured' error, got %v", err)
	}
}

func TestEmailProviderRequiresTarget(t *testing.T) {
	provider := NewEmailProvider(EmailConfig{Host: "smtp.example.com"})
	if err := provider.Send(context.Background(), "", Message{Title: "x"}); err == nil {
		t.Fatalf("expected error for empty target")
	}
}

func TestBuildEmailMessageIncludesHeadersAndBody(t *testing.T) {
	msg := Message{Title: "Deployment failed", Body: "image pull error", Fields: map[string]string{"Environment": "production"}}
	raw := string(buildEmailMessage("ops@opspilot.dev", "oncall@opspilot.dev", msg))

	if !strings.Contains(raw, "From: ops@opspilot.dev\r\n") {
		t.Fatalf("expected From header, got %q", raw)
	}
	if !strings.Contains(raw, "To: oncall@opspilot.dev\r\n") {
		t.Fatalf("expected To header, got %q", raw)
	}
	if !strings.Contains(raw, "Subject: Deployment failed\r\n") {
		t.Fatalf("expected Subject header, got %q", raw)
	}
	if !strings.Contains(raw, "image pull error") || !strings.Contains(raw, "Environment: production") {
		t.Fatalf("expected body and fields in message, got %q", raw)
	}
}

func TestBuildEmailMessageWithHTMLIsMultipartAlternative(t *testing.T) {
	msg := Message{Title: "Hi", Body: "plain body", HTML: "<p>html body</p>"}
	raw := string(buildEmailMessage("OpsPilot <no-reply@mail.example.com>", "user@example.com", msg))

	if !strings.Contains(raw, "Content-Type: multipart/alternative; boundary=") {
		t.Fatalf("expected multipart/alternative, got %q", raw)
	}
	plainAt := strings.Index(raw, "plain body")
	htmlAt := strings.Index(raw, "<p>html body</p>")
	if plainAt < 0 || htmlAt < 0 || plainAt > htmlAt {
		t.Fatalf("expected plain part before html part (least to most preferred), got %q", raw)
	}
	if !strings.Contains(raw, "Message-ID: <") || !strings.Contains(raw, "@mail.example.com>") {
		t.Fatalf("expected Message-ID on the sender's domain, got %q", raw)
	}
	if !strings.Contains(raw, "\r\nDate: ") {
		t.Fatalf("expected Date header, got %q", raw)
	}
	if strings.Count(raw, "--opspilot-") != 3 {
		t.Fatalf("expected two parts and a closing boundary, got %q", raw)
	}
}

func TestBuildEmailMessageWithoutHTMLStaysPlainText(t *testing.T) {
	raw := string(buildEmailMessage("a@b.example", "c@d.example", Message{Title: "t", Body: "b"}))
	if !strings.Contains(raw, "Content-Type: text/plain; charset=\"UTF-8\"") || strings.Contains(raw, "multipart") {
		t.Fatalf("a message without HTML must remain a single plain-text part, got %q", raw)
	}
}

func TestBuildEmailMessageStripsHeaderInjection(t *testing.T) {
	raw := string(buildEmailMessage(
		"a@b.example",
		"victim@example.com\r\nBcc: attacker@example.com",
		Message{Title: "Hello\r\nBcc: attacker@example.com", Body: "b"},
	))
	headers := raw[:strings.Index(raw, "\r\n\r\n")]
	for _, line := range strings.Split(headers, "\r\n") {
		if strings.HasPrefix(line, "Bcc:") {
			t.Fatalf("a caller-supplied value injected a header: %q", headers)
		}
	}
}

func TestEmailProviderConfigured(t *testing.T) {
	if NewEmailProvider(EmailConfig{}).Configured() {
		t.Fatalf("provider with no host must not report configured")
	}
	if !NewEmailProvider(EmailConfig{Host: "smtp.example.com"}).Configured() {
		t.Fatalf("provider with a host must report configured")
	}
	if err := NewEmailProvider(EmailConfig{}).Send(context.Background(), "u@example.com", Message{}); !errors.Is(err, ErrEmailNotConfigured) {
		t.Fatalf("expected ErrEmailNotConfigured, got %v", err)
	}
}
