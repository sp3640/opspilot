package ingresses

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

type IngressService struct {
	applicationRepo  *repository.ApplicationRepository
	clusterRepo      *repository.ClusterRepository
	credentialCipher security.ClusterCredentialCipher
	mapper           *IngressMapper
	clientFactory    ClientsetFactory
}

func NewIngressService(applicationRepo *repository.ApplicationRepository, clusterRepo *repository.ClusterRepository, credentialCipher security.ClusterCredentialCipher) *IngressService {
	service := &IngressService{
		applicationRepo:  applicationRepo,
		clusterRepo:      clusterRepo,
		credentialCipher: credentialCipher,
		mapper:           NewIngressMapper(NewIngressStatusHelper()),
	}

	service.clientFactory = service.defaultClientFactory
	return service
}

func (s *IngressService) WithClientsetFactory(factory ClientsetFactory) *IngressService {
	if factory != nil {
		s.clientFactory = factory
	}

	return s
}

func (s *IngressService) ListIngressesByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.IngressListResponse, error) {
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

	ingressList, err := clientset.NetworkingV1().Ingresses(ns).List(ctx, metav1.ListOptions{
		LabelSelector: s.applicationLabelSelector(applicationID, organizationID),
	})
	if err != nil {
		return nil, mapIngressError(err)
	}

	items := make([]dto.IngressResponse, 0, len(ingressList.Items))
	for _, item := range ingressList.Items {
		items = append(items, s.mapper.MapIngress(item, false))
	}

	return &dto.IngressListResponse{Items: items, Total: len(items)}, nil
}

// ListIngressesForCluster lists every Ingress in the cluster (optionally
// filtered by namespace), reusing the same decrypt/clientFactory plumbing.
func (s *IngressService) ListIngressesForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.IngressListResponse, error) {
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

	ingressList, err := clientset.NetworkingV1().Ingresses(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, mapIngressError(err)
	}

	items := make([]dto.IngressResponse, 0, len(ingressList.Items))
	for _, item := range ingressList.Items {
		items = append(items, s.mapper.MapIngress(item, false))
	}

	s.recordDiscovery(cluster.ID, organizationID)
	return &dto.IngressListResponse{Items: items, Total: len(items)}, nil
}

func (s *IngressService) GetIngress(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.IngressResponse, error) {
	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)
	if namespace == "" {
		return nil, apperrors.ErrInvalidDeploymentNamespace
	}
	if name == "" {
		return nil, apperrors.ErrIngressNotFound
	}

	application, err := s.getOwnedApplication(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForApplication(application)
	if err != nil {
		return nil, err
	}

	ingressResource, err := clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, mapIngressError(err)
	}

	if strings.TrimSpace(ingressResource.Labels["opspilot/application-id"]) != applicationID.String() || strings.TrimSpace(ingressResource.Labels["opspilot/organization-id"]) != organizationID.String() {
		return nil, apperrors.ErrIngressForbidden
	}

	response := s.mapper.MapIngress(*ingressResource, true)
	return &response, nil
}

func (s *IngressService) applicationLabelSelector(applicationID uuid.UUID, organizationID uuid.UUID) string {
	return "opspilot/application-id=" + applicationID.String() + ",opspilot/organization-id=" + organizationID.String()
}

func (s *IngressService) clientsetForApplication(application *models.Application) (kubernetes.Interface, error) {
	cluster, err := s.resolveProjectCluster(application.ProjectID, application.OrganizationID)
	if err != nil {
		return nil, err
	}

	if s.credentialCipher == nil {
		return nil, apperrors.ErrIngressInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrIngressInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapIngressError(err)
	}

	return clientset, nil
}

func (s *IngressService) resolveProjectCluster(projectID uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
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

func (s *IngressService) getOwnedApplication(id uuid.UUID, organizationID uuid.UUID) (*models.Application, error) {
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

func (s *IngressService) defaultClientFactory(kubeconfig []byte) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}

func (s *IngressService) getOwnedCluster(id uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	cluster, err := s.clusterRepo.FindByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	return cluster, nil
}

func (s *IngressService) clientsetForCluster(cluster *models.Cluster) (kubernetes.Interface, error) {
	if s.credentialCipher == nil {
		return nil, apperrors.ErrIngressInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrIngressInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapIngressError(err)
	}

	return clientset, nil
}

func (s *IngressService) recordDiscovery(clusterID uuid.UUID, organizationID uuid.UUID) {
	_ = s.clusterRepo.UpdateDiscovery(clusterID, organizationID, time.Now())
}
