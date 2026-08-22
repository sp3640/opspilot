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

	// ContainerState is the pod's aggregate container state (e.g. "Running",
	// "CrashLoopBackOff", "ImagePullBackOff", "OOMKilled") - the first
	// non-Running container state found across the pod's containers, or
	// "Running" if all are running. Used to detect crash-loop conditions
	// that a bare restart count/ready flag can't distinguish from a healthy
	// pod that simply restarted once and recovered.
	ContainerState string

	// ApplicationID is set only when the pod carries the
	// opspilot/application-id label (i.e. it was deployed through
	// OpsPilot) - used to correlate alerts back to the owning application.
	// Empty for unmanaged/discovered pods.
	ApplicationID string

	CPUMilliCores int64
	MemoryBytes   int64

	Age  time.Duration
	Node string
}

type DeploymentMetrics struct {
	Name      string
	Namespace string

	// ApplicationID is set only when the deployment carries the
	// opspilot/application-id label - see PodMetrics.ApplicationID.
	ApplicationID string

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
