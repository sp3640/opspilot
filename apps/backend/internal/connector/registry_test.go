package connector

import (
	"context"
	"testing"
)

func TestRegistryResolveUnregisteredTypeReturnsNotImplemented(t *testing.T) {
	registry := NewRegistry()

	c := registry.Resolve("github")
	if c == nil {
		t.Fatalf("expected Resolve to always return a non-nil connector")
	}

	capabilities := c.Capabilities()
	if capabilities.Implemented {
		t.Fatalf("expected an unregistered type to report Implemented=false")
	}

	if _, err := c.Connect(context.Background(), Config{}); err == nil {
		t.Fatalf("expected Connect to fail for an unimplemented connector")
	}
	if _, err := c.TestConnection(context.Background(), Config{}); err == nil {
		t.Fatalf("expected TestConnection to fail for an unimplemented connector")
	}
	if _, err := c.HealthCheck(context.Background(), Config{}); err == nil {
		t.Fatalf("expected HealthCheck to fail for an unimplemented connector")
	}
	if err := c.Disconnect(context.Background(), Config{}); err == nil {
		t.Fatalf("expected Disconnect to fail for an unimplemented connector")
	}
}

func TestRegistryResolveNeverReturnsAFakeSuccess(t *testing.T) {
	registry := NewRegistry()

	for _, integrationType := range []string{"github", "slack", "email", "prometheus", "loki", "opentelemetry", "azure", "teams", "webhook"} {
		result, err := registry.Resolve(integrationType).TestConnection(context.Background(), Config{})
		if err == nil {
			t.Fatalf("expected %q to fail rather than fabricate a successful test connection", integrationType)
		}
		if result != nil && result.Success {
			t.Fatalf("expected %q to never report Success=true", integrationType)
		}
	}
}

type stubConnector struct{}

func (stubConnector) Connect(_ context.Context, _ Config) (*Result, error) {
	return &Result{Success: true, Message: "connected"}, nil
}
func (stubConnector) Disconnect(_ context.Context, _ Config) error { return nil }
func (stubConnector) TestConnection(_ context.Context, _ Config) (*Result, error) {
	return &Result{Success: true, Message: "ok"}, nil
}
func (stubConnector) HealthCheck(_ context.Context, _ Config) (*Result, error) {
	return &Result{Success: true, Message: "healthy"}, nil
}
func (stubConnector) Capabilities() Capabilities {
	return Capabilities{Implemented: true, SupportsTestConnection: true, SupportsHealthCheck: true, Description: "stub"}
}

func TestRegistryResolveReturnsRegisteredConnector(t *testing.T) {
	registry := NewRegistry()
	registry.Register("slack", stubConnector{})

	resolved := registry.Resolve("slack")
	if !resolved.Capabilities().Implemented {
		t.Fatalf("expected the registered connector's capabilities to report Implemented=true")
	}

	// An unrelated, never-registered type must still fall back safely.
	if registry.Resolve("github").Capabilities().Implemented {
		t.Fatalf("expected an unrelated unregistered type to remain unimplemented")
	}
}
