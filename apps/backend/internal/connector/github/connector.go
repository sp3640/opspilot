package github

import (
	"context"
	"fmt"

	"github.com/sp3640/opspilot/backend/internal/connector"
)

// tokenKey is the credentials map key IntegrationService/GitHubService use
// to store the GitHub access token - both the OAuth-issued token and a
// manually-pasted personal access token are stored under this same key, so
// Connect/TestConnection/HealthCheck never need to know which path issued it.
const tokenKey = "access_token"

// Connector adapts Client to the generic internal/connector.Connector
// interface from Sprint 27, so the existing, unmodified
// IntegrationService.TestConnection/CheckStatus and the generic
// POST /integrations/:id/test and /check endpoints work correctly for a
// "github" integration the moment this Connector is registered - no
// Sprint 27 file needs to change for that to happen.
type Connector struct {
	client *Client
}

func NewConnector(client *Client) *Connector {
	return &Connector{client: client}
}

// Connect verifies the given credentials actually work. GitHub has no
// separate "connect" step beyond having a valid token, so this simply
// delegates to TestConnection.
func (c *Connector) Connect(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	return c.TestConnection(ctx, cfg)
}

// Disconnect is a local, no-op lifecycle transition only - it never calls
// GitHub to revoke the underlying OAuth grant/token, since revoking a grant
// is a GitHub *write* operation and this sprint is read-only by design. An
// operator who wants to fully revoke access should do so from their GitHub
// account settings; OpsPilot only forgets the credential locally (handled
// by IntegrationService.DeleteIntegration, not by this method).
func (c *Connector) Disconnect(_ context.Context, _ connector.Config) error {
	return nil
}

func (c *Connector) TestConnection(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	token, err := extractToken(cfg)
	if err != nil {
		return nil, err
	}

	user, err := c.client.AuthenticatedUser(ctx, token)
	if err != nil {
		return nil, err
	}

	return &connector.Result{Success: true, Message: fmt.Sprintf("Connected as GitHub user %s", user.Login)}, nil
}

// HealthCheck reuses the same identity lookup as TestConnection: GitHub
// offers no cheaper, more specific "still healthy" signal than "is this
// token still valid", so the two are intentionally identical here.
func (c *Connector) HealthCheck(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	return c.TestConnection(ctx, cfg)
}

func (c *Connector) Capabilities() connector.Capabilities {
	return connector.Capabilities{
		Implemented:            true,
		SupportsTestConnection: true,
		SupportsHealthCheck:    true,
		Description:            "Discover repositories and read commit/pull request history via the GitHub REST API. Read-only - no write operations are supported.",
	}
}

func extractToken(cfg connector.Config) (string, error) {
	token := cfg.Credentials[tokenKey]
	if token == "" {
		return "", ErrTokenRequired
	}
	return token, nil
}
