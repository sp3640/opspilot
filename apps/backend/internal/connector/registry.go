package connector

import (
	"context"
	"fmt"
	"sync"

	"github.com/sp3640/opspilot/backend/internal/notification"
)

// notImplementedConnector is what Registry.Resolve returns for any
// integration type with no real connector registered - every type, in this
// sprint. Every operation fails with a clear, honest message rather than a
// caller needing to special-case "no connector found", and rather than this
// package ever fabricating a successful connect/test/health-check.
type notImplementedConnector struct {
	integrationType string
}

func (c *notImplementedConnector) Connect(_ context.Context, _ Config) (*Result, error) {
	return nil, c.err()
}

func (c *notImplementedConnector) Disconnect(_ context.Context, _ Config) error {
	return c.err()
}

func (c *notImplementedConnector) TestConnection(_ context.Context, _ Config) (*Result, error) {
	return nil, c.err()
}

func (c *notImplementedConnector) HealthCheck(_ context.Context, _ Config) (*Result, error) {
	return nil, c.err()
}

func (c *notImplementedConnector) SendNotification(_ context.Context, _ Config, _ notification.Message) error {
	return c.err()
}

func (c *notImplementedConnector) Capabilities() Capabilities {
	return Capabilities{
		Implemented: false,
		Description: fmt.Sprintf("The %s connector is not implemented yet.", c.integrationType),
	}
}

func (c *notImplementedConnector) err() error {
	return fmt.Errorf("connector for integration type %q is not implemented yet", c.integrationType)
}

// Registry resolves an integration type to its Connector. Safe for
// concurrent use.
type Registry struct {
	mu         sync.RWMutex
	connectors map[string]Connector
}

func NewRegistry() *Registry {
	return &Registry{connectors: make(map[string]Connector)}
}

// Register wires a real Connector implementation for integrationType.
// Not called anywhere in this sprint - kept for the future sprints that
// implement GitHub/Slack/Email/Prometheus/Loki/OpenTelemetry/Azure.
func (r *Registry) Register(integrationType string, c Connector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connectors[integrationType] = c
}

// Resolve always returns a non-nil Connector: a registered implementation
// if one exists, otherwise a notImplementedConnector that fails every
// operation clearly. Never returns a fake successful connector.
func (r *Registry) Resolve(integrationType string) Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if c, ok := r.connectors[integrationType]; ok {
		return c
	}
	return &notImplementedConnector{integrationType: integrationType}
}
