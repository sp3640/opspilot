package notification

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WebhookProvider posts a generic JSON envelope to an arbitrary URL - the
// integration point for any downstream system this backend doesn't have a
// dedicated provider for (a custom on-call tool, an internal automation
// endpoint, etc). Unlike Slack/Teams, the payload shape is this backend's
// own, not a third-party schema.
type WebhookProvider struct {
	client *http.Client
}

func NewWebhookProvider() *WebhookProvider {
	return &WebhookProvider{client: newHTTPClient()}
}

func (p *WebhookProvider) Send(ctx context.Context, target string, msg Message) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("webhook url is required")
	}

	payload := map[string]any{
		"event":     msg.Event,
		"title":     msg.Title,
		"body":      msg.Body,
		"severity":  msg.Severity,
		"fields":    msg.Fields,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return postJSON(ctx, p.client, target, payload)
}
