package nodes

import (
	"context"
	"errors"

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
	"time"
)

// Nodes are a cluster-level concept only — there is no per-application
// scoping (unlike Pods/Services/etc, which are also labeled and listed per
// application). This service only ever lists Nodes for a whole cluster.
type ClientsetFactory func(kubeconfig []byte) (kubernetes.Interface, error)

type NodeService struct {
	clusterRepo      *repository.ClusterRepository
	credentialCipher security.ClusterCredentialCipher
	mapper           *NodeMapper
	clientFactory    ClientsetFactory
}

func NewNodeService(clusterRepo *repository.ClusterRepository, credentialCipher security.ClusterCredentialCipher) *NodeService {
	service := &NodeService{
		clusterRepo:      clusterRepo,
		credentialCipher: credentialCipher,
		mapper:           NewNodeMapper(),
	}

	service.clientFactory = service.defaultClientFactory
	return service
}

func (s *NodeService) WithClientsetFactory(factory ClientsetFactory) *NodeService {
	if factory != nil {
		s.clientFactory = factory
	}

	return s
}

// ListNodesForCluster lists every Node in the cluster, reusing the same
// decrypt/clientFactory plumbing as every other Kubernetes runtime service —
// no separate Kubernetes client is created.
func (s *NodeService) ListNodesForCluster(ctx context.Context, clusterID uuid.UUID, organizationID uuid.UUID) (*dto.NodeListResponse, error) {
	cluster, err := s.getOwnedCluster(clusterID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForCluster(cluster)
	if err != nil {
		return nil, err
	}

	nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, mapNodeError(err)
	}

	items := s.mapper.MapNodeList(nodeList.Items)

	s.recordDiscovery(cluster.ID, organizationID)
	return &dto.NodeListResponse{Items: items, Total: len(items)}, nil
}

func (s *NodeService) getOwnedCluster(id uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
	cluster, err := s.clusterRepo.FindByID(id, organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrClusterNotFound
		}
		return nil, err
	}

	return cluster, nil
}

func (s *NodeService) clientsetForCluster(cluster *models.Cluster) (kubernetes.Interface, error) {
	if s.credentialCipher == nil {
		return nil, apperrors.ErrNodeInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrNodeInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapNodeError(err)
	}

	return clientset, nil
}

func (s *NodeService) recordDiscovery(clusterID uuid.UUID, organizationID uuid.UUID) {
	_ = s.clusterRepo.UpdateDiscovery(clusterID, organizationID, time.Now())
}

func (s *NodeService) defaultClientFactory(kubeconfig []byte) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}
