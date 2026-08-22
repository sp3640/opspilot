package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	kubeintegration "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/security"
)

const applicationMetricsUnavailableReason = "No application observability provider (e.g. Prometheus/APM) is connected yet."

// MetricsStatusService answers "is this cluster's metrics provider actually
// working right now" - live reachability checks plus when data was last
// collected - rather than leaving the frontend to guess from an empty
// history array whether nothing has happened yet or the provider is down.
type MetricsStatusService struct {
	clusterRepo      *repository.ClusterRepository
	metricRepo       *repository.MetricRepository
	credentialCipher security.ClusterCredentialCipher
}

func NewMetricsStatusService(clusterRepo *repository.ClusterRepository, metricRepo *repository.MetricRepository, credentialCipher security.ClusterCredentialCipher) *MetricsStatusService {
	return &MetricsStatusService{
		clusterRepo:      clusterRepo,
		metricRepo:       metricRepo,
		credentialCipher: credentialCipher,
	}
}

func (s *MetricsStatusService) GetClusterMetricsStatus(ctx context.Context, organizationID uuid.UUID, clusterID uuid.UUID) (*dto.MetricsStatusResponse, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	response := &dto.MetricsStatusResponse{
		ClusterID: clusterID.String(),
		ApplicationMetrics: dto.ApplicationMetricsStatusResponse{
			Available: false,
			Reason:    applicationMetricsUnavailableReason,
		},
	}

	if s.credentialCipher == nil {
		return response, nil
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return response, nil
	}

	client := kubeintegration.NewClient([]byte(kubeconfig))
	collector := kubeintegration.NewMetricsCollector("status-check", client)
	response.KubernetesReachable = collector.Supports(ctx)

	if response.KubernetesReachable {
		response.MetricsServerAvailable = kubeintegration.MetricsServerReachable(ctx, client)
	}

	latest, err := s.metricRepo.Latest(cluster.ProjectID, &clusterID, nil, constants.MetricTypeCPU, "cluster.cpu.usage.millicores", organizationID)
	if err == nil && latest != nil {
		response.LastCollectedAt = &latest.Timestamp
	}

	return response, nil
}
