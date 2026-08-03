package monitoring

import "context"

type ConnectorType string

const (
	ConnectorTypeDiscovery ConnectorType = "DISCOVERY"
	ConnectorTypeTelemetry ConnectorType = "TELEMETRY"
	ConnectorTypeHybrid    ConnectorType = "HYBRID"
)

type ConnectorHealth string

const (
	ConnectorHealthHealthy   ConnectorHealth = "HEALTHY"
	ConnectorHealthDegraded  ConnectorHealth = "DEGRADED"
	ConnectorHealthUnhealthy ConnectorHealth = "UNHEALTHY"
	ConnectorHealthUnknown   ConnectorHealth = "UNKNOWN"
)

type DiscoverRequest struct {
	ProjectID string
	Metadata  map[string]any
}

type DiscoverResult struct {
	Resources []map[string]any
	Metadata  map[string]any
}

type CollectRequest struct {
	ProjectID string
	Metadata  map[string]any
}

type CollectResult struct {
	Samples  []map[string]any
	Metadata map[string]any
}

type Connector interface {
	ID() string
	Name() string
	Provider() ProviderType
	Type() ConnectorType
	Health() ConnectorHealth
	Discover(ctx context.Context, request DiscoverRequest) (*DiscoverResult, error)
	Collect(ctx context.Context, request CollectRequest) (*CollectResult, error)
	Validate(ctx context.Context) error
}
