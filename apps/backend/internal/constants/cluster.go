package constants

// Cluster status values.
const (
	ClusterStatusPendingValidation = "PENDING_VALIDATION"
	ClusterStatusHealthy           = "HEALTHY"
	ClusterStatusInvalid           = "INVALID"

	ClusterStatusPending      = ClusterStatusPendingValidation
	ClusterStatusConnected    = ClusterStatusHealthy
	ClusterStatusDisconnected = ClusterStatusInvalid
	ClusterStatusFailed       = ClusterStatusInvalid
)

// Cluster provider values.
//
// ClusterProviderKubernetes is its own distinct value (a generic/vanilla
// Kubernetes cluster), separate from ClusterProviderKind (the "KIND"
// local-dev tool) — they used to be aliased to the same string, which meant
// the frontend's generic "Kubernetes" option was silently rejected as an
// invalid provider.
const (
	ClusterProviderAKS      = "AKS"
	ClusterProviderEKS      = "EKS"
	ClusterProviderGKE      = "GKE"
	ClusterProviderOnPrem   = "ON-PREM"
	ClusterProviderKind     = "KIND"
	ClusterProviderMinikube = "MINIKUBE"

	ClusterProviderKubernetes = "KUBERNETES"
	ClusterProviderDocker     = "DOCKER"
	ClusterProviderVM         = "VM"
	ClusterProviderAzure      = "AZURE"
	ClusterProviderAWS        = "AWS"
	ClusterProviderCustom     = "CUSTOM"
)

// Cluster credential type values.
const (
	ClusterCredentialTypeKubeConfig     = "KUBECONFIG"
	ClusterCredentialTypeBearerToken    = "BEARER_TOKEN"
	ClusterCredentialTypeServiceAccount = "SERVICE_ACCOUNT"

	ClusterConnectionTypeKubeconfig = ClusterCredentialTypeKubeConfig
	ClusterConnectionTypeAPI        = "API"
	ClusterConnectionTypeSocket     = "SOCKET"
	ClusterConnectionTypeCustom     = "CUSTOM"
)

// ValidClusterStatuses is the slice of all valid cluster status values.
var ValidClusterStatuses = []string{
	ClusterStatusPendingValidation,
	ClusterStatusHealthy,
	ClusterStatusInvalid,
	ClusterStatusConnected,
	ClusterStatusDisconnected,
	ClusterStatusFailed,
	ClusterStatusPending,
}

// ValidClusterProviders is the slice of all valid cluster provider values.
var ValidClusterProviders = []string{
	ClusterProviderAKS,
	ClusterProviderEKS,
	ClusterProviderGKE,
	ClusterProviderOnPrem,
	ClusterProviderKind,
	ClusterProviderMinikube,
	ClusterProviderKubernetes,
	ClusterProviderDocker,
	ClusterProviderVM,
	ClusterProviderAzure,
	ClusterProviderAWS,
	ClusterProviderCustom,
}

// ValidClusterConnectionTypes is the slice of all valid cluster credential type values.
var ValidClusterConnectionTypes = []string{
	ClusterCredentialTypeKubeConfig,
	ClusterCredentialTypeBearerToken,
	ClusterCredentialTypeServiceAccount,
	ClusterConnectionTypeKubeconfig,
	ClusterConnectionTypeAPI,
	ClusterConnectionTypeSocket,
	ClusterConnectionTypeCustom,
}

func IsValidClusterStatus(status string) bool {
	switch status {
	case ClusterStatusPendingValidation,
		ClusterStatusHealthy,
		ClusterStatusInvalid:
		return true
	default:
		return false
	}
}

func IsValidClusterProvider(provider string) bool {
	switch provider {
	case ClusterProviderAKS,
		ClusterProviderEKS,
		ClusterProviderGKE,
		ClusterProviderOnPrem,
		ClusterProviderKind,
		ClusterProviderMinikube,
		ClusterProviderKubernetes:
		return true
	case ClusterProviderDocker,
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
	return IsValidClusterCredentialType(connectionType)
}

func IsValidClusterCredentialType(credentialType string) bool {
	switch credentialType {
	case ClusterCredentialTypeKubeConfig,
		ClusterCredentialTypeBearerToken,
		ClusterCredentialTypeServiceAccount,
		ClusterConnectionTypeAPI,
		ClusterConnectionTypeSocket,
		ClusterConnectionTypeCustom:
		return true
	default:
		return false
	}
}
