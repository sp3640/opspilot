package notification

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// sendTimeout matches the timeout already used for the Slack/Teams/Webhook
// providers (internal/notification/http.go), so email doesn't stand out as
// the one channel that can block a caller indefinitely.
const sendTimeout = 10 * time.Second

// EmailConfig is sourced entirely from environment configuration
// (config.Config) - never from a database row - so no SMTP credential is
// ever stored, logged, or exposed through the API. Host empty means email
// notifications are simply not configured; Send reports that plainly rather
// than failing in a confusing way deeper in the stack.
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type EmailProvider struct {
	cfg EmailConfig
}

func NewEmailProvider(cfg EmailConfig) *EmailProvider {
	return &EmailProvider{cfg: cfg}
}

// Send is called synchronously from the alert-creation request path
// (AlertService.CreateAlert -> NotificationService.Dispatch), so it must
// never block that request indefinitely. net/smtp.SendMail accepts no
// context and has no dial/read timeout of its own, so it's run in a
// goroutine and raced against ctx here. Note this bounds how long the
// *caller* waits, not the underlying TCP connection itself - net/smtp gives
// no way to cancel a call already in flight, so a truly wedged SMTP server
// still leaves that one goroutine running until the OS-level TCP timeout
// eventually fires. That's an acceptable, well-isolated leak (one goroutine,
// no shared state) compared to blocking the request that triggered it.
func (p *EmailProvider) Send(ctx context.Context, target string, msg Message) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return fmt.Errorf("email target address is required")
	}
	if strings.TrimSpace(p.cfg.Host) == "" {
		return fmt.Errorf("email notifications are not configured (SMTP_HOST is unset)")
	}

	addr := net.JoinHostPort(p.cfg.Host, p.cfg.Port)

	var auth smtp.Auth
	if p.cfg.Username != "" {
		auth = smtp.PlainAuth("", p.cfg.Username, p.cfg.Password, p.cfg.Host)
	}

	sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	result := make(chan error, 1)
	go func() {
		result <- smtp.SendMail(addr, auth, p.cfg.From, []string{target}, buildEmailMessage(p.cfg.From, target, msg))
	}()

	select {
	case err := <-result:
		return err
	case <-sendCtx.Done():
		return fmt.Errorf("send email to %s: %w", target, sendCtx.Err())
	}
}

// buildEmailMessage produces a minimal, valid RFC 5322 message (headers,
// blank line, plain-text body) - no MIME multipart/attachments are needed
// for a short operational alert.
func buildEmailMessage(from, to string, msg Message) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", msg.Title)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(plainTextBody(msg))
	b.WriteString("\r\n")
	return []byte(b.String())
}
