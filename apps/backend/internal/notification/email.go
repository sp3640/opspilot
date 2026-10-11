package notification

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// ErrEmailNotConfigured is returned by Send when no SMTP host is set, so
// callers can tell "email is not set up" apart from "the SMTP server failed".
var ErrEmailNotConfigured = errors.New("email notifications are not configured (SMTP_HOST is unset)")

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

// Configured reports whether an SMTP host is set, i.e. whether Send can
// possibly succeed. It says nothing about whether the server is reachable.
func (p *EmailProvider) Configured() bool {
	return strings.TrimSpace(p.cfg.Host) != ""
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
		return ErrEmailNotConfigured
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
		envelopeFrom, headerFrom := p.sender()
		result <- smtp.SendMail(addr, auth, envelopeFrom, []string{target}, buildEmailMessage(headerFrom, target, msg))
	}()

	select {
	case err := <-result:
		return err
	case <-sendCtx.Done():
		return fmt.Errorf("send email to %s: %w", target, sendCtx.Err())
	}
}

// sender splits SMTP_FROM into the two forms SMTP needs. The envelope sender
// (MAIL FROM) must be a bare address - servers reject "Name <addr>" there -
// while the From header may carry the display name. A value that does not
// parse as an address is used as given, exactly as before.
func (p *EmailProvider) sender() (envelope, header string) {
	from := strings.TrimSpace(p.cfg.From)
	if parsed, err := mail.ParseAddress(from); err == nil {
		return parsed.Address, parsed.String()
	}
	return from, from
}

// buildEmailMessage produces a minimal, valid RFC 5322 message: headers, a
// blank line, then a plain-text body - or, when msg.HTML is set, a
// multipart/alternative body carrying both renderings. No attachments are
// needed for a short operational or account email.
//
// Header values come from callers (subject, recipient), so CR/LF is
// stripped from each to make header injection impossible.
func buildEmailMessage(from, to string, msg Message) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", headerValue(from))
	fmt.Fprintf(&b, "To: %s\r\n", headerValue(to))
	fmt.Fprintf(&b, "Subject: %s\r\n", headerValue(msg.Title))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Message-ID: <%s@%s>\r\n", randomHex(16), messageIDDomain(from))
	b.WriteString("MIME-Version: 1.0\r\n")

	if msg.HTML == "" {
		b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		b.WriteString("\r\n")
		b.WriteString(plainTextBody(msg))
		b.WriteString("\r\n")
		return []byte(b.String())
	}

	boundary := "opspilot-" + randomHex(12)
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	b.WriteString("\r\n")
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s\r\n", boundary, plainTextBody(msg))
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s\r\n", boundary, msg.HTML)
	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return []byte(b.String())
}

// headerValue removes line breaks so a value can never terminate its header
// and smuggle in another one.
func headerValue(value string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(value)
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "0"
	}
	return hex.EncodeToString(buf)
}

// messageIDDomain is the domain of the sender address (falling back to a
// neutral host), which is what a Message-ID's right-hand side should be.
func messageIDDomain(from string) string {
	from = strings.TrimSpace(from)
	if open := strings.LastIndex(from, "<"); open >= 0 {
		from = strings.TrimSuffix(from[open+1:], ">")
	}
	if at := strings.LastIndex(from, "@"); at >= 0 && at < len(from)-1 {
		return strings.TrimSpace(from[at+1:])
	}
	return "opspilot.local"
}
