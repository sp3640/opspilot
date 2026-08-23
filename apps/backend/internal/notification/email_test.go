package notification

import (
	"context"
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
