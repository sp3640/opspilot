package constants

// Resource Kind values
const (
	ResourceKindCluster               = "Cluster"
	ResourceKindNamespace             = "Namespace"
	ResourceKindNode                  = "Node"
	ResourceKindDeployment            = "Deployment"
	ResourceKindReplicaSet            = "ReplicaSet"
	ResourceKindPod                   = "Pod"
	ResourceKindContainer             = "Container"
	ResourceKindService               = "Service"
	ResourceKindIngress               = "Ingress"
	ResourceKindConfigMap             = "ConfigMap"
	ResourceKindSecret                = "Secret"
	ResourceKindPersistentVolume      = "PersistentVolume"
	ResourceKindPersistentVolumeClaim = "PersistentVolumeClaim"
	ResourceKindDatabase              = "Database"
	ResourceKindVM                    = "VM"
	ResourceKindApplication           = "Application"
	ResourceKindQueue                 = "Queue"
	ResourceKindStorage               = "Storage"
	ResourceKindCache                 = "Cache"
	ResourceKindLoadBalancer          = "LoadBalancer"
)

// Resource Status values
const (
	ResourceStatusActive      = "ACTIVE"
	ResourceStatusPending     = "PENDING"
	ResourceStatusUpdating    = "UPDATING"
	ResourceStatusTerminating = "TERMINATING"
	ResourceStatusDeleted     = "DELETED"
	ResourceStatusUnknown     = "UNKNOWN"
)

// Resource Health values
const (
	ResourceHealthHealthy   = "HEALTHY"
	ResourceHealthDegraded  = "DEGRADED"
	ResourceHealthUnhealthy = "UNHEALTHY"
	ResourceHealthUnknown   = "UNKNOWN"
)

// ValidResourceKinds is the slice of all valid resource kind values.
var ValidResourceKinds = []string{
	ResourceKindCluster,
	ResourceKindNamespace,
	ResourceKindNode,
	ResourceKindDeployment,
	ResourceKindReplicaSet,
	ResourceKindPod,
	ResourceKindContainer,
	ResourceKindService,
	ResourceKindIngress,
	ResourceKindConfigMap,
	ResourceKindSecret,
	ResourceKindPersistentVolume,
	ResourceKindPersistentVolumeClaim,
	ResourceKindDatabase,
	ResourceKindVM,
	ResourceKindApplication,
	ResourceKindQueue,
	ResourceKindStorage,
	ResourceKindCache,
	ResourceKindLoadBalancer,
}

// ValidResourceStatuses is the slice of all valid resource status values.
var ValidResourceStatuses = []string{
	ResourceStatusActive,
	ResourceStatusPending,
	ResourceStatusUpdating,
	ResourceStatusTerminating,
	ResourceStatusDeleted,
	ResourceStatusUnknown,
}

// ValidResourceHealthValues is the slice of all valid resource health values.
var ValidResourceHealthValues = []string{
	ResourceHealthHealthy,
	ResourceHealthDegraded,
	ResourceHealthUnhealthy,
	ResourceHealthUnknown,
}

func IsValidResourceKind(kind string) bool {
	switch kind {
	case ResourceKindCluster,
		ResourceKindNamespace,
		ResourceKindNode,
		ResourceKindDeployment,
		ResourceKindReplicaSet,
		ResourceKindPod,
		ResourceKindContainer,
		ResourceKindService,
		ResourceKindIngress,
		ResourceKindConfigMap,
		ResourceKindSecret,
		ResourceKindPersistentVolume,
		ResourceKindPersistentVolumeClaim,
		ResourceKindDatabase,
		ResourceKindVM,
		ResourceKindApplication,
		ResourceKindQueue,
		ResourceKindStorage,
		ResourceKindCache,
		ResourceKindLoadBalancer:
		return true
	default:
		return false
	}
}

func IsValidResourceStatus(status string) bool {
	switch status {
	case ResourceStatusActive,
		ResourceStatusPending,
		ResourceStatusUpdating,
		ResourceStatusTerminating,
		ResourceStatusDeleted,
		ResourceStatusUnknown:
		return true
	default:
		return false
	}
}

func IsValidResourceHealth(health string) bool {
	switch health {
	case ResourceHealthHealthy,
		ResourceHealthDegraded,
		ResourceHealthUnhealthy,
		ResourceHealthUnknown:
		return true
	default:
		return false
	}
}
