package bootstrap

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/discovery"
	kubeintegration "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
)

type RuntimeCluster struct {
	ID                  uuid.UUID
	ProjectID           uuid.UUID
	Name                string
	Provider            string
	Status              string
	KubeconfigEncrypted []byte
	CreatedBy           uint
}

type DiscoveryBootstrap struct {
	Executor       *discovery.DiscoveryExecutor
	Runner         *discovery.BackgroundRunner
	ProviderFactor discovery.DiscoveryProviderFactory
}

func NewDiscoveryBootstrap(interval time.Duration, maxConcurrent int, worker *discovery.DiscoveryWorker, providerFactory discovery.DiscoveryProviderFactory) *DiscoveryBootstrap {
	if interval <= 0 {
		interval = defaultDiscoveryInterval
	}
	if providerFactory == nil {
		providerFactory = NewKubernetesDiscoveryProviderFactory()
	}

	executor := discovery.NewDiscoveryExecutor(maxConcurrent)
	runner := discovery.NewBackgroundRunner(interval, worker, executor)

	return &DiscoveryBootstrap{
		Executor:       executor,
		Runner:         runner,
		ProviderFactor: providerFactory,
	}
}

type KubernetesDiscoveryProviderFactory struct{}

func NewKubernetesDiscoveryProviderFactory() *KubernetesDiscoveryProviderFactory {
	return &KubernetesDiscoveryProviderFactory{}
}

func (f *KubernetesDiscoveryProviderFactory) NewDiscoveryProvider(cluster *discovery.ClusterDescriptor) (discovery.DiscoveryProvider, error) {
	if cluster == nil {
		return nil, nil
	}

	provider := strings.TrimSpace(strings.ToUpper(cluster.Provider))
	if provider != constants.ClusterProviderKubernetes {
		return nil, nil
	}

	client := kubeintegration.NewClient(cluster.KubeconfigEncrypted)
	validator := kubeintegration.NewValidator(client)

	return kubeintegration.NewDiscoveryProvider(cluster.ProjectID, cluster.ID, cluster.Name, cluster.CreatedBy, client, validator), nil
}
