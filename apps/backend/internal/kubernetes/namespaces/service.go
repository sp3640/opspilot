package namespaces

import (
	"context"
	"errors"
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

// Namespaces are a cluster-level concept only, like Nodes — there is no
// per-application scoping.
type ClientsetFactory func(kubeconfig []byte) (kubernetes.Interface, error)

type NamespaceService struct {
	clusterRepo      *repository.ClusterRepository
	credentialCipher security.ClusterCredentialCipher
	mapper           *NamespaceMapper
	clientFactory    ClientsetFactory
}

func NewNamespaceService(clusterRepo *repository.ClusterRepository, credentialCipher security.ClusterCredentialCipher) *NamespaceService {
	service := &NamespaceService{
		clusterRepo:      clusterRepo,
		credentialCipher: credentialCipher,
		mapper:           NewNamespaceMapper(),
	}

	service.clientFactory = service.defaultClientFactory
	return service
}

func (s *NamespaceService) WithClientsetFactory(factory ClientsetFactory) *NamespaceService {
	if factory != nil {
		s.clientFactory = factory
	}

	return s
}

// ListNamespacesForCluster lists every Namespace in the cluster, reusing the
// same decrypt/clientFactory plumbing as every other Kubernetes runtime
// service — no separate Kubernetes client is created.
func (s *NamespaceService) ListNamespacesForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID) (*dto.NamespaceListResponse, error) {
	cluster, err := s.getOwnedCluster(clusterID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForCluster(cluster)
	if err != nil {
		return nil, err
	}

	namespaceList, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, mapNamespaceError(err)
	}

	items := s.mapper.MapNamespaceList(namespaceList.Items)

	s.recordDiscovery(cluster.ID, organizationID)
	return &dto.NamespaceListResponse{Items: items, Total: len(items)}, nil
}

func (s *NamespaceService) getOwnedCluster(id uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	cluster, err := s.clusterRepo.FindByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	return cluster, nil
}

func (s *NamespaceService) clientsetForCluster(cluster *models.Cluster) (kubernetes.Interface, error) {
	if s.credentialCipher == nil {
		return nil, apperrors.ErrNamespaceInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrNamespaceInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapNamespaceError(err)
	}

	return clientset, nil
}

func (s *NamespaceService) recordDiscovery(clusterID uuid.UUID, organizationID uuid.UUID) {
	_ = s.clusterRepo.UpdateDiscovery(clusterID, organizationID, time.Now())
}

func (s *NamespaceService) defaultClientFactory(kubeconfig []byte) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}
