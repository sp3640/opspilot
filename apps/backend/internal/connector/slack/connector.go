package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/sp3640/opspilot/backend/internal/connector"
	"github.com/sp3640/opspilot/backend/internal/notification"
)

const (
	webhookCredentialKey = "webhook_url"
	testTimeout          = 5 * time.Second
)

var (
	ErrWebhookRequired = errors.New("slack webhook url is required")
	ErrInvalidWebhook  = errors.New("slack webhook url is invalid")
)

type Connector struct {
	client *http.Client
}

func NewConnector() *Connector {
	return &Connector{client: &http.Client{Timeout: testTimeout}}
}

func (c *Connector) Connect(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	return c.TestConnection(ctx, cfg)
}

func (c *Connector) Disconnect(_ context.Context, _ connector.Config) error {
	return nil
}

func (c *Connector) TestConnection(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	webhook, err := extractWebhook(cfg)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{"text": "OpsPilot Slack connection test succeeded."}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode Slack test payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("build Slack webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send Slack test request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return &connector.Result{Success: true, Message: "Slack webhook is configured and responding successfully."}, nil
}

func (c *Connector) HealthCheck(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	return c.TestConnection(ctx, cfg)
}

func (c *Connector) Capabilities() connector.Capabilities {
	return connector.Capabilities{
		Implemented:            true,
		SupportsTestConnection: true,
		SupportsHealthCheck:    true,
		Description:            "Send operational alerts to Slack using an incoming webhook URL.",
	}
}

func (c *Connector) SendNotification(ctx context.Context, cfg connector.Config, msg notification.Message) error {
	webhook, err := extractWebhook(cfg)
	if err != nil {
		return err
	}

	payload := map[string]any{"text": renderSlackText(msg)}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode Slack notification payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("build Slack notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func extractWebhook(cfg connector.Config) (string, error) {
	for _, key := range []string{webhookCredentialKey, "webhookUrl", "url", "slack_webhook_url"} {
		if value := strings.TrimSpace(cfg.Credentials[key]); value != "" {
			return validateWebhook(value)
		}
		if cfg.Metadata != nil {
			if value := strings.TrimSpace(cfg.Metadata[key]); value != "" {
				return validateWebhook(value)
			}
		}
	}
	return "", ErrWebhookRequired
}

func validateWebhook(value string) (string, error) {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil {
		return "", ErrInvalidWebhook
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidWebhook
	}
	if parsed.Host == "" {
		return "", ErrInvalidWebhook
	}
	return parsed.String(), nil
}

func renderSlackText(msg notification.Message) string {
	var b strings.Builder
	if msg.Title != "" {
		b.WriteString("*")
		b.WriteString(msg.Title)
		b.WriteString("*")
	}
	if msg.Body != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(msg.Body)
	}
	keys := make([]string, 0, len(msg.Fields))
	for k := range msg.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		for _, key := range keys {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(key)
			b.WriteString(": ")
			b.WriteString(msg.Fields[key])
		}
	}
	return b.String()
}
