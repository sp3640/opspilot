package kubernetes

import (
	"encoding/json"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/monitoring"
)

const (
	resourceKindStatefulSet = "StatefulSet"
	resourceKindDaemonSet   = "DaemonSet"
)

type resourceMapperContext struct {
	projectID   uuid.UUID
	clusterID   uuid.UUID
	clusterName string
	createdBy   uint
}

func MapNamespace(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, namespace corev1.Namespace) models.Resource {
	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        constants.ResourceKindNamespace,
		Name:        namespace.Name,
		DisplayName: namespace.Name,
		ExternalID:  string(namespace.UID),
		Namespace:   namespace.Name,
		Status:      namespaceStatus(namespace),
		Health:      constants.ResourceHealthHealthy,
		Labels:      marshalMap(namespace.Labels),
		Annotations: marshalMap(namespace.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"phase":             string(namespace.Status.Phase),
			"resourceVersion":   namespace.ResourceVersion,
			"deletionTimestamp": formatDeletionTimestamp(namespace.DeletionTimestamp),
		}),
	})
}

func MapNode(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, node corev1.Node) models.Resource {
	ready := nodeReady(node)
	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        constants.ResourceKindNode,
		Name:        node.Name,
		DisplayName: node.Name,
		ExternalID:  string(node.UID),
		Status:      clusterScopedStatus(node.DeletionTimestamp != nil),
		Health:      nodeHealth(node),
		Labels:      marshalMap(node.Labels),
		Annotations: marshalMap(node.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"ready":             ready,
			"unschedulable":     node.Spec.Unschedulable,
			"kubeletVersion":    node.Status.NodeInfo.KubeletVersion,
			"kernelVersion":     node.Status.NodeInfo.KernelVersion,
			"operatingSystem":   node.Status.NodeInfo.OperatingSystem,
			"osImage":           node.Status.NodeInfo.OSImage,
			"architecture":      node.Status.NodeInfo.Architecture,
			"containerRuntime":  node.Status.NodeInfo.ContainerRuntimeVersion,
			"resourceVersion":   node.ResourceVersion,
			"deletionTimestamp": formatDeletionTimestamp(node.DeletionTimestamp),
		}),
	})
}

func MapDeployment(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, deployment appsv1.Deployment) models.Resource {
	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        constants.ResourceKindDeployment,
		Name:        deployment.Name,
		DisplayName: deployment.Name,
		ExternalID:  string(deployment.UID),
		Namespace:   deployment.Namespace,
		Status:      workloadStatus(deployment.DeletionTimestamp != nil, deployment.Generation != deployment.Status.ObservedGeneration),
		Health:      deploymentHealth(deployment),
		Labels:      marshalMap(deployment.Labels),
		Annotations: marshalMap(deployment.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"replicas":            deploymentStatusReplicas(deployment.Spec.Replicas),
			"readyReplicas":       deployment.Status.ReadyReplicas,
			"availableReplicas":   deployment.Status.AvailableReplicas,
			"updatedReplicas":     deployment.Status.UpdatedReplicas,
			"unavailableReplicas": deployment.Status.UnavailableReplicas,
			"resourceVersion":     deployment.ResourceVersion,
			"deletionTimestamp":   formatDeletionTimestamp(deployment.DeletionTimestamp),
		}),
	})
}

func MapStatefulSet(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, statefulSet appsv1.StatefulSet) models.Resource {
	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        resourceKindStatefulSet,
		Name:        statefulSet.Name,
		DisplayName: statefulSet.Name,
		ExternalID:  string(statefulSet.UID),
		Namespace:   statefulSet.Namespace,
		Status:      workloadStatus(statefulSet.DeletionTimestamp != nil, statefulSet.Generation != statefulSet.Status.ObservedGeneration),
		Health:      replicaHealth(statefulSet.Status.ReadyReplicas, deploymentStatusReplicas(statefulSet.Spec.Replicas)),
		Labels:      marshalMap(statefulSet.Labels),
		Annotations: marshalMap(statefulSet.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"replicas":          deploymentStatusReplicas(statefulSet.Spec.Replicas),
			"readyReplicas":     statefulSet.Status.ReadyReplicas,
			"currentReplicas":   statefulSet.Status.CurrentReplicas,
			"updatedReplicas":   statefulSet.Status.UpdatedReplicas,
			"serviceName":       statefulSet.Spec.ServiceName,
			"resourceVersion":   statefulSet.ResourceVersion,
			"deletionTimestamp": formatDeletionTimestamp(statefulSet.DeletionTimestamp),
		}),
	})
}

func MapDaemonSet(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, daemonSet appsv1.DaemonSet) models.Resource {
	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        resourceKindDaemonSet,
		Name:        daemonSet.Name,
		DisplayName: daemonSet.Name,
		ExternalID:  string(daemonSet.UID),
		Namespace:   daemonSet.Namespace,
		Status:      workloadStatus(daemonSet.DeletionTimestamp != nil, daemonSet.Generation != daemonSet.Status.ObservedGeneration),
		Health:      replicaHealth(daemonSet.Status.NumberReady, daemonSet.Status.DesiredNumberScheduled),
		Labels:      marshalMap(daemonSet.Labels),
		Annotations: marshalMap(daemonSet.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"desiredNumberScheduled": daemonSet.Status.DesiredNumberScheduled,
			"currentNumberScheduled": daemonSet.Status.CurrentNumberScheduled,
			"numberReady":            daemonSet.Status.NumberReady,
			"updatedNumberScheduled": daemonSet.Status.UpdatedNumberScheduled,
			"numberUnavailable":      daemonSet.Status.NumberUnavailable,
			"resourceVersion":        daemonSet.ResourceVersion,
			"deletionTimestamp":      formatDeletionTimestamp(daemonSet.DeletionTimestamp),
		}),
	})
}

func MapService(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, service corev1.Service) models.Resource {
	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        constants.ResourceKindService,
		Name:        service.Name,
		DisplayName: service.Name,
		ExternalID:  string(service.UID),
		Namespace:   service.Namespace,
		Status:      clusterScopedStatus(service.DeletionTimestamp != nil),
		Health:      serviceHealth(service),
		Labels:      marshalMap(service.Labels),
		Annotations: marshalMap(service.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"type":                string(service.Spec.Type),
			"clusterIP":           service.Spec.ClusterIP,
			"externalIPs":         service.Spec.ExternalIPs,
			"portCount":           len(service.Spec.Ports),
			"selectorCount":       len(service.Spec.Selector),
			"resourceVersion":     service.ResourceVersion,
			"deletionTimestamp":   formatDeletionTimestamp(service.DeletionTimestamp),
			"loadBalancerIngress": len(service.Status.LoadBalancer.Ingress),
		}),
	})
}

func MapPod(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, pod corev1.Pod) models.Resource {
	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        constants.ResourceKindPod,
		Name:        pod.Name,
		DisplayName: pod.Name,
		ExternalID:  string(pod.UID),
		Namespace:   pod.Namespace,
		Status:      podStatus(pod),
		Health:      podHealth(pod),
		Labels:      marshalMap(pod.Labels),
		Annotations: marshalMap(pod.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"phase":             string(pod.Status.Phase),
			"nodeName":          pod.Spec.NodeName,
			"podIP":             pod.Status.PodIP,
			"hostIP":            pod.Status.HostIP,
			"containerCount":    len(pod.Spec.Containers),
			"restartCount":      podRestartCount(pod),
			"resourceVersion":   pod.ResourceVersion,
			"deletionTimestamp": formatDeletionTimestamp(pod.DeletionTimestamp),
		}),
	})
}

func MapIngress(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, ingress networkingv1.Ingress) models.Resource {
	className := ""
	if ingress.Spec.IngressClassName != nil {
		className = *ingress.Spec.IngressClassName
	}

	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        constants.ResourceKindIngress,
		Name:        ingress.Name,
		DisplayName: ingress.Name,
		ExternalID:  string(ingress.UID),
		Namespace:   ingress.Namespace,
		Status:      clusterScopedStatus(ingress.DeletionTimestamp != nil),
		Health:      healthyUnlessDeleting(ingress.DeletionTimestamp != nil),
		Labels:      marshalMap(ingress.Labels),
		Annotations: marshalMap(ingress.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"className":         className,
			"ruleCount":         len(ingress.Spec.Rules),
			"tlsCount":          len(ingress.Spec.TLS),
			"loadBalancerCount": len(ingress.Status.LoadBalancer.Ingress),
			"resourceVersion":   ingress.ResourceVersion,
			"deletionTimestamp": formatDeletionTimestamp(ingress.DeletionTimestamp),
		}),
	})
}

func MapConfigMap(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, configMap corev1.ConfigMap) models.Resource {
	immutable := false
	if configMap.Immutable != nil {
		immutable = *configMap.Immutable
	}

	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        constants.ResourceKindConfigMap,
		Name:        configMap.Name,
		DisplayName: configMap.Name,
		ExternalID:  string(configMap.UID),
		Namespace:   configMap.Namespace,
		Status:      clusterScopedStatus(configMap.DeletionTimestamp != nil),
		Health:      healthyUnlessDeleting(configMap.DeletionTimestamp != nil),
		Labels:      marshalMap(configMap.Labels),
		Annotations: marshalMap(configMap.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"dataCount":         len(configMap.Data),
			"binaryDataCount":   len(configMap.BinaryData),
			"immutable":         immutable,
			"resourceVersion":   configMap.ResourceVersion,
			"deletionTimestamp": formatDeletionTimestamp(configMap.DeletionTimestamp),
		}),
	})
}

func MapPV(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, pv corev1.PersistentVolume) models.Resource {
	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        constants.ResourceKindPersistentVolume,
		Name:        pv.Name,
		DisplayName: pv.Name,
		ExternalID:  string(pv.UID),
		Status:      persistentVolumeStatus(pv),
		Health:      persistentVolumeHealth(pv.Status.Phase),
		Labels:      marshalMap(pv.Labels),
		Annotations: marshalMap(pv.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"phase":             string(pv.Status.Phase),
			"storageClassName":  pv.Spec.StorageClassName,
			"capacityStorage":   quantityString(pv.Spec.Capacity[corev1.ResourceStorage]),
			"accessModes":       accessModesToStrings(pv.Spec.AccessModes),
			"reclaimPolicy":     string(pv.Spec.PersistentVolumeReclaimPolicy),
			"volumeMode":        persistentVolumeMode(pv.Spec.VolumeMode),
			"claimNamespace":    claimRefNamespace(pv.Spec.ClaimRef),
			"claimName":         claimRefName(pv.Spec.ClaimRef),
			"resourceVersion":   pv.ResourceVersion,
			"deletionTimestamp": formatDeletionTimestamp(pv.DeletionTimestamp),
		}),
	})
}

func MapPVC(projectID, clusterID uuid.UUID, clusterName string, createdBy uint, pvc corev1.PersistentVolumeClaim) models.Resource {
	return newResource(resourceMapperContext{projectID: projectID, clusterID: clusterID, clusterName: clusterName, createdBy: createdBy}, models.Resource{
		Kind:        constants.ResourceKindPersistentVolumeClaim,
		Name:        pvc.Name,
		DisplayName: pvc.Name,
		ExternalID:  string(pvc.UID),
		Namespace:   pvc.Namespace,
		Status:      persistentVolumeClaimStatus(pvc),
		Health:      persistentVolumeHealth(pvc.Status.Phase),
		Labels:      marshalMap(pvc.Labels),
		Annotations: marshalMap(pvc.Annotations),
		Metadata: marshalMetadata(map[string]any{
			"phase":             string(pvc.Status.Phase),
			"storageClassName":  storageClassName(pvc.Spec.StorageClassName),
			"volumeName":        pvc.Spec.VolumeName,
			"requestedStorage":  quantityString(pvc.Spec.Resources.Requests[corev1.ResourceStorage]),
			"accessModes":       accessModesToStrings(pvc.Spec.AccessModes),
			"volumeMode":        persistentVolumeMode(pvc.Spec.VolumeMode),
			"resourceVersion":   pvc.ResourceVersion,
			"deletionTimestamp": formatDeletionTimestamp(pvc.DeletionTimestamp),
		}),
	})
}

func newResource(ctx resourceMapperContext, resource models.Resource) models.Resource {
	resource.ProjectID = ctx.projectID
	resource.Provider = string(monitoring.ProviderKubernetes)
	resource.Cluster = ctx.clusterID.String()
	resource.CreatedBy = ctx.createdBy

	metadata := unmarshalMetadata(resource.Metadata)
	if ctx.clusterName != "" {
		metadata["clusterName"] = ctx.clusterName
	}
	metadata["clusterId"] = ctx.clusterID.String()
	resource.Metadata = marshalMetadata(metadata)

	if len(resource.Labels) == 0 {
		resource.Labels = json.RawMessage(`{}`)
	}
	if len(resource.Annotations) == 0 {
		resource.Annotations = json.RawMessage(`{}`)
	}
	if len(resource.Metadata) == 0 {
		resource.Metadata = json.RawMessage(`{}`)
	}

	return resource
}

func deploymentStatusReplicas(replicas *int32) int32 {
	if replicas == nil {
		return 0
	}

	return *replicas
}

func namespaceStatus(namespace corev1.Namespace) string {
	if namespace.DeletionTimestamp != nil || namespace.Status.Phase == corev1.NamespaceTerminating {
		return constants.ResourceStatusTerminating
	}

	return constants.ResourceStatusActive
}

func clusterScopedStatus(isDeleting bool) string {
	if isDeleting {
		return constants.ResourceStatusTerminating
	}

	return constants.ResourceStatusActive
}

func workloadStatus(isDeleting, isUpdating bool) string {
	if isDeleting {
		return constants.ResourceStatusTerminating
	}
	if isUpdating {
		return constants.ResourceStatusUpdating
	}

	return constants.ResourceStatusActive
}

func podStatus(pod corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return constants.ResourceStatusTerminating
	}

	switch pod.Status.Phase {
	case corev1.PodPending:
		return constants.ResourceStatusPending
	case corev1.PodRunning, corev1.PodSucceeded:
		return constants.ResourceStatusActive
	case corev1.PodFailed:
		return constants.ResourceStatusUnknown
	default:
		return constants.ResourceStatusUnknown
	}
}

func podHealth(pod corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return constants.ResourceHealthDegraded
	}

	switch pod.Status.Phase {
	case corev1.PodRunning, corev1.PodSucceeded:
		return constants.ResourceHealthHealthy
	case corev1.PodPending:
		return constants.ResourceHealthDegraded
	case corev1.PodFailed:
		return constants.ResourceHealthUnhealthy
	default:
		return constants.ResourceHealthUnknown
	}
}

func nodeHealth(node corev1.Node) string {
	if node.DeletionTimestamp != nil {
		return constants.ResourceHealthDegraded
	}
	if nodeReady(node) {
		return constants.ResourceHealthHealthy
	}

	return constants.ResourceHealthUnhealthy
}

func nodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

func deploymentHealth(deployment appsv1.Deployment) string {
	if deployment.DeletionTimestamp != nil {
		return constants.ResourceHealthDegraded
	}
	if deployment.Status.AvailableReplicas == deploymentStatusReplicas(deployment.Spec.Replicas) {
		return constants.ResourceHealthHealthy
	}

	return constants.ResourceHealthDegraded
}

func serviceHealth(service corev1.Service) string {
	if service.DeletionTimestamp != nil {
		return constants.ResourceHealthDegraded
	}

	return constants.ResourceHealthHealthy
}

func healthyUnlessDeleting(isDeleting bool) string {
	if isDeleting {
		return constants.ResourceHealthDegraded
	}

	return constants.ResourceHealthHealthy
}

func replicaHealth(ready, desired int32) string {
	if desired == 0 {
		return constants.ResourceHealthUnknown
	}
	if ready == desired {
		return constants.ResourceHealthHealthy
	}

	return constants.ResourceHealthDegraded
}

func persistentVolumeStatus(pv corev1.PersistentVolume) string {
	if pv.DeletionTimestamp != nil {
		return constants.ResourceStatusTerminating
	}

	switch pv.Status.Phase {
	case corev1.VolumePending:
		return constants.ResourceStatusPending
	case corev1.VolumeBound, corev1.VolumeAvailable:
		return constants.ResourceStatusActive
	case corev1.VolumeReleased, corev1.VolumeFailed:
		return constants.ResourceStatusUnknown
	default:
		return constants.ResourceStatusUnknown
	}
}

func persistentVolumeClaimStatus(pvc corev1.PersistentVolumeClaim) string {
	if pvc.DeletionTimestamp != nil {
		return constants.ResourceStatusTerminating
	}

	switch pvc.Status.Phase {
	case corev1.ClaimPending:
		return constants.ResourceStatusPending
	case corev1.ClaimBound:
		return constants.ResourceStatusActive
	case corev1.ClaimLost:
		return constants.ResourceStatusUnknown
	default:
		return constants.ResourceStatusUnknown
	}
}

func persistentVolumeHealth[T ~string](phase T) string {
	value := strings.ToUpper(string(phase))
	switch value {
	case "BOUND", "AVAILABLE":
		return constants.ResourceHealthHealthy
	case "PENDING":
		return constants.ResourceHealthDegraded
	case "FAILED", "LOST":
		return constants.ResourceHealthUnhealthy
	default:
		return constants.ResourceHealthUnknown
	}
}

func marshalMap(values map[string]string) json.RawMessage {
	if len(values) == 0 {
		return json.RawMessage(`{}`)
	}

	payload, err := json.Marshal(values)
	if err != nil {
		return json.RawMessage(`{}`)
	}

	return payload
}

func marshalMetadata(values map[string]any) json.RawMessage {
	if len(values) == 0 {
		return json.RawMessage(`{}`)
	}

	payload, err := json.Marshal(values)
	if err != nil {
		return json.RawMessage(`{}`)
	}

	return payload
}

func unmarshalMetadata(payload json.RawMessage) map[string]any {
	if len(payload) == 0 {
		return map[string]any{}
	}

	values := make(map[string]any)
	if err := json.Unmarshal(payload, &values); err != nil {
		return map[string]any{}
	}

	return values
}

func formatDeletionTimestamp(timestamp interface{ String() string }) string {
	if timestamp == nil {
		return ""
	}

	return timestamp.String()
}

func podRestartCount(pod corev1.Pod) int32 {
	var count int32
	for _, status := range pod.Status.ContainerStatuses {
		count += status.RestartCount
	}

	return count
}

func accessModesToStrings(modes []corev1.PersistentVolumeAccessMode) []string {
	values := make([]string, 0, len(modes))
	for _, mode := range modes {
		values = append(values, string(mode))
	}

	return values
}

func quantityString(quantity apiresource.Quantity) string {
	value := quantity.String()
	if value == "0" {
		return ""
	}

	return value
}

func persistentVolumeMode(mode *corev1.PersistentVolumeMode) string {
	if mode == nil {
		return ""
	}

	return string(*mode)
}

func claimRefNamespace(reference *corev1.ObjectReference) string {
	if reference == nil {
		return ""
	}

	return reference.Namespace
}

func claimRefName(reference *corev1.ObjectReference) string {
	if reference == nil {
		return ""
	}

	return reference.Name
}

func storageClassName(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
