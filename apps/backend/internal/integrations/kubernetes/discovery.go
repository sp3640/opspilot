package kubernetes

import (
	"context"
	"errors"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/discovery"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/monitoring"
)

type KubernetesDiscoveryProvider struct {
	projectID   uuid.UUID
	clusterID   uuid.UUID
	clusterName string
	createdBy   uint
	client      Client
	validator   Validator
}

func NewDiscoveryProvider(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, client Client, validator Validator) discovery.DiscoveryProvider {
	if validator == nil && client != nil {
		validator = NewValidator(client)
	}

	return &KubernetesDiscoveryProvider{
		projectID:   projectID,
		clusterID:   clusterID,
		clusterName: clusterName,
		createdBy:   createdBy,
		client:      client,
		validator:   validator,
	}
}

func (p *KubernetesDiscoveryProvider) Name() string {
	if p.clusterName != "" {
		return p.clusterName
	}

	return p.clusterID.String()
}

func (p *KubernetesDiscoveryProvider) Provider() monitoring.ProviderType {
	return monitoring.ProviderKubernetes
}

func (p *KubernetesDiscoveryProvider) Validate(ctx context.Context) error {
	if p.validator == nil {
		return &ErrConnectionFailed{Err: errors.New("kubernetes validator is required")}
	}

	_, err := p.validator.ValidateConnection(ctx)
	return err
}

func (p *KubernetesDiscoveryProvider) Discover(ctx context.Context) (*discovery.DiscoveryResult, error) {
	if p.client == nil {
		return nil, &ErrConnectionFailed{Err: errors.New("kubernetes client is required")}
	}

	startedAt := time.Now()

	clientset, err := p.client.Clientset()
	if err != nil {
		return nil, err
	}

	resources := make([]models.Resource, 0)

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapNamespaces(p.projectID, p.clusterID, p.clusterName, p.createdBy, namespaces.Items)...)

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapNodes(p.projectID, p.clusterID, p.clusterName, p.createdBy, nodes.Items)...)

	deployments, err := clientset.AppsV1().Deployments(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapDeployments(p.projectID, p.clusterID, p.clusterName, p.createdBy, deployments.Items)...)

	statefulSets, err := clientset.AppsV1().StatefulSets(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapStatefulSets(p.projectID, p.clusterID, p.clusterName, p.createdBy, statefulSets.Items)...)

	daemonSets, err := clientset.AppsV1().DaemonSets(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapDaemonSets(p.projectID, p.clusterID, p.clusterName, p.createdBy, daemonSets.Items)...)

	services, err := clientset.CoreV1().Services(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapServices(p.projectID, p.clusterID, p.clusterName, p.createdBy, services.Items)...)

	pods, err := clientset.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapPods(p.projectID, p.clusterID, p.clusterName, p.createdBy, pods.Items)...)

	ingresses, err := clientset.NetworkingV1().Ingresses(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapIngresses(p.projectID, p.clusterID, p.clusterName, p.createdBy, ingresses.Items)...)

	configMaps, err := clientset.CoreV1().ConfigMaps(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapConfigMaps(p.projectID, p.clusterID, p.clusterName, p.createdBy, configMaps.Items)...)

	persistentVolumes, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapPersistentVolumes(p.projectID, p.clusterID, p.clusterName, p.createdBy, persistentVolumes.Items)...)

	persistentVolumeClaims, err := clientset.CoreV1().PersistentVolumeClaims(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	resources = append(resources, mapPersistentVolumeClaims(p.projectID, p.clusterID, p.clusterName, p.createdBy, persistentVolumeClaims.Items)...)

	return &discovery.DiscoveryResult{
		Resources: resources,
		Warnings:  []string{},
		Errors:    []string{},
		Duration:  time.Since(startedAt),
	}, nil
}

func mapNamespaces(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []corev1.Namespace) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapNamespace(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapNodes(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []corev1.Node) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapNode(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapDeployments(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []appsv1.Deployment) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapDeployment(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapStatefulSets(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []appsv1.StatefulSet) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapStatefulSet(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapDaemonSets(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []appsv1.DaemonSet) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapDaemonSet(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapServices(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []corev1.Service) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapService(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapPods(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []corev1.Pod) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapPod(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapIngresses(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []networkingv1.Ingress) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapIngress(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapConfigMaps(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []corev1.ConfigMap) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapConfigMap(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapPersistentVolumes(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []corev1.PersistentVolume) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapPV(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}

func mapPersistentVolumeClaims(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, items []corev1.PersistentVolumeClaim) []models.Resource {
	resources := make([]models.Resource, 0, len(items))
	for _, item := range items {
		resources = append(resources, MapPVC(projectID, clusterID, clusterName, createdBy, item))
	}

	return resources
}
