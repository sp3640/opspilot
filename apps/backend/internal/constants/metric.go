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

	// Application-category metric types. No provider (e.g. Prometheus/APM)
	// is connected yet, so nothing populates these today - they exist so the
	// type system has a defined, honest place to report "unavailable"
	// rather than a caller inventing an ad hoc CUSTOM metric name for them.
	MetricTypeLatency     = "LATENCY"
	MetricTypeErrorRate   = "ERRORRATE"
	MetricTypeRequestRate = "REQUESTRATE"
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
	MetricTypeLatency,
	MetricTypeErrorRate,
	MetricTypeRequestRate,
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
		MetricTypeCustom,
		MetricTypeLatency,
		MetricTypeErrorRate,
		MetricTypeRequestRate:
		return true
	default:
		return false
	}
}
