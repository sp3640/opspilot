package services

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"github.com/sp3640/opspilot/backend/internal/notification"
)

// ErrVerificationEmailNotSent marks any failure to hand the verification
// email to the provider (SMTP down, rejected recipient, timeout, or email
// not configured). Callers use it to tell "could not send" from a database
// failure without matching on message text.
var ErrVerificationEmailNotSent = errors.New("verification email could not be sent")

// EmailService is the application-level way to send account emails
// (currently: email verification). It owns the wording and the links; the
// wire transport is whatever notification.Provider it is given - the same
// SMTP-backed EmailProvider the alert notifications use, so there is one
// email path and one SMTP configuration. Swapping providers means passing a
// different notification.Provider; nothing in the auth flow changes.
type EmailService struct {
	provider        notification.Provider
	frontendBaseURL string
}

func NewEmailService(provider notification.Provider, frontendBaseURL string) *EmailService {
	return &EmailService{provider: provider, frontendBaseURL: frontendBaseURL}
}

// Available reports whether email can be sent at all. It is a configuration
// check (is a provider set up?), never a per-recipient answer, so callers
// can use it without revealing anything about a particular address.
func (s *EmailService) Available() bool {
	if s == nil || s.provider == nil {
		return false
	}
	if configurable, ok := s.provider.(interface{ Configured() bool }); ok {
		return configurable.Configured()
	}
	return true
}

// VerificationURL is the link placed in the verification email: the
// frontend's verify page carrying the raw token. It is built from the
// configured frontend base URL (FRONTEND_BASE_URL) and is never logged.
func (s *EmailService) VerificationURL(rawToken string) string {
	base := strings.TrimRight(strings.TrimSpace(s.frontendBaseURL), "/")
	return base + "/verify-email?token=" + url.QueryEscape(rawToken)
}

// SendVerificationEmail mails the verification link to toEmail. The raw
// token appears only in the message handed to the provider; it is not
// logged and not part of any returned error.
func (s *EmailService) SendVerificationEmail(ctx context.Context, toEmail, rawToken string, ttl time.Duration) error {
	if !s.Available() {
		return fmt.Errorf("%w: %w", ErrVerificationEmailNotSent, notification.ErrEmailNotConfigured)
	}

	link := s.VerificationURL(rawToken)
	expires := describeDuration(ttl)

	text := "Welcome to OpsPilot.\n\n" +
		"Please verify your email address to activate your account:\n\n" +
		link + "\n\n" +
		"This verification link expires in " + expires + ".\n\n" +
		"If you did not create this account, you can safely ignore this email."

	page := "<!doctype html><html><body style=\"margin:0;padding:24px;background:#f4f6fa;font-family:Arial,Helvetica,sans-serif;color:#0f172a\">" +
		"<table role=\"presentation\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\"><tr><td align=\"center\">" +
		"<table role=\"presentation\" width=\"480\" cellpadding=\"0\" cellspacing=\"0\" style=\"max-width:480px;background:#ffffff;border-radius:12px;padding:32px\">" +
		"<tr><td>" +
		"<h1 style=\"margin:0 0 16px;font-size:22px\">Welcome to OpsPilot</h1>" +
		"<p style=\"margin:0 0 24px;font-size:15px;line-height:1.5\">Please verify your email address to activate your account.</p>" +
		"<p style=\"margin:0 0 24px\"><a href=\"" + html.EscapeString(link) + "\" style=\"display:inline-block;background:#3b82f6;color:#ffffff;text-decoration:none;font-weight:600;padding:12px 24px;border-radius:8px\">Verify Email</a></p>" +
		"<p style=\"margin:0 0 8px;font-size:13px;color:#475569\">This verification link expires in " + html.EscapeString(expires) + ".</p>" +
		"<p style=\"margin:0 0 16px;font-size:13px;color:#475569\">If the button does not work, copy and paste this link into your browser:<br><span style=\"word-break:break-all\">" + html.EscapeString(link) + "</span></p>" +
		"<p style=\"margin:0;font-size:13px;color:#475569\">If you did not create this account, you can safely ignore this email.</p>" +
		"</td></tr></table></td></tr></table></body></html>"

	err := s.provider.Send(ctx, toEmail, notification.Message{
		Title: "Verify your OpsPilot email",
		Body:  text,
		HTML:  page,
	})
	if err != nil {
		// The provider error describes the SMTP failure; wrapping adds
		// context without ever including the message (and thus the link).
		return fmt.Errorf("%w: %w", ErrVerificationEmailNotSent, err)
	}
	return nil
}

// describeDuration renders a token lifetime for humans: "24 hours",
// "1 hour", "90 minutes".
func describeDuration(d time.Duration) string {
	if d >= time.Hour && d%time.Hour == 0 {
		hours := int(d / time.Hour)
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	minutes := int((d + time.Minute - 1) / time.Minute)
	if minutes <= 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}
