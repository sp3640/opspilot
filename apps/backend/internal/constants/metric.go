package constants

// Metric Type values
const (
	MetricTypeCPU          = "CPU"
	MetricTypeMemory       = "MEMORY"
	MetricTypeStorage      = "STORAGE"
	MetricTypeNetwork      = "NETWORK"
	MetricTypeRestartCount = "RESTARTCOUNT"
	MetricTypeReplicaCount = "REPLICACOUNT"
	MetricTypeAvailability = "AVAILABILITY"
	MetricTypeCustom       = "CUSTOM"
)

// ValidMetricTypes is the slice of all valid metric type values.
var ValidMetricTypes = []string{
	MetricTypeCPU,
	MetricTypeMemory,
	MetricTypeStorage,
	MetricTypeNetwork,
	MetricTypeRestartCount,
	MetricTypeReplicaCount,
	MetricTypeAvailability,
	MetricTypeCustom,
}

func IsValidMetricType(metricType string) bool {
	switch metricType {
	case MetricTypeCPU,
		MetricTypeMemory,
		MetricTypeStorage,
		MetricTypeNetwork,
		MetricTypeRestartCount,
		MetricTypeReplicaCount,
		MetricTypeAvailability,
		MetricTypeCustom:
		return true
	default:
		return false
	}
}
