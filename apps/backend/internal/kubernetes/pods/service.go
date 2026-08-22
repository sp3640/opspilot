package pods

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

type PodService struct {
	applicationRepo  *repository.ApplicationRepository
	clusterRepo      *repository.ClusterRepository
	credentialCipher security.ClusterCredentialCipher
	mapper           *PodMapper
	clientFactory    ClientsetFactory
}

func NewPodService(applicationRepo *repository.ApplicationRepository, clusterRepo *repository.ClusterRepository, credentialCipher security.ClusterCredentialCipher) *PodService {
	service := &PodService{
		applicationRepo:  applicationRepo,
		clusterRepo:      clusterRepo,
		credentialCipher: credentialCipher,
		mapper:           NewPodMapper(NewPodStatusHelper()),
	}

	service.clientFactory = service.defaultClientFactory
	return service
}

func (s *PodService) WithClientsetFactory(factory ClientsetFactory) *PodService {
	if factory != nil {
		s.clientFactory = factory
	}

	return s
}

func (s *PodService) ListPodsByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.PodListResponse, error) {
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

	pods, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: s.applicationLabelSelector(applicationID, organizationID),
	})
	if err != nil {
		return nil, mapPodError(err)
	}

	items := s.mapper.MapPodList(pods.Items)
	return &dto.PodListResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// ListPodsForCluster lists every Pod in the cluster (optionally filtered by
// namespace), reusing the same decrypt/clientFactory plumbing.
func (s *PodService) ListPodsForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.PodListResponse, error) {
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

	pods, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, mapPodError(err)
	}

	items := s.mapper.MapPodList(pods.Items)

	s.recordDiscovery(cluster.ID, organizationID)
	return &dto.PodListResponse{
		Items: items,
		Total: len(items),
	}, nil
}

func (s *PodService) GetPod(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.PodResponse, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, apperrors.ErrInvalidDeploymentNamespace
	}
	if strings.TrimSpace(name) == "" {
		return nil, apperrors.ErrPodNotFound
	}

	application, err := s.getOwnedApplication(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForApplication(application)
	if err != nil {
		return nil, err
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, strings.TrimSpace(name), metav1.GetOptions{})
	if err != nil {
		return nil, mapPodError(err)
	}

	if strings.TrimSpace(pod.Labels["opspilot/application-id"]) != applicationID.String() || strings.TrimSpace(pod.Labels["opspilot/organization-id"]) != organizationID.String() {
		return nil, apperrors.ErrPodForbidden
	}

	response := s.mapper.MapPod(*pod)
	return &response, nil
}

func (s *PodService) applicationLabelSelector(applicationID uuid.UUID, organizationID uuid.UUID) string {
	return "opspilot/application-id=" + applicationID.String() + ",opspilot/organization-id=" + organizationID.String()
}

func (s *PodService) clientsetForApplication(application *models.Application) (kubernetes.Interface, error) {
	cluster, err := s.resolveProjectCluster(application.ProjectID, application.OrganizationID)
	if err != nil {
		return nil, err
	}

	if s.credentialCipher == nil {
		return nil, apperrors.ErrPodInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrPodInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapPodError(err)
	}

	return clientset, nil
}

func (s *PodService) resolveProjectCluster(projectID uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
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

func (s *PodService) getOwnedApplication(id uuid.UUID, organizationID uuid.UUID) (*models.Application, error) {
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

func (s *PodService) defaultClientFactory(kubeconfig []byte) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}

func (s *PodService) getOwnedCluster(id uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	cluster, err := s.clusterRepo.FindByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	return cluster, nil
}

func (s *PodService) clientsetForCluster(cluster *models.Cluster) (kubernetes.Interface, error) {
	if s.credentialCipher == nil {
		return nil, apperrors.ErrPodInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrPodInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapPodError(err)
	}

	return clientset, nil
}

func (s *PodService) recordDiscovery(clusterID uuid.UUID, organizationID uuid.UUID) {
	_ = s.clusterRepo.UpdateDiscovery(clusterID, organizationID, time.Now())
}
