package notification

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// SlackProvider posts to a Slack "Incoming Webhook" URL - the simplest
// Slack integration available with no OAuth app/bot token required, which
// is why it is the one implemented here ("only implement providers that can
// be implemented cleanly with the existing architecture").
type SlackProvider struct {
	client *http.Client
}

func NewSlackProvider() *SlackProvider {
	return &SlackProvider{client: newHTTPClient()}
}

func (p *SlackProvider) Send(ctx context.Context, target string, msg Message) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("slack webhook url is required")
	}

	payload := map[string]any{"text": formatSlackText(msg)}
	return postJSON(ctx, p.client, target, payload)
}

func formatSlackText(msg Message) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%s*\n%s", msg.Title, msg.Body)
	for _, line := range sortedFieldLines(msg.Fields) {
		b.WriteString("\n> ")
		b.WriteString(line)
	}
	return b.String()
}
