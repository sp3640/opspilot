package alerting

// Default thresholds for engine-generated conditions. Chosen as reasonable,
// conservative infrastructure defaults (not tunable per-organization yet -
// that would be a natural follow-up, not part of this phase).
const (
	CPUWarningPercent     = 80.0
	CPUCriticalPercent    = 90.0
	MemoryWarningPercent  = 80.0
	MemoryCriticalPercent = 90.0

	PodRestartWarningCount  = 5.0
	PodRestartCriticalCount = 10.0
)
