package alerting

// EvaluatedCondition is one currently-firing condition detected from a
// Kubernetes metrics snapshot. Title must be stable across evaluations of
// the same underlying condition (it feeds Alert deduplication fingerprints
// upstream) - it must never embed a value that changes every pass, such as
// the current CPU percentage or restart count. Those go in Description and
// CurrentValue/Threshold instead.
type EvaluatedCondition struct {
	ResourceType  string // one of constants.AlertResourceType*
	ResourceID    string // stable identifier: cluster ID, node name, or "namespace/name"
	ResourceName  string // human-readable label for the resource
	ConditionKey  string // one of the Condition* keys below
	Title         string // stable, no embedded current value
	Description   string // human text, may embed the current value
	Severity      string // constants.AlertSeverity*
	CurrentValue  float64
	Threshold     float64
	Unit          string
	ClusterID     string
	ClusterName   string
	ApplicationID string // set only when the resource carries an opspilot/application-id label
}

// ConditionMetadata is the JSON shape stored in Alert.Metadata for
// engine-generated alerts, letting the reconciler recognize and correlate an
// existing alert with a currently-evaluated condition without needing a
// dedicated database column.
type ConditionMetadata struct {
	Condition     string  `json:"condition"`
	CurrentValue  float64 `json:"currentValue"`
	Threshold     float64 `json:"threshold"`
	Unit          string  `json:"unit,omitempty"`
	ClusterID     string  `json:"clusterId,omitempty"`
	ClusterName   string  `json:"clusterName,omitempty"`
	ApplicationID string  `json:"applicationId,omitempty"`
}

// Condition keys - stable identifiers for the conditions this engine can
// detect with the metrics architecture built in prior phases. Error rate and
// latency are deliberately absent: no application-level metrics provider
// exists anywhere in the system (see Phase 9/12/13), so evaluating those
// conditions would mean fabricating alerts from data that doesn't exist.
const (
	ConditionCPUThreshold      = "cpu_threshold"
	ConditionMemoryThreshold   = "memory_threshold"
	ConditionPodRestartCount   = "pod_restart_count"
	ConditionCrashLoopBackOff  = "crashloopbackoff"
	ConditionNodeNotReady      = "node_not_ready"
	ConditionDeploymentFailure = "deployment_failure"
)
