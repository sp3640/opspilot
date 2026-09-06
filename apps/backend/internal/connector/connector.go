// Package connector defines the reusable abstraction future GitHub, Slack,
// Email, Prometheus, Loki, OpenTelemetry, and Azure integrations will
// implement (Sprint 27: Integration Foundation). No real connector is
// implemented in this package yet - only the interface, a Registry that
// resolves an integration type to its Connector, and a safe fallback for
// any type that has none registered.
//
// A Connector never touches HTTP handlers or the database: IntegrationService
// resolves credentials/configuration from storage, decrypts them, and passes
// a Config in; the Connector returns a structured Result or error and knows
// nothing about persistence, organizations, or audit logging.
package connector

import "context"

// Config is everything a Connector needs to act, already resolved by the
// caller (IntegrationService). Credentials are the decrypted secret fields
// for exactly as long as one call takes - a Connector must never persist,
// log, or otherwise retain them beyond the call it was given for.
type Config struct {
	Metadata    map[string]string
	Credentials map[string]string
}

// Result is the outcome of a connect/test/health-check attempt. Message is
// safe, human-readable text - never a raw external API error body (that
// could itself leak sensitive detail about the remote system).
type Result struct {
	Success bool
	Message string
}

// Capabilities describes what a Connector can actually do, so callers (and
// the frontend) never have to guess whether a "successful" call means
// anything. Implemented=false is the correct, honest answer for every
// integration type in this sprint.
type Capabilities struct {
	Implemented            bool
	SupportsTestConnection bool
	SupportsHealthCheck    bool
	Description            string
}

// Connector is the operation set every future integration type will
// implement. Connect/Disconnect model the integration's own lifecycle;
// TestConnection validates that the given configuration/credentials
// actually work; HealthCheck is a lighter-weight, repeatable check used for
// the ongoing "is this still working" status refresh.
type Connector interface {
	Connect(ctx context.Context, cfg Config) (*Result, error)
	Disconnect(ctx context.Context, cfg Config) error
	TestConnection(ctx context.Context, cfg Config) (*Result, error)
	HealthCheck(ctx context.Context, cfg Config) (*Result, error)
	Capabilities() Capabilities
}
