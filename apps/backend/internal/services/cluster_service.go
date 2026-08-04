package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/mapper"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/security"
	"gorm.io/gorm"
)

type ClusterService struct {
	repo             *repository.ClusterRepository
	auditRepo        *AuditService
	credentialCipher security.ClusterCredentialCipher
}

func NewClusterService(repo *repository.ClusterRepository, auditService *AuditService, credentialCipher security.ClusterCredentialCipher) *ClusterService {
	return &ClusterService{
		repo:             repo,
		auditRepo:        auditService,
		credentialCipher: credentialCipher,
	}
}

func (s *ClusterService) CreateCluster(
	ctx context.Context,
	projectID uuid.UUID,
	name,
	provider,
	status,
	connectionType,
	kubeconfigEncrypted,
	apiEndpoint,
	region,
	version,
	validationError string,
	metadata json.RawMessage,
	lastValidatedAt,
	lastDiscoveryAt *time.Time,
	userID uint,
	organizationID uuid.UUID,
) (*dto.ClusterResponse, error) {
	provider = strings.TrimSpace(strings.ToUpper(provider))
	status = strings.TrimSpace(strings.ToUpper(status))
	connectionType = strings.TrimSpace(strings.ToUpper(connectionType))
	name = strings.TrimSpace(name)
	kubeconfigEncrypted = strings.TrimSpace(kubeconfigEncrypted)
	apiEndpoint = strings.TrimSpace(apiEndpoint)
	region = strings.TrimSpace(region)
	version = strings.TrimSpace(version)
	validationError = strings.TrimSpace(validationError)
	encryptedCredential, err := s.encryptCredential(kubeconfigEncrypted)
	if err != nil {
		return nil, err
	}

	if err := validateClusterInput(provider, status, connectionType); err != nil {
		return nil, err
	}

	belongs, err := s.repo.ProjectBelongsToOrganization(projectID, organizationID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	resolvedMetadata := metadata
	if len(resolvedMetadata) == 0 {
		resolvedMetadata = json.RawMessage(`{}`)
	}

	isDefault := false
	if _, err := s.repo.GetDefaultCluster(projectID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			isDefault = true
		} else {
			return nil, err
		}
	}

	cluster := &models.Cluster{
		OrganizationID:      organizationID,
		ProjectID:           projectID,
		Name:                name,
		Provider:            provider,
		ConnectionType:      connectionType,
		CredentialType:      connectionType,
		Status:              status,
		IsDefault:           isDefault,
		KubeconfigEncrypted: encryptedCredential,
		EncryptedCredential: encryptedCredential,
		APIEndpoint:         apiEndpoint,
		Region:              region,
		Version:             version,
		LastValidatedAt:     lastValidatedAt,
		LastDiscoveryAt:     lastDiscoveryAt,
		CreatedBy:           userID,
		ValidationError:     validationError,
		Metadata:            resolvedMetadata,
	}

	if err := s.repo.Create(cluster); err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogCreate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil); err != nil {
			logAuditFailure(ctx, "create", "cluster", 0, err)
		}
		if cluster.IsDefault {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "is_default", "false", "true"); err != nil {
				logAuditFailure(ctx, "default_change", "cluster", 0, err)
			}
		}
	}

	response := mapper.MapCluster(*cluster)
	return &response, nil
}

func (s *ClusterService) UpdateCluster(
	ctx context.Context,
	id uuid.UUID,
	userID uint,
	organizationID uuid.UUID,
	projectID uuid.UUID,
	name,
	provider,
	status,
	connectionType,
	kubeconfigEncrypted,
	apiEndpoint,
	region,
	version,
	validationError string,
	metadata json.RawMessage,
	lastValidatedAt,
	lastDiscoveryAt *time.Time,
) (*dto.ClusterResponse, error) {
	provider = strings.TrimSpace(strings.ToUpper(provider))
	status = strings.TrimSpace(strings.ToUpper(status))
	connectionType = strings.TrimSpace(strings.ToUpper(connectionType))
	name = strings.TrimSpace(name)
	kubeconfigEncrypted = strings.TrimSpace(kubeconfigEncrypted)
	apiEndpoint = strings.TrimSpace(apiEndpoint)
	region = strings.TrimSpace(region)
	version = strings.TrimSpace(version)
	validationError = strings.TrimSpace(validationError)
	encryptedCredential, err := s.encryptCredential(kubeconfigEncrypted)
	if err != nil {
		return nil, err
	}

	if err := validateClusterInput(provider, status, connectionType); err != nil {
		return nil, err
	}

	cluster, err := s.getOwnedCluster(id, organizationID)
	if err != nil {
		return nil, err
	}

	belongs, err := s.repo.ProjectBelongsToOrganization(projectID, organizationID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrInvalidProject
	}

	previousProjectID := cluster.ProjectID
	previousName := cluster.Name
	previousProvider := cluster.Provider
	previousStatus := cluster.Status
	previousIsDefault := cluster.IsDefault
	previousConnectionType := cluster.ConnectionType
	previousKubeconfigEncrypted := cluster.KubeconfigEncrypted
	previousAPIEndpoint := cluster.APIEndpoint
	previousRegion := cluster.Region
	previousVersion := cluster.Version
	previousValidationError := cluster.ValidationError
	previousLastValidatedAt := cluster.LastValidatedAt
	previousLastDiscoveryAt := cluster.LastDiscoveryAt

	shouldBecomeDefault := cluster.IsDefault
	if previousProjectID != projectID {
		shouldBecomeDefault = false
		if _, err := s.repo.GetDefaultCluster(projectID, organizationID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				shouldBecomeDefault = true
			} else {
				return nil, err
			}
		}
	}

	cluster.ProjectID = projectID
	cluster.Name = name
	cluster.Provider = provider
	cluster.ConnectionType = connectionType
	cluster.CredentialType = connectionType
	cluster.Status = status
	cluster.IsDefault = shouldBecomeDefault
	cluster.KubeconfigEncrypted = encryptedCredential
	cluster.EncryptedCredential = encryptedCredential
	cluster.APIEndpoint = apiEndpoint
	cluster.Region = region
	cluster.Version = version
	cluster.LastValidatedAt = lastValidatedAt
	cluster.LastDiscoveryAt = lastDiscoveryAt
	cluster.ValidationError = validationError
	if len(metadata) > 0 {
		cluster.Metadata = metadata
	}

	if err := s.repo.Update(cluster); err != nil {
		return nil, err
	}

	if previousProjectID != projectID && previousIsDefault {
		if err := s.assignNewDefaultCluster(ctx, previousProjectID, userID, organizationID); err != nil {
			return nil, err
		}
	}

	if s.auditRepo != nil {
		if previousProjectID != cluster.ProjectID {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "project_id", previousProjectID.String(), cluster.ProjectID.String()); err != nil {
				logAuditFailure(ctx, "update", "cluster", 0, err)
			}
		}
		if previousName != cluster.Name {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "name", previousName, cluster.Name); err != nil {
				logAuditFailure(ctx, "update", "cluster", 0, err)
			}
		}
		if previousProvider != cluster.Provider {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "provider", previousProvider, cluster.Provider); err != nil {
				logAuditFailure(ctx, "update", "cluster", 0, err)
			}
		}
		if previousStatus != cluster.Status {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "status", previousStatus, cluster.Status); err != nil {
				logAuditFailure(ctx, "validation_status_change", "cluster", 0, err)
			}
		}
		if previousIsDefault != cluster.IsDefault {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "is_default", boolString(previousIsDefault), boolString(cluster.IsDefault)); err != nil {
				logAuditFailure(ctx, "default_change", "cluster", 0, err)
			}
		}
		if previousConnectionType != cluster.ConnectionType {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "connection_type", previousConnectionType, cluster.ConnectionType); err != nil {
				logAuditFailure(ctx, "update", "cluster", 0, err)
			}
		}
		if previousKubeconfigEncrypted != cluster.KubeconfigEncrypted {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "kubeconfig_encrypted", previousKubeconfigEncrypted, cluster.KubeconfigEncrypted); err != nil {
				logAuditFailure(ctx, "update", "cluster", 0, err)
			}
		}
		if previousAPIEndpoint != cluster.APIEndpoint {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "api_endpoint", previousAPIEndpoint, cluster.APIEndpoint); err != nil {
				logAuditFailure(ctx, "update", "cluster", 0, err)
			}
		}
		if previousRegion != cluster.Region {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "region", previousRegion, cluster.Region); err != nil {
				logAuditFailure(ctx, "update", "cluster", 0, err)
			}
		}
		if previousVersion != cluster.Version {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "version", previousVersion, cluster.Version); err != nil {
				logAuditFailure(ctx, "update", "cluster", 0, err)
			}
		}
		if previousValidationError != cluster.ValidationError {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "validation_error", previousValidationError, cluster.ValidationError); err != nil {
				logAuditFailure(ctx, "validation_status_change", "cluster", 0, err)
			}
		}
		if !timePointersEqual(previousLastValidatedAt, cluster.LastValidatedAt) {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "last_validated_at", timePointerString(previousLastValidatedAt), timePointerString(cluster.LastValidatedAt)); err != nil {
				logAuditFailure(ctx, "validation_status_change", "cluster", 0, err)
			}
		}
		if !timePointersEqual(previousLastDiscoveryAt, cluster.LastDiscoveryAt) {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "last_discovery_at", timePointerString(previousLastDiscoveryAt), timePointerString(cluster.LastDiscoveryAt)); err != nil {
				logAuditFailure(ctx, "discovery_timestamp_update", "cluster", 0, err)
			}
		}
	}

	response := mapper.MapCluster(*cluster)
	return &response, nil
}

func (s *ClusterService) DeleteCluster(ctx context.Context, id uuid.UUID, userID uint, organizationID uuid.UUID) error {
	cluster, err := s.getOwnedCluster(id, organizationID)
	if err != nil {
		return err
	}

	projectID := cluster.ProjectID
	wasDefault := cluster.IsDefault

	if err := s.repo.Delete(cluster.ID, organizationID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogDelete(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil); err != nil {
			logAuditFailure(ctx, "delete", "cluster", 0, err)
		}
	}

	if wasDefault {
		if err := s.assignNewDefaultCluster(ctx, projectID, userID, organizationID); err != nil {
			return err
		}
	}

	return nil
}

func (s *ClusterService) GetCluster(id uuid.UUID, organizationID uuid.UUID) (*dto.ClusterResponse, error) {
	cluster, err := s.getOwnedCluster(id, organizationID)
	if err != nil {
		return nil, err
	}

	response := mapper.MapCluster(*cluster)
	return &response, nil
}

func (s *ClusterService) ListClusters(organizationID uuid.UUID, req *models.PaginationRequest) (*dto.ClusterListResponse, error) {
	items, total, err := s.repo.List(req, organizationID)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &dto.ClusterListResponse{
		Items:      mapper.MapClusters(items),
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *ClusterService) SetDefaultCluster(ctx context.Context, id uuid.UUID, userID uint, organizationID uuid.UUID) (*dto.ClusterResponse, error) {
	cluster, err := s.getOwnedCluster(id, organizationID)
	if err != nil {
		return nil, err
	}

	previousDefault, err := s.repo.GetDefaultCluster(cluster.ProjectID, organizationID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.repo.SetDefaultCluster(cluster.ProjectID, cluster.ID, organizationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	if s.auditRepo != nil {
		if previousDefault != nil && previousDefault.ID != cluster.ID {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", previousDefault.ID.String(), &cluster.ProjectID, nil, "is_default", "true", "false"); err != nil {
				logAuditFailure(ctx, "default_change", "cluster", 0, err)
			}
		}
		if previousDefault == nil || previousDefault.ID != cluster.ID {
			if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", cluster.ID.String(), &cluster.ProjectID, nil, "is_default", "false", "true"); err != nil {
				logAuditFailure(ctx, "default_change", "cluster", 0, err)
			}
		}
	}

	updatedCluster, err := s.repo.FindByID(cluster.ID, organizationID)
	if err != nil {
		return nil, err
	}

	response := mapper.MapCluster(*updatedCluster)
	return &response, nil
}

func (s *ClusterService) ValidateClusterCredential(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*dto.ValidationResponse, error) {
	cluster, err := s.getOwnedCluster(id, organizationID)
	if err != nil {
		return nil, err
	}

	validatedAt := time.Now().UTC()
	response := &dto.ValidationResponse{
		Status:      constants.ClusterStatusPendingValidation,
		Healthy:     true,
		ValidatedAt: validatedAt,
	}

	if _, err := s.decryptCredential(cluster.KubeconfigEncrypted); err != nil {
		response.Status = constants.ClusterStatusInvalid
		response.Healthy = false
		response.Error = err.Error()
		if updateErr := s.repo.UpdateValidation(cluster.ID, organizationID, response.Status, &validatedAt, response.Error); updateErr != nil {
			return nil, updateErr
		}
	} else {
		if err := s.repo.UpdateValidation(cluster.ID, organizationID, response.Status, &validatedAt, ""); err != nil {
			return nil, err
		}
	}

	updatedCluster, err := s.repo.FindByID(cluster.ID, organizationID)
	if err != nil {
		return nil, err
	}
	response.Cluster = ptrClusterResponse(mapper.MapCluster(*updatedCluster))

	return response, nil
}

func (s *ClusterService) getOwnedCluster(id uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	cluster, err := s.repo.FindByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	belongs, err := s.repo.ProjectBelongsToOrganization(cluster.ProjectID, organizationID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrProjectForbidden
	}

	return cluster, nil
}

func (s *ClusterService) assignNewDefaultCluster(ctx context.Context, projectID uuid.UUID, userID uint, organizationID uuid.UUID) error {
	clusters, err := s.repo.FindByProject(projectID, organizationID)
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		return nil
	}

	newDefault := clusters[0]
	if newDefault.IsDefault {
		return nil
	}

	if err := s.repo.SetDefaultCluster(projectID, newDefault.ID, organizationID); err != nil {
		return err
	}

	if s.auditRepo != nil {
		if err := s.auditRepo.LogUpdate(userID, organizationID, "cluster", newDefault.ID.String(), &projectID, nil, "is_default", "false", "true"); err != nil {
			logAuditFailure(ctx, "default_change", "cluster", 0, err)
		}
	}

	return nil
}

func validateClusterInput(provider, status, connectionType string) error {
	if !constants.IsValidClusterProvider(provider) {
		return apperrors.ErrInvalidClusterProvider
	}
	if !constants.IsValidClusterStatus(status) {
		return apperrors.ErrInvalidClusterStatus
	}
	if !constants.IsValidClusterConnectionType(connectionType) {
		return apperrors.ErrInvalidClusterConnectionType
	}

	return nil
}

func (s *ClusterService) encryptCredential(credential string) (string, error) {
	if s.credentialCipher == nil {
		return "", errors.New("cluster credential cipher is required")
	}

	return s.credentialCipher.Encrypt(credential)
}

func (s *ClusterService) decryptCredential(ciphertext string) (string, error) {
	if s.credentialCipher == nil {
		return "", errors.New("cluster credential cipher is required")
	}

	return s.credentialCipher.Decrypt(ciphertext)
}

func ptrClusterResponse(resp dto.ClusterResponse) *dto.ClusterResponse {
	return &resp
}

func boolString(value bool) string {
	if value {
		return "true"
	}

	return "false"
}

func timePointerString(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.RFC3339)
}

func timePointersEqual(left, right *time.Time) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	return left.Equal(*right)
}
