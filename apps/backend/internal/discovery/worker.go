package discovery

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/monitoring"
	"github.com/sp3640/opspilot/backend/internal/resourcesync"
)

var (
	ErrClusterDisconnected    = errors.New("cluster is not connected")
	ErrDiscoveryNotSupported  = errors.New("discovery provider is not supported for cluster")
	ErrClusterLoaderRequired  = errors.New("cluster loader is required")
	ErrResourceSyncRequired   = errors.New("resource sync service is required")
	ErrProviderFactoryMissing = errors.New("discovery provider factory is required")
)

type ClusterDescriptor struct {
	ID                  uuid.UUID
	OrganizationID      uuid.UUID
	ProjectID           uuid.UUID
	Name                string
	Provider            string
	Status              string
	KubeconfigEncrypted []byte
	CreatedBy           uint
}

type ClusterLoader interface {
	LoadCluster(ctx context.Context, clusterID uuid.UUID) (*ClusterDescriptor, error)
}

type ResourceSyncService interface {
	SyncResources(ctx context.Context, projectID uuid.UUID, userID uint, organizationID uuid.UUID, discoveredResources []models.Resource) (*resourcesync.SyncResult, error)
}

type DiscoveryProviderFactory interface {
	NewDiscoveryProvider(cluster *ClusterDescriptor) (DiscoveryProvider, error)
}

type DiscoveryAuditor interface {
	LogCreate(userID uint, organizationID uuid.UUID, entityType string, entityID string, projectID *uuid.UUID, incidentID *uint) error
	LogUpdate(userID uint, organizationID uuid.UUID, entityType string, entityID string, projectID *uuid.UUID, incidentID *uint, fieldName string, oldValue string, newValue string) error
}

type WorkerResult struct {
	ClusterID   uuid.UUID
	ProjectID   uuid.UUID
	Execution   *DiscoveryExecution
	SyncResult  *resourcesync.SyncResult
	Skipped     bool
	SkipReason  string
	Provider    monitoring.ProviderType
	ResourceSum int
}

type DiscoveryWorker struct {
	manager         *DiscoveryManager
	clusters        ClusterLoader
	resources       ResourceSyncService
	providerFactory DiscoveryProviderFactory
	audit           DiscoveryAuditor
}

func NewDiscoveryWorker(
	manager *DiscoveryManager,
	clusters ClusterLoader,
	resources ResourceSyncService,
	providerFactory DiscoveryProviderFactory,
	audit DiscoveryAuditor,
) *DiscoveryWorker {
	if manager == nil {
		manager = NewDiscoveryManager(nil)
	}

	return &DiscoveryWorker{
		manager:         manager,
		clusters:        clusters,
		resources:       resources,
		providerFactory: providerFactory,
		audit:           audit,
	}
}

func (w *DiscoveryWorker) Run(ctx context.Context, clusterID uuid.UUID) (*WorkerResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if w.clusters == nil {
		return nil, ErrClusterLoaderRequired
	}
	if w.resources == nil {
		return nil, ErrResourceSyncRequired
	}
	if w.providerFactory == nil {
		return nil, ErrProviderFactoryMissing
	}

	cluster, err := w.clusters.LoadCluster(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	status := strings.TrimSpace(strings.ToUpper(cluster.Status))
	if status != constants.ClusterStatusConnected {
		w.auditValidationFailure(cluster, ErrClusterDisconnected)
		return &WorkerResult{
			ClusterID:  cluster.ID,
			ProjectID:  cluster.ProjectID,
			Skipped:    true,
			SkipReason: ErrClusterDisconnected.Error(),
		}, nil
	}

	provider, err := w.providerFactory.NewDiscoveryProvider(cluster)
	if err != nil {
		w.auditDiscoveryFailed(cluster, err)
		return nil, err
	}
	if provider == nil {
		w.auditDiscoveryFailed(cluster, ErrDiscoveryNotSupported)
		return nil, ErrDiscoveryNotSupported
	}

	w.manager.RegisterProvider(provider, cluster.ProjectID, cluster.ID)
	defer w.manager.RemoveProvider(provider.Provider(), provider.Name(), cluster.ProjectID, cluster.ID)

	w.auditDiscoveryStarted(cluster, provider.Provider())

	execution, err := w.manager.ExecuteDiscovery(ctx, provider.Provider(), provider.Name(), cluster.ProjectID, cluster.ID)
	if err != nil {
		if execution != nil && execution.Result == nil {
			w.auditValidationFailure(cluster, err)
		} else {
			w.auditDiscoveryFailed(cluster, err)
		}
		return &WorkerResult{
			ClusterID: cluster.ID,
			ProjectID: cluster.ProjectID,
			Execution: execution,
			Provider:  provider.Provider(),
		}, err
	}
	if execution == nil || execution.Result == nil {
		validationErr := errors.New("discovery execution returned no result")
		w.auditDiscoveryFailed(cluster, validationErr)
		return &WorkerResult{
			ClusterID: cluster.ID,
			ProjectID: cluster.ProjectID,
			Execution: execution,
			Provider:  provider.Provider(),
		}, validationErr
	}

	syncResult, err := w.resources.SyncResources(ctx, cluster.ProjectID, cluster.CreatedBy, cluster.OrganizationID, execution.Result.Resources)
	if err != nil {
		w.auditDiscoveryFailed(cluster, err)
		return &WorkerResult{
			ClusterID:   cluster.ID,
			ProjectID:   cluster.ProjectID,
			Execution:   execution,
			Provider:    provider.Provider(),
			ResourceSum: len(execution.Result.Resources),
		}, err
	}

	w.auditResourcesSynced(cluster, syncResult)
	w.auditDiscoveryCompleted(cluster, len(execution.Result.Resources))

	return &WorkerResult{
		ClusterID:   cluster.ID,
		ProjectID:   cluster.ProjectID,
		Execution:   execution,
		SyncResult:  syncResult,
		Provider:    provider.Provider(),
		ResourceSum: len(execution.Result.Resources),
	}, nil
}

func (w *DiscoveryWorker) auditDiscoveryStarted(cluster *ClusterDescriptor, provider monitoring.ProviderType) {
	if w.audit == nil || cluster == nil {
		return
	}

	_ = w.audit.LogCreate(cluster.CreatedBy, cluster.OrganizationID, "discovery", cluster.ID.String(), &cluster.ProjectID, nil)
	_ = w.audit.LogUpdate(cluster.CreatedBy, cluster.OrganizationID, "discovery", cluster.ID.String(), &cluster.ProjectID, nil, "provider", "", string(provider))
}

func (w *DiscoveryWorker) auditDiscoveryCompleted(cluster *ClusterDescriptor, discoveredCount int) {
	if w.audit == nil || cluster == nil {
		return
	}

	_ = w.audit.LogUpdate(cluster.CreatedBy, cluster.OrganizationID, "discovery", cluster.ID.String(), &cluster.ProjectID, nil, "status", "RUNNING", "COMPLETED")
	_ = w.audit.LogUpdate(cluster.CreatedBy, cluster.OrganizationID, "discovery", cluster.ID.String(), &cluster.ProjectID, nil, "discovered_count", "0", strconv.Itoa(discoveredCount))
}

func (w *DiscoveryWorker) auditDiscoveryFailed(cluster *ClusterDescriptor, discoveryErr error) {
	if w.audit == nil || cluster == nil || discoveryErr == nil {
		return
	}

	_ = w.audit.LogUpdate(cluster.CreatedBy, cluster.OrganizationID, "discovery", cluster.ID.String(), &cluster.ProjectID, nil, "status", "RUNNING", "FAILED")
	_ = w.audit.LogUpdate(cluster.CreatedBy, cluster.OrganizationID, "discovery", cluster.ID.String(), &cluster.ProjectID, nil, "error", "", discoveryErr.Error())
}

func (w *DiscoveryWorker) auditValidationFailure(cluster *ClusterDescriptor, validationErr error) {
	if w.audit == nil || cluster == nil || validationErr == nil {
		return
	}

	_ = w.audit.LogUpdate(cluster.CreatedBy, cluster.OrganizationID, "discovery", cluster.ID.String(), &cluster.ProjectID, nil, "validation", "passed", "failed")
	_ = w.audit.LogUpdate(cluster.CreatedBy, cluster.OrganizationID, "discovery", cluster.ID.String(), &cluster.ProjectID, nil, "validation_error", "", validationErr.Error())
}

func (w *DiscoveryWorker) auditResourcesSynced(cluster *ClusterDescriptor, syncResult *resourcesync.SyncResult) {
	if w.audit == nil || cluster == nil || syncResult == nil {
		return
	}

	total := syncResult.Created + syncResult.Updated + syncResult.Deleted + syncResult.Restored
	_ = w.audit.LogUpdate(cluster.CreatedBy, cluster.OrganizationID, "discovery", cluster.ID.String(), &cluster.ProjectID, nil, "resources_synced", "0", strconv.Itoa(total))
}
