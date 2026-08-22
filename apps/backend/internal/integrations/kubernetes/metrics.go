package kubernetes

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/sp3640/opspilot/backend/internal/metrics"
)

const (
	usageSourceMetricsServer     = "metrics-server"
	usageSourceRequestedEstimate = "requested-capacity"
)

// MetricsClientsetFactory builds a metrics.k8s.io client from a REST config.
// Kept swappable (mirroring ClientsetFactory elsewhere) so tests can inject a
// fake metrics clientset without a real metrics-server, and so production
// code can gracefully treat "metrics-server not installed" as a fallback
// rather than a hard failure.
type MetricsClientsetFactory func(cfg *rest.Config) (metricsclientset.Interface, error)

type MetricsCollector struct {
	name                 string
	client               Client
	metricsClientFactory MetricsClientsetFactory
}

func NewMetricsCollector(name string, client Client) *MetricsCollector {
	return &MetricsCollector{
		name:                 name,
		client:               client,
		metricsClientFactory: defaultMetricsClientsetFactory,
	}
}

func (c *MetricsCollector) WithMetricsClientsetFactory(factory MetricsClientsetFactory) *MetricsCollector {
	if factory != nil {
		c.metricsClientFactory = factory
	}

	return c
}

func defaultMetricsClientsetFactory(cfg *rest.Config) (metricsclientset.Interface, error) {
	return metricsclientset.NewForConfig(cfg)
}

// MetricsServerReachable performs a live, side-effect-free check for whether
// metrics-server is installed and responding on the target cluster. Used by
// the provider-status endpoint to distinguish "no data yet" from "the
// provider metrics/CPU/memory is showing is only an estimate."
func MetricsServerReachable(ctx context.Context, client Client) bool {
	if client == nil {
		return false
	}

	cfg, err := client.RESTConfig()
	if err != nil {
		return false
	}

	metricsClient, err := defaultMetricsClientsetFactory(cfg)
	if err != nil {
		return false
	}

	_, err = metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{Limit: 1})
	return err == nil
}

func (c *MetricsCollector) Name() string {
	if c == nil || c.name == "" {
		return "kubernetes"
	}

	return c.name
}

func (c *MetricsCollector) Supports(ctx context.Context) bool {
	if c == nil || c.client == nil {
		return false
	}

	set, err := c.client.Clientset()
	if err != nil {
		return false
	}

	if _, err := set.Discovery().RESTClient().Get().AbsPath("/readyz").DoRaw(ctx); err != nil {
		return false
	}

	return true
}

func (c *MetricsCollector) Collect(ctx context.Context) (*metrics.CollectedMetrics, error) {
	set, err := c.client.Clientset()
	if err != nil {
		return nil, wrapConnectionError(err)
	}

	now := time.Now().UTC()

	namespaceList, err := set.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	nodeList, err := set.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	podList, err := set.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}
	deploymentList, err := set.AppsV1().Deployments(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, wrapConnectionError(err)
	}

	podCountsByNamespace := make(map[string]int)
	for _, pod := range podList.Items {
		podCountsByNamespace[pod.Namespace]++
	}

	resourceCountsByNamespace := make(map[string]int)
	for _, pod := range podList.Items {
		resourceCountsByNamespace[pod.Namespace]++
	}
	for _, deployment := range deploymentList.Items {
		resourceCountsByNamespace[deployment.Namespace]++
	}

	usageByNodeCPU := make(map[string]int64)
	usageByNodeMemory := make(map[string]int64)
	usageByNodeDisk := make(map[string]int64)

	var totalCPUUsage int64
	var totalMemoryUsage int64
	var totalStorageUsage int64

	liveNodeCPU, liveNodeMemory, livePodCPU, livePodMemory, liveUsageAvailable := c.collectLiveUsage(ctx)
	usageSource := usageSourceRequestedEstimate
	if liveUsageAvailable {
		usageSource = usageSourceMetricsServer
	}

	containerMetrics := make([]metrics.ContainerMetrics, 0)
	podMetrics := make([]metrics.PodMetrics, 0, len(podList.Items))

	for _, pod := range podList.Items {
		requestedCPU, requestedMemory, podStorage := podRequestedResources(pod)
		podCPU, podMemory := requestedCPU, requestedMemory
		if liveUsageAvailable {
			key := pod.Namespace + "/" + pod.Name
			if usage, ok := livePodCPU[key]; ok {
				podCPU = usage
			}
			if usage, ok := livePodMemory[key]; ok {
				podMemory = usage
			}
		}

		totalCPUUsage += podCPU
		totalMemoryUsage += podMemory
		totalStorageUsage += podStorage

		if pod.Spec.NodeName != "" {
			usageByNodeCPU[pod.Spec.NodeName] += podCPU
			usageByNodeMemory[pod.Spec.NodeName] += podMemory
			usageByNodeDisk[pod.Spec.NodeName] += podStorage
		}

		podMetrics = append(podMetrics, metrics.PodMetrics{
			Name:          pod.Name,
			Namespace:     pod.Namespace,
			Phase:         string(pod.Status.Phase),
			Ready:         isPodReady(pod),
			Restarts:      sumPodRestarts(pod),
			CPUMilliCores: podCPU,
			MemoryBytes:   podMemory,
			Age:           now.Sub(pod.CreationTimestamp.Time),
			Node:          pod.Spec.NodeName,
		})

		for index := range pod.Spec.Containers {
			container := pod.Spec.Containers[index]
			containerReady, containerRestarts := containerStatus(pod, container.Name)
			containerMetrics = append(containerMetrics, metrics.ContainerMetrics{
				Name:          container.Name,
				Pod:           pod.Name,
				Namespace:     pod.Namespace,
				CPUMilliCores: quantityMilli(container.Resources.Requests[corev1.ResourceCPU]),
				MemoryBytes:   quantityValue(container.Resources.Requests[corev1.ResourceMemory]),
				Restarts:      containerRestarts,
				Ready:         containerReady,
				Age:           now.Sub(pod.CreationTimestamp.Time),
				Node:          pod.Spec.NodeName,
			})
		}
	}

	nodeMetrics := make([]metrics.NodeMetrics, 0, len(nodeList.Items))
	var totalCPUCapacity int64
	var totalMemoryCapacity int64
	var totalStorageCapacity int64

	for _, node := range nodeList.Items {
		cpuCapacity := quantityMilli(node.Status.Capacity[corev1.ResourceCPU])
		memoryCapacity := quantityValue(node.Status.Capacity[corev1.ResourceMemory])
		storageCapacity := quantityValue(node.Status.Capacity[corev1.ResourceEphemeralStorage])

		totalCPUCapacity += cpuCapacity
		totalMemoryCapacity += memoryCapacity
		totalStorageCapacity += storageCapacity

		nodeCPUUsage := usageByNodeCPU[node.Name]
		nodeMemoryUsage := usageByNodeMemory[node.Name]
		nodeDiskUsage := usageByNodeDisk[node.Name]
		if liveUsageAvailable {
			// Real per-node usage from metrics-server is more accurate than
			// summing per-pod usage (which misses system/daemon overhead).
			if usage, ok := liveNodeCPU[node.Name]; ok {
				nodeCPUUsage = usage
			}
			if usage, ok := liveNodeMemory[node.Name]; ok {
				nodeMemoryUsage = usage
			}
		}

		nodeMetrics = append(nodeMetrics, metrics.NodeMetrics{
			Name:          node.Name,
			CPUPercent:    usagePercent(nodeCPUUsage, cpuCapacity),
			MemoryPercent: usagePercent(nodeMemoryUsage, memoryCapacity),
			DiskPercent:   usagePercent(nodeDiskUsage, storageCapacity),
			Ready:         isNodeReady(node),
			Pressure:      nodePressure(node),
			Version:       node.Status.NodeInfo.KubeletVersion,
			InternalIP:    nodeInternalIP(node),
		})
	}

	deploymentMetrics := make([]metrics.DeploymentMetrics, 0, len(deploymentList.Items))
	for _, deployment := range deploymentList.Items {
		desired := int32(0)
		if deployment.Spec.Replicas != nil {
			desired = *deployment.Spec.Replicas
		}

		deploymentMetrics = append(deploymentMetrics, metrics.DeploymentMetrics{
			Name:                deployment.Name,
			Namespace:           deployment.Namespace,
			DesiredReplicas:     desired,
			AvailableReplicas:   deployment.Status.AvailableReplicas,
			UnavailableReplicas: deployment.Status.UnavailableReplicas,
			UpdatedReplicas:     deployment.Status.UpdatedReplicas,
		})
	}

	namespaceMetrics := make([]metrics.NamespaceMetrics, 0, len(namespaceList.Items))
	for _, namespace := range namespaceList.Items {
		namespaceMetrics = append(namespaceMetrics, metrics.NamespaceMetrics{
			Name:      namespace.Name,
			Phase:     string(namespace.Status.Phase),
			PodCount:  podCountsByNamespace[namespace.Name],
			Age:       now.Sub(namespace.CreationTimestamp.Time),
			Resources: resourceCountsByNamespace[namespace.Name],
		})
	}

	clusterMetrics := metrics.ClusterMetrics{
		NodeCount:             len(nodeList.Items),
		PodCount:              len(podList.Items),
		DeploymentCount:       len(deploymentList.Items),
		NamespaceCount:        len(namespaceList.Items),
		CPUCapacityMilliCores: totalCPUCapacity,
		CPUUsageMilliCores:    totalCPUUsage,
		MemoryCapacityBytes:   totalMemoryCapacity,
		MemoryUsageBytes:      totalMemoryUsage,
		StorageCapacityBytes:  totalStorageCapacity,
		StorageUsageBytes:     totalStorageUsage,
		UsageSource:           usageSource,
	}

	return &metrics.CollectedMetrics{
		CollectorName: c.Name(),
		CollectedAt:   now,
		Cluster:       clusterMetrics,
		Nodes:         nodeMetrics,
		Pods:          podMetrics,
		Deployments:   deploymentMetrics,
		Namespaces:    namespaceMetrics,
		Containers:    containerMetrics,
	}, nil
}

// collectLiveUsage attempts to fetch real CPU/memory usage from the
// metrics.k8s.io API (metrics-server). Any failure - metrics-server not
// installed, unreachable, or not yet reporting for a given node/pod - is
// treated as "not available" (ok=false, or missing map entries), never as an
// error: callers must fall back to the requested-capacity estimate rather
// than fabricate usage numbers.
func (c *MetricsCollector) collectLiveUsage(ctx context.Context) (nodeCPU, nodeMemory, podCPU, podMemory map[string]int64, ok bool) {
	cfg, err := c.client.RESTConfig()
	if err != nil {
		return nil, nil, nil, nil, false
	}

	factory := c.metricsClientFactory
	if factory == nil {
		factory = defaultMetricsClientsetFactory
	}

	metricsClient, err := factory(cfg)
	if err != nil {
		return nil, nil, nil, nil, false
	}

	nodeMetricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, nil, nil, false
	}
	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, nil, nil, false
	}

	nodeCPU = make(map[string]int64, len(nodeMetricsList.Items))
	nodeMemory = make(map[string]int64, len(nodeMetricsList.Items))
	for _, item := range nodeMetricsList.Items {
		nodeCPU[item.Name] = quantityMilli(item.Usage[corev1.ResourceCPU])
		nodeMemory[item.Name] = quantityValue(item.Usage[corev1.ResourceMemory])
	}

	podCPU = make(map[string]int64, len(podMetricsList.Items))
	podMemory = make(map[string]int64, len(podMetricsList.Items))
	for _, item := range podMetricsList.Items {
		key := item.Namespace + "/" + item.Name
		var cpu, mem int64
		for _, container := range item.Containers {
			cpu += quantityMilli(container.Usage[corev1.ResourceCPU])
			mem += quantityValue(container.Usage[corev1.ResourceMemory])
		}
		podCPU[key] = cpu
		podMemory[key] = mem
	}

	return nodeCPU, nodeMemory, podCPU, podMemory, true
}

func usagePercent(usage, capacity int64) float64 {
	if capacity <= 0 {
		return 0
	}

	return (float64(usage) / float64(capacity)) * 100
}

func podRequestedResources(pod corev1.Pod) (cpuMilliCores, memoryBytes, storageBytes int64) {
	for index := range pod.Spec.Containers {
		container := pod.Spec.Containers[index]
		cpuMilliCores += quantityMilli(container.Resources.Requests[corev1.ResourceCPU])
		memoryBytes += quantityValue(container.Resources.Requests[corev1.ResourceMemory])
		storageBytes += quantityValue(container.Resources.Requests[corev1.ResourceEphemeralStorage])
	}

	return cpuMilliCores, memoryBytes, storageBytes
}

func quantityMilli(quantity resource.Quantity) int64 {
	return quantity.MilliValue()
}

func quantityValue(quantity resource.Quantity) int64 {
	return quantity.Value()
}

func isPodReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

func sumPodRestarts(pod corev1.Pod) int32 {
	var total int32
	for _, status := range pod.Status.ContainerStatuses {
		total += status.RestartCount
	}

	return total
}

func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

func nodePressure(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case corev1.NodeMemoryPressure, corev1.NodeDiskPressure, corev1.NodePIDPressure:
			if condition.Status == corev1.ConditionTrue {
				return true
			}
		}
	}

	return false
}

func nodeInternalIP(node corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}

	return ""
}

func containerStatus(pod corev1.Pod, containerName string) (ready bool, restarts int32) {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
			return status.Ready, status.RestartCount
		}
	}

	return false, 0
}
