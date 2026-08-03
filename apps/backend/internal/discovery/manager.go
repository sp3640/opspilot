package discovery

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/monitoring"
)

type DiscoveryManager struct {
	mu        sync.RWMutex
	providers map[string]ProviderRegistration
	engine    *DiscoveryEngine
}

func NewDiscoveryManager(engine *DiscoveryEngine) *DiscoveryManager {
	if engine == nil {
		engine = NewDiscoveryEngine()
	}

	return &DiscoveryManager{
		providers: make(map[string]ProviderRegistration),
		engine:    engine,
	}
}

func (m *DiscoveryManager) RegisterProvider(provider DiscoveryProvider, projectID, clusterID uuid.UUID) {
	if m == nil || provider == nil {
		return
	}

	key := providerKey(provider.Provider(), provider.Name(), projectID, clusterID)

	m.mu.Lock()
	m.providers[key] = ProviderRegistration{
		ProjectID: projectID,
		ClusterID: clusterID,
		Provider:  provider,
	}
	m.mu.Unlock()
}

func (m *DiscoveryManager) RemoveProvider(providerType monitoring.ProviderType, name string, projectID, clusterID uuid.UUID) {
	if m == nil {
		return
	}

	key := providerKey(providerType, name, projectID, clusterID)

	m.mu.Lock()
	delete(m.providers, key)
	m.mu.Unlock()
}

func (m *DiscoveryManager) GetProvider(providerType monitoring.ProviderType, name string, projectID, clusterID uuid.UUID) (ProviderRegistration, bool) {
	if m == nil {
		return ProviderRegistration{}, false
	}

	key := providerKey(providerType, name, projectID, clusterID)

	m.mu.RLock()
	registration, ok := m.providers[key]
	m.mu.RUnlock()

	return registration, ok
}

func (m *DiscoveryManager) ListProviders() []ProviderRegistration {
	if m == nil {
		return nil
	}

	m.mu.RLock()
	providers := make([]ProviderRegistration, 0, len(m.providers))
	for _, registration := range m.providers {
		providers = append(providers, registration)
	}
	m.mu.RUnlock()

	return providers
}

func (m *DiscoveryManager) ExecuteDiscovery(ctx context.Context, providerType monitoring.ProviderType, name string, projectID, clusterID uuid.UUID) (*DiscoveryExecution, error) {
	registration, ok := m.GetProvider(providerType, name, projectID, clusterID)
	if !ok {
		return nil, fmt.Errorf("discovery provider not found: %s/%s", providerType, name)
	}

	return m.execute(ctx, registration)
}

func (m *DiscoveryManager) ExecuteDiscoveryForProject(ctx context.Context, projectID uuid.UUID) ([]DiscoveryExecution, error) {
	registrations := m.listByProject(projectID)
	results := make([]DiscoveryExecution, 0, len(registrations))

	for _, registration := range registrations {
		execution, err := m.execute(ctx, registration)
		if execution != nil {
			results = append(results, *execution)
		}
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

func (m *DiscoveryManager) ExecuteDiscoveryForCluster(ctx context.Context, clusterID uuid.UUID) ([]DiscoveryExecution, error) {
	registrations := m.listByCluster(clusterID)
	results := make([]DiscoveryExecution, 0, len(registrations))

	for _, registration := range registrations {
		execution, err := m.execute(ctx, registration)
		if execution != nil {
			results = append(results, *execution)
		}
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

func (m *DiscoveryManager) execute(ctx context.Context, registration ProviderRegistration) (*DiscoveryExecution, error) {
	if m == nil || registration.Provider == nil {
		return nil, fmt.Errorf("discovery provider is required")
	}

	job := m.engine.StartJob(registration.Provider.Provider(), registration.ProjectID, registration.ClusterID)

	if err := registration.Provider.Validate(ctx); err != nil {
		finishedJob, finishErr := m.engine.FinishJob(job.ID, nil, err)
		if finishErr != nil {
			return nil, finishErr
		}

		return &DiscoveryExecution{Job: finishedJob, Error: err}, err
	}

	result, err := registration.Provider.Discover(ctx)
	finishedJob, finishErr := m.engine.FinishJob(job.ID, result, err)
	if finishErr != nil {
		return nil, finishErr
	}

	return &DiscoveryExecution{
		Job:    finishedJob,
		Result: result,
		Error:  err,
	}, err
}

func (m *DiscoveryManager) listByProject(projectID uuid.UUID) []ProviderRegistration {
	if m == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]ProviderRegistration, 0)
	for _, registration := range m.providers {
		if registration.ProjectID == projectID {
			providers = append(providers, registration)
		}
	}

	return providers
}

func (m *DiscoveryManager) listByCluster(clusterID uuid.UUID) []ProviderRegistration {
	if m == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]ProviderRegistration, 0)
	for _, registration := range m.providers {
		if registration.ClusterID == clusterID {
			providers = append(providers, registration)
		}
	}

	return providers
}

func providerKey(providerType monitoring.ProviderType, name string, projectID, clusterID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:%s:%s", providerType, name, projectID.String(), clusterID.String())
}
