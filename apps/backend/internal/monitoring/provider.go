package monitoring

type ProviderType string

const (
	ProviderKubernetes ProviderType = "KUBERNETES"
	ProviderDocker     ProviderType = "DOCKER"
	ProviderAzure      ProviderType = "AZURE"
	ProviderAWS        ProviderType = "AWS"
	ProviderPrometheus ProviderType = "PROMETHEUS"
	ProviderGrafana    ProviderType = "GRAFANA"
	ProviderVM         ProviderType = "VM"
	ProviderCustom     ProviderType = "CUSTOM"
)
