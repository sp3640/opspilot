package metrics

import "time"

type CollectedMetrics struct {
	CollectorName string
	CollectedAt   time.Time
	Cluster       ClusterMetrics
	Nodes         []NodeMetrics
	Pods          []PodMetrics
	Deployments   []DeploymentMetrics
	Namespaces    []NamespaceMetrics
	Containers    []ContainerMetrics
}

type ClusterMetrics struct {
	NodeCount       int
	PodCount        int
	DeploymentCount int
	NamespaceCount  int

	CPUCapacityMilliCores int64
	CPUUsageMilliCores    int64

	MemoryCapacityBytes int64
	MemoryUsageBytes    int64

	StorageCapacityBytes int64
	StorageUsageBytes    int64

	// UsageSource records whether CPUUsageMilliCores/MemoryUsageBytes are
	// real live usage ("metrics-server") or an estimate derived from
	// declared container resource requests ("requested-capacity") when
	// metrics-server is not installed on the target cluster.
	UsageSource string
}

type NodeMetrics struct {
	Name string

	CPUPercent    float64
	MemoryPercent float64
	DiskPercent   float64

	Ready    bool
	Pressure bool

	Version    string
	InternalIP string
}

type PodMetrics struct {
	Name      string
	Namespace string

	Phase string
	Ready bool

	Restarts int32

	CPUMilliCores int64
	MemoryBytes   int64

	Age  time.Duration
	Node string
}

type DeploymentMetrics struct {
	Name      string
	Namespace string

	DesiredReplicas     int32
	AvailableReplicas   int32
	UnavailableReplicas int32
	UpdatedReplicas     int32
}

type NamespaceMetrics struct {
	Name      string
	Phase     string
	PodCount  int
	Age       time.Duration
	Resources int
}

type ContainerMetrics struct {
	Name      string
	Pod       string
	Namespace string

	CPUMilliCores int64
	MemoryBytes   int64

	Restarts int32
	Ready    bool

	Age  time.Duration
	Node string
}
