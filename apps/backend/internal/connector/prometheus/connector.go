package prometheus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sp3640/opspilot/backend/internal/connector"
	"github.com/sp3640/opspilot/backend/internal/notification"
)

const (
	defaultTimeout = 10 * time.Second
	queryParamName = "query"
)

var (
	ErrAPIURLRequired = errors.New("prometheus api url is required")
	ErrInvalidAPIURL  = errors.New("prometheus api url is invalid")
)

type Connector struct {
	client *http.Client
}

func NewConnector() *Connector {
	return &Connector{client: &http.Client{Timeout: defaultTimeout}}
}

func (c *Connector) Connect(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	return c.TestConnection(ctx, cfg)
}

func (c *Connector) Disconnect(_ context.Context, _ connector.Config) error {
	return nil
}

func (c *Connector) TestConnection(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	baseURL, err := extractAPIURL(cfg)
	if err != nil {
		return nil, err
	}

	result, err := c.query(ctx, baseURL, extractBearerToken(cfg), "up")
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return &connector.Result{Success: false, Message: result.Message}, nil
	}
	return &connector.Result{Success: true, Message: "Prometheus API is reachable and responding successfully."}, nil
}

func (c *Connector) HealthCheck(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	return c.TestConnection(ctx, cfg)
}

func (c *Connector) Capabilities() connector.Capabilities {
	return connector.Capabilities{
		Implemented:            true,
		SupportsTestConnection: true,
		SupportsHealthCheck:    true,
		Description:            "Query Prometheus-compatible time series data in read-only mode via the HTTP API.",
	}
}

func (c *Connector) SendNotification(_ context.Context, _ connector.Config, _ notification.Message) error {
	return fmt.Errorf("prometheus integration does not support outbound notification dispatch")
}

func (c *Connector) query(ctx context.Context, baseURL, bearerToken, query string) (*connector.Result, error) {
	endpoint, err := buildQueryURL(baseURL, query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build Prometheus query request: %w", err)
	}

	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call Prometheus API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("prometheus api returned status %d", resp.StatusCode)
	}

	var payload struct {
		Status string `json:"status"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode Prometheus API response: %w", err)
	}
	if payload.Status == "error" {
		return &connector.Result{Success: false, Message: payload.Error}, nil
	}
	return &connector.Result{Success: true, Message: "Prometheus query succeeded"}, nil
}

func extractAPIURL(cfg connector.Config) (string, error) {
	for _, key := range []string{"api_url", "apiUrl", "url", "prometheus_url"} {
		if value := strings.TrimSpace(cfg.Metadata[key]); value != "" {
			return validateAPIURL(value)
		}
	}
	for _, key := range []string{"api_url", "apiUrl", "url", "prometheus_url"} {
		if value := strings.TrimSpace(cfg.Credentials[key]); value != "" {
			return validateAPIURL(value)
		}
	}
	return "", ErrAPIURLRequired
}

func extractBearerToken(cfg connector.Config) string {
	credentialSources := []map[string]string{}
	if cfg.Credentials != nil {
		credentialSources = append(credentialSources, cfg.Credentials)
	}
	if cfg.Metadata != nil {
		credentialSources = append(credentialSources, cfg.Metadata)
	}

	for _, source := range credentialSources {
		for _, key := range []string{"bearer_token", "bearerToken", "token", "access_token", "authorization_token"} {
			if value := strings.TrimSpace(source[key]); value != "" {
				return value
			}
		}
	}
	return ""
}

func validateAPIURL(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil {
		return "", ErrInvalidAPIURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidAPIURL
	}
	if parsed.Host == "" {
		return "", ErrInvalidAPIURL
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func buildQueryURL(baseURL, query string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", ErrInvalidAPIURL
	}
	queryURL := *parsed
	q := queryURL.Query()
	q.Set(queryParamName, query)
	queryURL.RawQuery = q.Encode()
	queryURL.Path = strings.TrimRight(queryURL.Path, "/") + "/api/v1/query"
	if queryURL.Path == "/api/v1/query" {
		queryURL.Path = "/api/v1/query"
	}
	return queryURL.String(), nil
}
