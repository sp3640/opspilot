package events

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/dto"
	intkube "github.com/sp3640/opspilot/backend/internal/integrations/kubernetes"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/repository"
	"github.com/sp3640/opspilot/backend/internal/security"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

type ClientsetFactory func(kubeconfig []byte) (kubernetes.Interface, error)

type EventService struct {
	applicationRepo  *repository.ApplicationRepository
	clusterRepo      *repository.ClusterRepository
	credentialCipher security.ClusterCredentialCipher
	mapper           *EventMapper
	clientFactory    ClientsetFactory
}

func NewEventService(applicationRepo *repository.ApplicationRepository, clusterRepo *repository.ClusterRepository, credentialCipher security.ClusterCredentialCipher) *EventService {
	service := &EventService{
		applicationRepo:  applicationRepo,
		clusterRepo:      clusterRepo,
		credentialCipher: credentialCipher,
		mapper:           NewEventMapper(NewReasonHelper()),
	}

	service.clientFactory = service.defaultClientFactory
	return service
}

func (s *EventService) WithClientsetFactory(factory ClientsetFactory) *EventService {
	if factory != nil {
		s.clientFactory = factory
	}

	return s
}

func (s *EventService) ListEventsByApplication(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string) (*dto.EventListResponse, error) {
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

	eventList, err := clientset.CoreV1().Events(ns).List(ctx, metav1.ListOptions{
		LabelSelector: s.applicationLabelSelector(applicationID, organizationID),
	})
	if err != nil {
		return nil, mapEventError(err)
	}

	items := make([]dto.EventResponse, 0, len(eventList.Items))
	for _, item := range eventList.Items {
		mapped := s.mapper.MapEvent(item, false)
		items = append(items, mapped.EventResponse)
	}

	return &dto.EventListResponse{Items: items, Total: len(items)}, nil
}

// ListEventsForPod returns the real Kubernetes events involving a single
// pod. Events are not labeled by opspilot (they are created by the kubelet
// and controllers, not by anything opspilot deploys), so ownership is
// verified by fetching the pod itself and checking its opspilot labels
// before listing events for it — the same check PodService.GetPod performs.
func (s *EventService) ListEventsForPod(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, podName string) (*dto.EventListResponse, error) {
	namespace = strings.TrimSpace(namespace)
	podName = strings.TrimSpace(podName)
	if namespace == "" {
		return nil, apperrors.ErrInvalidDeploymentNamespace
	}
	if podName == "" {
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

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, mapEventError(err)
	}
	if strings.TrimSpace(pod.Labels["opspilot/application-id"]) != applicationID.String() || strings.TrimSpace(pod.Labels["opspilot/organization-id"]) != organizationID.String() {
		return nil, apperrors.ErrEventForbidden
	}

	eventList, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{
			"involvedObject.kind":      "Pod",
			"involvedObject.name":      podName,
			"involvedObject.namespace": namespace,
		}).String(),
	})
	if err != nil {
		return nil, mapEventError(err)
	}

	items := make([]dto.EventResponse, 0, len(eventList.Items))
	for _, item := range eventList.Items {
		involved := item.InvolvedObject
		// Field selectors on Events are not honored by every backend (notably
		// the client-go fake clientset used in tests), so re-check here to
		// guarantee only this pod's events are ever returned.
		if involved.Kind != "Pod" || involved.Name != podName || involved.Namespace != namespace {
			continue
		}

		mapped := s.mapper.MapEvent(item, false)
		items = append(items, mapped.EventResponse)
	}

	return &dto.EventListResponse{Items: items, Total: len(items)}, nil
}

func (s *EventService) GetEvent(ctx context.Context, applicationID uuid.UUID, organizationID uuid.UUID, namespace string, name string) (*dto.EventDetailResponse, error) {
	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)
	if namespace == "" {
		return nil, apperrors.ErrInvalidDeploymentNamespace
	}
	if name == "" {
		return nil, apperrors.ErrEventNotFound
	}

	application, err := s.getOwnedApplication(applicationID, organizationID)
	if err != nil {
		return nil, err
	}

	clientset, err := s.clientsetForApplication(application)
	if err != nil {
		return nil, err
	}

	eventResource, err := clientset.CoreV1().Events(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, mapEventError(err)
	}

	if strings.TrimSpace(eventResource.Labels["opspilot/application-id"]) != applicationID.String() || strings.TrimSpace(eventResource.Labels["opspilot/organization-id"]) != organizationID.String() {
		return nil, apperrors.ErrEventForbidden
	}

	response := s.mapper.MapEvent(*eventResource, true)
	return &response, nil
}

func (s *EventService) applicationLabelSelector(applicationID uuid.UUID, organizationID uuid.UUID) string {
	return "opspilot/application-id=" + applicationID.String() + ",opspilot/organization-id=" + organizationID.String()
}

func (s *EventService) clientsetForApplication(application *models.Application) (kubernetes.Interface, error) {
	cluster, err := s.resolveProjectCluster(application.ProjectID, application.OrganizationID)
	if err != nil {
		return nil, err
	}

	if s.credentialCipher == nil {
		return nil, apperrors.ErrEventInvalidKubeconfig
	}

	kubeconfig, err := s.credentialCipher.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, apperrors.ErrEventInvalidKubeconfig
	}

	clientset, err := s.clientFactory([]byte(kubeconfig))
	if err != nil {
		return nil, mapEventError(err)
	}

	return clientset, nil
}

func (s *EventService) resolveProjectCluster(projectID uuid.UUID, organizationID uuid.UUID) (*models.Cluster, error) {
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

func (s *EventService) getOwnedApplication(id uuid.UUID, organizationID uuid.UUID) (*models.Application, error) {
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

func (s *EventService) defaultClientFactory(kubeconfig []byte) (kubernetes.Interface, error) {
	return intkube.NewClient(kubeconfig).Clientset()
}
