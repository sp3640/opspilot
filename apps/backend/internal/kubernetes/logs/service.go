package logs

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/security"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClientsetFactory func(kubeconfig []byte) (kubernetes.Interface, error)

type LogService struct {
	applicationRepo  *repository.ApplicationRepository
	clusterRepo      *repository.ClusterRepository
	credentialCipher security.ClusterCredentialCipher
	mapper           *LogMapper
	clientFactory    ClientsetFactory
}

func NewLogService(applicationRepo *repository.ApplicationRepository, clusterRepo *repository.ClusterRepository, credentialCipher security.ClusterCredentialCipher) *LogService {
	service := &LogService{
		applicationRepo:  applicationRepo,
		clusterRepo:      clusterRepo,
		credentialCipher: credentialCipher,
		mapper:           NewLogMapper(),
	}

	service.clientFactory = service.defaultClientFactory
	return service
}

func (s *LogService) WithClientsetFactory(factory ClientsetFactory) *LogService {
	if factory != nil {
		s.clientFactory = factory
	}

	return s
}

func (s *LogService) GetPodLogs(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, podName string, container string, tailLines *int64, sinceSeconds *int64, timestamps bool) (*dto.PodLogResponse, error) {
	namespace = strings.TrimSpace(namespace)
	podName = strings.TrimSpace(podName)
	container = strings.TrimSpace(container)

	if namespace == "" {
		return nil, apperrors.ErrInvalidDeploymentNamespace
	}
	if podName == "" {
		return nil, apperrors.ErrLogPodNotFound
	}

	application, err := s.getOwnedApplication(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForApplication(application)
	if err != nil {
		return nil, err
	}

	pod, err := s.ValidatePodOwnership(ctx, clientset, applicationID, organizationID, namespace, podName)
	if err != nil {
		return nil, err
	}

	if container != "" && !podContainsContainer(pod, container) {
		return nil, apperrors.ErrLogContainerNotFound
	}
	if container == "" {
		container = defaultContainerName(pod)
	}

	logOptions := &corev1.PodLogOptions{Container: container, TailLines: tailLines, SinceSeconds: sinceSeconds, Timestamps: timestamps}
	stream, err := clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions).Stream(ctx)
	if err != nil {
		return nil, mapLogError(err)
	}
	defer stream.Close()

	logBytes, err := io.ReadAll(stream)
	if err != nil {
		return nil, mapLogError(err)
	}

	return s.mapper.MapPodLogs(podName, container, namespace, string(logBytes)), nil
}

func (s *LogService) ValidatePodOwnership(ctx context.Context, clientset kubernetes.Interface, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, podName string) (*corev1.Pod, error) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, mapLogError(err)
	}

	if strings.TrimSpace(pod.Labels["opspilot/application-id"]) != applicationID.String() || strings.TrimSpace(pod.Labels["opspilot/organization-id"]) != organizationID.String() {
		return nil, apperrors.ErrLogForbidden
	}

	return pod, nil
}

func podContainsContainer(pod *corev1.Pod, container string) bool {
	for _, item := range pod.Spec.Containers {
		if strings.TrimSpace(item.Name) == container {
			return true
		}
	}
	for _, item := range pod.Spec.InitContainers {
		if strings.TrimSpace(item.Name) == container {
			return true
		}
	}

	return false
}

func defaultContainerName(pod *corev1.Pod) string {
	if len(pod.Spec.Containers) > 0 {
		return strings.TrimSpace(pod.Spec.Containers[0].Name)
	}
	if len(pod.Spec.InitContainers) > 0 {
		return strings.TrimSpace(pod.Spec.InitContainers[0].Name)
	}

	return ""
}

func (s *LogService) clientsetForApplication(application *models.Application) (kubernetes.Interface, error) {
	cluster, err := s.resolveProjectCluster(application.ProjectID, application.OrganizationID)
	if err != nil {
		return nil, err
	}

	if s.credentialCipher == nil {
		return nil, apperrors.ErrLogInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrLogInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapLogError(err)
	}

	return clientset, nil
}

func (s *LogService) resolveProjectCluster(projectID uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
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

func (s *LogService) getOwnedApplication(id uuid.UUID, organizationID uuid.UUID) (*models.Application, error) {
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

func (s *LogService) defaultClientFactory(kubeconfig []byte) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}
