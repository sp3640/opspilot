package services

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

type ServiceService struct {
	applicationRepo  *repository.ApplicationRepository
	clusterRepo      *repository.ClusterRepository
	credentialCipher security.ClusterCredentialCipher
	mapper           *ServiceMapper
	clientFactory    ClientsetFactory
}

func NewServiceService(applicationRepo *repository.ApplicationRepository, clusterRepo *repository.ClusterRepository, credentialCipher security.ClusterCredentialCipher) *ServiceService {
	service := &ServiceService{
		applicationRepo:  applicationRepo,
		clusterRepo:      clusterRepo,
		credentialCipher: credentialCipher,
		mapper:           NewServiceMapper(NewServiceStatusHelper()),
	}

	service.clientFactory = service.defaultClientFactory
	return service
}

func (s *ServiceService) WithClientsetFactory(factory ClientsetFactory) *ServiceService {
	if factory != nil {
		s.clientFactory = factory
	}

	return s
}

func (s *ServiceService) ListServicesByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.ServiceListResponse, error) {
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

	serviceList, err := clientset.CoreV1().Services(ns).List(ctx, metav1.ListOptions{
		LabelSelector: s.applicationLabelSelector(applicationID, organizationID),
	})
	if err != nil {
		return nil, mapServiceError(err)
	}

	items := make([]dto.ServiceResponse, 0, len(serviceList.Items))
	for _, item := range serviceList.Items {
		items = append(items, s.mapper.MapService(item, false))
	}

	return &dto.ServiceListResponse{Items: items, Total: len(items)}, nil
}

// ListServicesForCluster lists every Service in the cluster (optionally
// filtered by namespace), reusing the same decrypt/clientFactory plumbing.
func (s *ServiceService) ListServicesForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.ServiceListResponse, error) {
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

	serviceList, err := clientset.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, mapServiceError(err)
	}

	items := make([]dto.ServiceResponse, 0, len(serviceList.Items))
	for _, item := range serviceList.Items {
		items = append(items, s.mapper.MapService(item, false))
	}

	s.recordDiscovery(cluster.ID, organizationID)
	return &dto.ServiceListResponse{Items: items, Total: len(items)}, nil
}

func (s *ServiceService) GetService(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.ServiceResponse, error) {
	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)
	if namespace == "" {
		return nil, apperrors.ErrInvalidDeploymentNamespace
	}
	if name == "" {
		return nil, apperrors.ErrServiceNotFound
	}

	application, err := s.getOwnedApplication(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForApplication(application)
	if err != nil {
		return nil, err
	}

	serviceResource, err := clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, mapServiceError(err)
	}

	if strings.TrimSpace(serviceResource.Labels["opspilot/application-id"]) != applicationID.String() || strings.TrimSpace(serviceResource.Labels["opspilot/organization-id"]) != organizationID.String() {
		return nil, apperrors.ErrServiceForbidden
	}

	response := s.mapper.MapService(*serviceResource, true)
	return &response, nil
}

func (s *ServiceService) applicationLabelSelector(applicationID uuid.UUID, organizationID uuid.UUID) string {
	return "opspilot/application-id=" + applicationID.String() + ",opspilot/organization-id=" + organizationID.String()
}

func (s *ServiceService) clientsetForApplication(application *models.Application) (kubernetes.Interface, error) {
	cluster, err := s.resolveProjectCluster(application.ProjectID, application.OrganizationID)
	if err != nil {
		return nil, err
	}

	if s.credentialCipher == nil {
		return nil, apperrors.ErrServiceInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrServiceInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapServiceError(err)
	}

	return clientset, nil
}

func (s *ServiceService) resolveProjectCluster(projectID uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
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

func (s *ServiceService) getOwnedApplication(id uuid.UUID, organizationID uuid.UUID) (*models.Application, error) {
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

func (s *ServiceService) defaultClientFactory(kubeconfig []byte) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}

func (s *ServiceService) getOwnedCluster(id uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	cluster, err := s.clusterRepo.FindByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	return cluster, nil
}

func (s *ServiceService) clientsetForCluster(cluster *models.Cluster) (kubernetes.Interface, error) {
	if s.credentialCipher == nil {
		return nil, apperrors.ErrServiceInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrServiceInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapServiceError(err)
	}

	return clientset, nil
}

func (s *ServiceService) recordDiscovery(clusterID uuid.UUID, organizationID uuid.UUID) {
	_ = s.clusterRepo.UpdateDiscovery(clusterID, organizationID, time.Now())
}
