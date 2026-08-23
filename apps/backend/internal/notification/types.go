// Package notification implements the Phase 24 provider abstraction: a
// small Message envelope plus one Provider per channel type (Email, Slack,
// Teams, Webhook). Providers are deliberately dumb - they know how to turn
// a Message into a wire request for their transport and nothing else. They
// never touch the database, never decrypt a secret, and never decide
// whether a notification should be sent; NotificationService owns all of
// that. This keeps every provider trivially testable in isolation and easy
// to add to without touching the dispatch logic.
package notification

import "context"

// Message is the channel-agnostic content of one notification. Providers
// render it however fits their transport (plain text email body, Slack
// markdown, a Teams MessageCard, or a generic JSON envelope).
type Message struct {
	// Event is the constants.NotificationEvent* value that triggered this
	// message, included verbatim so a generic Webhook consumer can route on it.
	Event string
	// Title is a short one-line summary (e.g. an email subject / Slack bold line).
	Title string
	// Body is the human-readable detail.
	Body string
	// Severity is optional, informational context (e.g. an alert's severity)
	// providers may use for visual treatment (color-coding, etc).
	Severity string
	// Fields are structured key/value context (Application, Project,
	// Resource, ...) rendered alongside Title/Body.
	Fields map[string]string
}

// Provider sends one Message to one resolved target (a decrypted email
// address or webhook URL - NotificationService resolves this before calling
// Send, so a Provider never sees an encrypted value).
type Provider interface {
	Send(ctx context.Context, target string, msg Message) error
}
