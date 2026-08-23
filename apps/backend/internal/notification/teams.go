package notification

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// TeamsProvider posts a legacy "Office 365 Connector" MessageCard to a
// Microsoft Teams Incoming Webhook URL - the simplest Teams integration
// available with no Azure AD app registration required.
type TeamsProvider struct {
	client *http.Client
}

func NewTeamsProvider() *TeamsProvider {
	return &TeamsProvider{client: newHTTPClient()}
}

func (p *TeamsProvider) Send(ctx context.Context, target string, msg Message) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("teams webhook url is required")
	}

	facts := make([]map[string]string, 0, len(msg.Fields))
	for _, line := range sortedFieldLines(msg.Fields) {
		name, value, found := strings.Cut(line, ": ")
		if !found {
			continue
		}
		facts = append(facts, map[string]string{"name": name, "value": value})
	}

	payload := map[string]any{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"summary":    msg.Title,
		"themeColor": themeColorForSeverity(msg.Severity),
		"title":      msg.Title,
		"text":       msg.Body,
		"sections": []map[string]any{
			{"facts": facts},
		},
	}
	return postJSON(ctx, p.client, target, payload)
}

func themeColorForSeverity(severity string) string {
	switch strings.ToUpper(severity) {
	case "CRITICAL", "P0":
		return "D93F3F"
	case "HIGH", "P1":
		return "E8A33D"
	default:
		return "5B8DEF"
	}
}
