package replicasets

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/security"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClientsetFactory func(kubeconfig []byte) (kubernetes.Interface, error)

type ReplicaSetService struct {
	applicationRepo  *repository.ApplicationRepository
	clusterRepo      *repository.ClusterRepository
	credentialCipher security.ClusterCredentialCipher
	mapper           *ReplicaSetMapper
	clientFactory    ClientsetFactory
}

func NewReplicaSetService(
	applicationRepo *repository.ApplicationRepository,
	clusterRepo *repository.ClusterRepository,
	credentialCipher security.ClusterCredentialCipher,
) *ReplicaSetService {

	service := &ReplicaSetService{
		applicationRepo:  applicationRepo,
		clusterRepo:      clusterRepo,
		credentialCipher: credentialCipher,
		mapper:           NewReplicaSetMapper(NewReplicaSetStatusHelper()),
	}

	service.clientFactory = service.defaultClientFactory
	return service
}

func (s *ReplicaSetService) WithClientsetFactory(factory ClientsetFactory) *ReplicaSetService {
	if factory != nil {
		s.clientFactory = factory
	}
	return s
}

func (s *ReplicaSetService) ListReplicaSetsByApplication(
	ctx context.Context,
	applicationID uuid.UUID,
	organizationID uuid.UUID,
	namespace string,
) (*dto.ReplicaSetListResponse, error) {

	application, err := s.getOwnedApplication(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForApplication(application)
	if err != nil {
		return nil, err
	}

	ns := strings.TrimSpace(namespace)
	if ns == "" {
		ns = metav1.NamespaceAll
	}

	replicaSets, err := clientset.AppsV1().ReplicaSets(ns).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: s.applicationLabelSelector(applicationID, organizationID),
		},
	)
	if err != nil {
		return nil, mapReplicaSetError(err)
	}

	items := s.mapper.MapReplicaSetList(replicaSets.Items)

	return &dto.ReplicaSetListResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// ListReplicaSetsForCluster lists every ReplicaSet in the cluster (optionally
// filtered by namespace), reusing the same decrypt/clientFactory plumbing.
func (s *ReplicaSetService) ListReplicaSetsForCluster(
	ctx context.Context,
	clusterID uuid.UUID,
	organizationID uuid.UUID,
	namespace string,
) (*dto.ReplicaSetListResponse, error) {
	cluster, err := s.getOwnedCluster(clusterID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForCluster(cluster)
	if err != nil {
		return nil, err
	}

	ns := strings.TrimSpace(namespace)
	if ns == "" {
		ns = metav1.NamespaceAll
	}

	replicaSets, err := clientset.AppsV1().ReplicaSets(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, mapReplicaSetError(err)
	}

	items := s.mapper.MapReplicaSetList(replicaSets.Items)

	s.recordDiscovery(cluster.ID, organizationID)
	return &dto.ReplicaSetListResponse{
		Items: items,
		Total: len(items),
	}, nil
}

func (s *ReplicaSetService) GetReplicaSet(
	ctx context.Context,
	applicationID uuid.UUID,
	organizationID uuid.UUID,
	namespace string,
	name string,
) (*dto.ReplicaSet, error) {

	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)

	if namespace == "" {
		return nil, apperrors.ErrInvalidDeploymentNamespace
	}

	if name == "" {
		return nil, apperrors.ErrRuntimeDeploymentNotFound
	}

	application, err := s.getOwnedApplication(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForApplication(application)
	if err != nil {
		return nil, err
	}

	replicaSet, err := clientset.AppsV1().ReplicaSets(namespace).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, mapReplicaSetError(err)
	}

	if strings.TrimSpace(replicaSet.Labels["opspilot/application-id"]) != applicationID.String() ||
		strings.TrimSpace(replicaSet.Labels["opspilot/organization-id"]) != organizationID.String() {
		return nil, apperrors.ErrRuntimeDeploymentForbidden
	}

	response := s.mapper.MapReplicaSet(*replicaSet)
	return &response, nil
}

func (s *ReplicaSetService) applicationLabelSelector(
	applicationID uuid.UUID,
	organizationID uuid.UUID,
) string {
	return "opspilot/application-id=" + applicationID.String() +
		",opspilot/organization-id=" + organizationID.String()
}

func (s *ReplicaSetService) clientsetForApplication(
	application *models.Application,
) (kubernetes.Interface, error) {

	cluster, err := s.resolveProjectCluster(application.ProjectID, application.OrganizationID)
	if err != nil {
		return nil, err
	}

	if s.credentialCipher == nil {
		return nil, apperrors.ErrRuntimeDeploymentInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrRuntimeDeploymentInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapReplicaSetError(err)
	}

	return clientset, nil
}

func (s *ReplicaSetService) resolveProjectCluster(
	projectID uuid.UUID,
	organizationID uuid.UUID,
) (*models.Cluster, error) {

	cluster, err := s.clusterRepo.GetDefaultCluster(projectID, organizationID)
	if err == nil {
		return cluster, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	clusters, err := s.clusterRepo.FindByProject(projectID, organizationID)
	if err != nil {
		return nil, err
	}

	if len(clusters) == 0 {
		return nil, apperrors.ErrClusterNotFound
	}

	return &clusters[0], nil
}

func (s *ReplicaSetService) getOwnedApplication(
	id uuid.UUID,
	organizationID uuid.UUID,
) (*models.Application, error) {

	application, err := s.applicationRepo.GetApplication(id, organizationID)
	if err == nil {
		return application, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if existing, anyErr := s.applicationRepo.FindByIDAnyOrganization(id); anyErr == nil && existing != nil {
			return nil, apperrors.ErrApplicationForbidden
		}

		return nil, apperrors.ErrApplicationNotFound
	}

	return nil, err
}

func (s *ReplicaSetService) defaultClientFactory(
	kubeconfig []byte,
) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}

func (s *ReplicaSetService) getOwnedCluster(id uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	cluster, err := s.clusterRepo.FindByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	return cluster, nil
}

func (s *ReplicaSetService) clientsetForCluster(cluster *models.Cluster) (kubernetes.Interface, error) {
	if s.credentialCipher == nil {
		return nil, apperrors.ErrRuntimeDeploymentInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrRuntimeDeploymentInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapReplicaSetError(err)
	}

	return clientset, nil
}

func (s *ReplicaSetService) recordDiscovery(clusterID uuid.UUID, organizationID uuid.UUID) {
	_ = s.clusterRepo.UpdateDiscovery(clusterID, organizationID, time.Now())
}
