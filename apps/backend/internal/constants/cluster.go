package constants

// Cluster status values.
const (
	ClusterStatusConnected    = "CONNECTED"
	ClusterStatusDisconnected = "DISCONNECTED"
	ClusterStatusFailed       = "FAILED"
	ClusterStatusPending      = "PENDING"
)

// Cluster provider values.
const (
	ClusterProviderKubernetes = "KUBERNETES"
	ClusterProviderDocker     = "DOCKER"
	ClusterProviderVM         = "VM"
	ClusterProviderAzure      = "AZURE"
	ClusterProviderAWS        = "AWS"
	ClusterProviderCustom     = "CUSTOM"
)

// Cluster connection type values.
const (
	ClusterConnectionTypeKubeconfig = "KUBECONFIG"
	ClusterConnectionTypeAPI        = "API"
	ClusterConnectionTypeSocket     = "SOCKET"
	ClusterConnectionTypeCustom     = "CUSTOM"
)

// ValidClusterStatuses is the slice of all valid cluster status values.
var ValidClusterStatuses = []string{
	ClusterStatusConnected,
	ClusterStatusDisconnected,
	ClusterStatusFailed,
	ClusterStatusPending,
}

// ValidClusterProviders is the slice of all valid cluster provider values.
var ValidClusterProviders = []string{
	ClusterProviderKubernetes,
	ClusterProviderDocker,
	ClusterProviderVM,
	ClusterProviderAzure,
	ClusterProviderAWS,
	ClusterProviderCustom,
}

// ValidClusterConnectionTypes is the slice of all valid cluster connection type values.
var ValidClusterConnectionTypes = []string{
	ClusterConnectionTypeKubeconfig,
	ClusterConnectionTypeAPI,
	ClusterConnectionTypeSocket,
	ClusterConnectionTypeCustom,
}

func IsValidClusterStatus(status string) bool {
	switch status {
	case ClusterStatusConnected,
		ClusterStatusDisconnected,
		ClusterStatusFailed,
		ClusterStatusPending:
		return true
	default:
		return false
	}
}

func IsValidClusterProvider(provider string) bool {
	switch provider {
	case ClusterProviderKubernetes,
		ClusterProviderDocker,
		ClusterProviderVM,
		ClusterProviderAzure,
		ClusterProviderAWS,
		ClusterProviderCustom:
		return true
	default:
		return false
	}
}

func IsValidClusterConnectionType(connectionType string) bool {
	switch connectionType {
	case ClusterConnectionTypeKubeconfig,
		ClusterConnectionTypeAPI,
		ClusterConnectionTypeSocket,
		ClusterConnectionTypeCustom:
		return true
	default:
		return false
	}
}
