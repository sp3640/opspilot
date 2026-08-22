package deployments

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

type DeploymentRuntimeService struct {
	applicationRepo  *repository.ApplicationRepository
	clusterRepo      *repository.ClusterRepository
	credentialCipher security.ClusterCredentialCipher
	mapper           *DeploymentRuntimeMapper
	clientFactory    ClientsetFactory
}

func NewDeploymentRuntimeService(applicationRepo *repository.ApplicationRepository, clusterRepo *repository.ClusterRepository, credentialCipher security.ClusterCredentialCipher) *DeploymentRuntimeService {
	service := &DeploymentRuntimeService{
		applicationRepo:  applicationRepo,
		clusterRepo:      clusterRepo,
		credentialCipher: credentialCipher,
		mapper:           NewDeploymentRuntimeMapper(),
	}

	service.clientFactory = service.defaultClientFactory
	return service
}

func (s *DeploymentRuntimeService) WithClientsetFactory(factory ClientsetFactory) *DeploymentRuntimeService {
	if factory != nil {
		s.clientFactory = factory
	}

	return s
}

func (s *DeploymentRuntimeService) ListDeploymentsByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.DeploymentRuntimeListResponse, error) {
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

	deploymentList, err := clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{LabelSelector: s.applicationLabelSelector(applicationID, organizationID)})
	if err != nil {
		return nil, mapRuntimeDeploymentError(err)
	}

	items := make([]dto.DeploymentRuntimeResponse, 0, len(deploymentList.Items))
	for _, item := range deploymentList.Items {
		items = append(items, s.mapper.MapDeployment(item, false))
	}

	return &dto.DeploymentRuntimeListResponse{Items: items, Total: len(items)}, nil
}

// ListDeploymentsForCluster lists every Deployment in the cluster (optionally
// filtered by namespace), reusing the same decrypt/clientFactory plumbing.
func (s *DeploymentRuntimeService) ListDeploymentsForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.DeploymentRuntimeListResponse, error) {
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

	deploymentList, err := clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, mapRuntimeDeploymentError(err)
	}

	items := make([]dto.DeploymentRuntimeResponse, 0, len(deploymentList.Items))
	for _, item := range deploymentList.Items {
		items = append(items, s.mapper.MapDeployment(item, false))
	}

	s.recordDiscovery(cluster.ID, organizationID)
	return &dto.DeploymentRuntimeListResponse{Items: items, Total: len(items)}, nil
}

func (s *DeploymentRuntimeService) GetDeployment(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.DeploymentRuntimeDetailResponse, error) {
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

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, mapRuntimeDeploymentError(err)
	}

	if strings.TrimSpace(deployment.Labels["opspilot/application-id"]) != applicationID.String() || strings.TrimSpace(deployment.Labels["opspilot/organization-id"]) != organizationID.String() {
		return nil, apperrors.ErrRuntimeDeploymentForbidden
	}

	response := s.mapper.MapDeploymentDetail(*deployment)
	return &response, nil
}

func (s *DeploymentRuntimeService) applicationLabelSelector(applicationID uuid.UUID, organizationID uuid.UUID) string {
	return "opspilot/application-id=" + applicationID.String() + ",opspilot/organization-id=" + organizationID.String()
}

func (s *DeploymentRuntimeService) clientsetForApplication(application *models.Application) (kubernetes.Interface, error) {
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
		return nil, mapRuntimeDeploymentError(err)
	}

	return clientset, nil
}

func (s *DeploymentRuntimeService) resolveProjectCluster(projectID uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
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

func (s *DeploymentRuntimeService) getOwnedApplication(id uuid.UUID, organizationID uuid.UUID) (*models.Application, error) {
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

func (s *DeploymentRuntimeService) defaultClientFactory(kubeconfig []byte) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}

func (s *DeploymentRuntimeService) getOwnedCluster(id uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	cluster, err := s.clusterRepo.FindByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	return cluster, nil
}

func (s *DeploymentRuntimeService) clientsetForCluster(cluster *models.Cluster) (kubernetes.Interface, error) {
	if s.credentialCipher == nil {
		return nil, apperrors.ErrRuntimeDeploymentInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrRuntimeDeploymentInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapRuntimeDeploymentError(err)
	}

	return clientset, nil
}

func (s *DeploymentRuntimeService) recordDiscovery(clusterID uuid.UUID, organizationID uuid.UUID) {
	_ = s.clusterRepo.UpdateDiscovery(clusterID, organizationID, time.Now())
}
