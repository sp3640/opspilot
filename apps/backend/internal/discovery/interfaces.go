package discovery

import (
	"context"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/monitoring"
)

type DiscoveryProvider interface {
	Validate(ctx context.Context) error
	Discover(ctx context.Context) (*DiscoveryResult, error)
	Name() string
	Provider() monitoring.ProviderType
}

type ProviderRegistration struct {
	ProjectID uuid.UUID
	ClusterID uuid.UUID
	Provider  DiscoveryProvider
}

type DiscoveryExecution struct {
	Job    DiscoveryJob
	Result *DiscoveryResult
	Error  error
}

type ResourceSet []models.Resource
