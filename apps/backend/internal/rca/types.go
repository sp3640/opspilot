// Package rca computes a deterministic Root Cause Analysis for one incident
// from already-fetched, real evidence only. Evaluate is a pure function: it
// never queries anything itself, so every claim it makes is traceable to a
// value the caller actually observed (mirroring internal/health's design).
//
// This package never asserts certainty. A deployment, config change, pod
// crash, or metric shift that lines up with the incident window is reported
// as a "possible cause" with the corroborating evidence listed alongside it
// - never as a confirmed root cause. Wording throughout uses "possible
// cause", "potentially related", and "evidence suggests" deliberately, and
// must keep doing so as this package grows.
package rca

import "time"

// EvidenceLevel is the deterministic confidence this analysis has in its
// own findings - not confidence that a cause is correct, but confidence
// that the available evidence meaningfully supports (or fails to support)
// any conclusion at all.
type EvidenceLevel string

const (
	// EvidenceLevelInsufficient means too little real evidence was
	// available (few/no evidence categories had data) to suggest anything
	// beyond the raw facts already on the incident.
	EvidenceLevelInsufficient EvidenceLevel = "INSUFFICIENT"
	// EvidenceLevelWeak means at least one signal was found, but it is
	// isolated - nothing else corroborates it.
	EvidenceLevelWeak EvidenceLevel = "WEAK"
	// EvidenceLevelModerate means one or more signals corroborate each
	// other (e.g. a deployment followed by a metric shift).
	EvidenceLevelModerate EvidenceLevel = "MODERATE"
	// EvidenceLevelStrong means multiple independent evidence categories
	// (deployment timing, metrics, pod health, logs, k8s events) all point
	// the same direction. Still never "confirmed" - only the strongest
	// correlation this package ever reports.
	EvidenceLevelStrong EvidenceLevel = "STRONG"
)

// TimelineEventType identifies the evidence category a TimelineEvent came
// from, so the UI can render a distinct icon/color per source.
type TimelineEventType string

const (
	TimelineEventIncident     TimelineEventType = "incident"
	TimelineEventAlert        TimelineEventType = "alert"
	TimelineEventDeployment   TimelineEventType = "deployment"
	TimelineEventConfigChange TimelineEventType = "config_change"
	TimelineEventKubernetes   TimelineEventType = "kubernetes_event"
)

// TimelineEvent is one already-real, timestamped occurrence folded into the
// incident's evidence timeline.
type TimelineEvent struct {
	Timestamp   time.Time
	Type        TimelineEventType
	Title       string
	Description string
	Source      string // free-text origin, e.g. "alert #42", "deployment abc123"
}

// AlertRecord is the minimal real alert data needed for correlation.
type AlertRecord struct {
	ID              uint
	Title           string
	Severity        string
	Status          string
	FirstSeenAt     time.Time
	LastSeenAt      time.Time
	OccurrenceCount int
}

// MetricPoint is one already-persisted metric datapoint.
type MetricPoint struct {
	MetricType string
	MetricName string
	Unit       string
	Value      float64
	Timestamp  time.Time
}

// LogLine is one already-fetched line of pod log output.
type LogLine struct {
	PodName   string
	Container string
	Line      string
}

// DeploymentRecord is one already-persisted deployment relevant to the
// incident's application.
type DeploymentRecord struct {
	ID            string
	ImageTag      string
	Status        string
	Environment   string
	CreatedAt     time.Time
	Revision      int
	ChangeSummary string
}

// K8sEvent is one already-fetched live Kubernetes event.
type K8sEvent struct {
	Reason         string
	Message        string
	Type           string // "Normal" or "Warning"
	Count          int32
	LastTimestamp  *time.Time
	InvolvedObject string
}

// PodStatus is the already-fetched live status of one pod.
type PodStatus struct {
	Name         string
	Ready        bool
	RestartCount int32
	Phase        string
	Reason       string
}

// ConfigChangeRecord is one already-persisted audit log entry describing a
// change to some entity (deployment, cluster, application, SLO, ...).
type ConfigChangeRecord struct {
	EntityType string
	EntityID   string
	Action     string
	FieldName  string
	OldValue   string
	NewValue   string
	ChangedAt  time.Time
}

// Input carries every already-fetched real signal for one incident at
// analysis time. Each evidence slice has a matching *Available flag: a
// caller sets it false (leaving the slice empty) to honestly report "this
// category could not be queried" rather than reporting an empty result as
// "queried and found nothing" - the same Unknown-vs-zero discipline
// internal/health.Input uses.
type Input struct {
	IncidentID             uint
	IncidentTitle          string
	IncidentDescription    string
	IncidentSeverity       string
	IncidentStatus         string
	IncidentCreatedAt      time.Time
	IncidentAcknowledgedAt *time.Time
	IncidentResolvedAt     *time.Time

	WindowStart time.Time
	WindowEnd   time.Time

	ApplicationAvailable bool
	ApplicationID        string
	ApplicationName      string

	Alerts          []AlertRecord
	AlertsAvailable bool

	Metrics          []MetricPoint
	MetricsAvailable bool

	Logs          []LogLine
	LogsAvailable bool

	Deployments          []DeploymentRecord
	DeploymentsAvailable bool

	K8sEvents          []K8sEvent
	K8sEventsAvailable bool

	Pods          []PodStatus
	PodsAvailable bool

	ConfigChanges          []ConfigChangeRecord
	ConfigChangesAvailable bool
}

// PossibleCause is one hedged, evidence-backed hypothesis - never a
// confirmed root cause.
type PossibleCause struct {
	Title      string
	Category   string
	Confidence EvidenceLevel
	Evidence   []string
}

// Result is the full deterministic RCA output.
type Result struct {
	Summary                       string
	AffectedApplication           string
	ApplicationKnown              bool
	Timeline                      []TimelineEvent
	RecentChanges                 []ConfigChangeRecord
	CorrelatedAlerts              []AlertRecord
	RelevantMetrics               []MetricSignal
	RelevantLogs                  []LogSignal
	KubernetesEvidence            []K8sEvent
	PodEvidence                   []PodStatus
	PossibleCauses                []PossibleCause
	RecommendedInvestigationSteps []string
	RecommendedRemediation        []string
	EvidenceLevel                 EvidenceLevel
	EvidenceLevelReason           string
}

// MetricSignal summarizes one metric's behavior across the incident window:
// the average value before the window midpoint vs. after it, so a caller
// can see at a glance whether it rose, fell, or stayed flat.
type MetricSignal struct {
	MetricType    string
	MetricName    string
	Unit          string
	Before        float64
	After         float64
	ChangePercent float64
	Direction     string // "increased", "decreased", "stable"
	SampleCount   int
}

// LogSignal is one log line flagged by deterministic keyword matching
// (never by an LLM) as potentially relevant to the incident.
type LogSignal struct {
	PodName   string
	Container string
	Line      string
	Keyword   string
}
