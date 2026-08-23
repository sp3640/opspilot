package dto

import "time"

// ─── Incident Intelligence / Root Cause Analysis DTOs ─────────────────────
//
// Every value below is a deterministic correlation over already-real
// evidence computed by internal/rca - see that package's doc comment. No
// field here is AI-generated. If an LLM-backed narrator is ever added on
// top of this evidence layer, its output must be isolated behind its own
// service interface and clearly labeled as AI-generated, never mixed into
// these fields (see apps/backend/phase.md).

type RCATimelineEventResponse struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Source      string    `json:"source,omitempty"`
}

type RCAAlertResponse struct {
	ID              uint      `json:"id"`
	Title           string    `json:"title"`
	Severity        string    `json:"severity"`
	Status          string    `json:"status"`
	FirstSeenAt     time.Time `json:"firstSeenAt"`
	LastSeenAt      time.Time `json:"lastSeenAt"`
	OccurrenceCount int       `json:"occurrenceCount"`
}

type RCAMetricSignalResponse struct {
	MetricType    string  `json:"metricType"`
	MetricName    string  `json:"metricName"`
	Unit          string  `json:"unit,omitempty"`
	Before        float64 `json:"before"`
	After         float64 `json:"after"`
	ChangePercent float64 `json:"changePercent"`
	Direction     string  `json:"direction"`
	SampleCount   int     `json:"sampleCount"`
}

type RCALogSignalResponse struct {
	PodName   string `json:"podName"`
	Container string `json:"container,omitempty"`
	Line      string `json:"line"`
	Keyword   string `json:"keyword"`
}

type RCAKubernetesEventResponse struct {
	Reason         string     `json:"reason"`
	Message        string     `json:"message"`
	Type           string     `json:"type"`
	Count          int32      `json:"count"`
	LastTimestamp  *time.Time `json:"lastTimestamp,omitempty"`
	InvolvedObject string     `json:"involvedObject"`
}

type RCAPodResponse struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restartCount"`
	Phase        string `json:"phase,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

type RCAConfigChangeResponse struct {
	EntityType string    `json:"entityType"`
	EntityID   string    `json:"entityId"`
	Action     string    `json:"action"`
	FieldName  string    `json:"fieldName,omitempty"`
	OldValue   string    `json:"oldValue,omitempty"`
	NewValue   string    `json:"newValue,omitempty"`
	ChangedAt  time.Time `json:"changedAt"`
}

// RCAPossibleCauseResponse is one hedged hypothesis. Confidence is always
// one of INSUFFICIENT/WEAK/MODERATE/STRONG - never a claim of certainty -
// and Title/Evidence use "possible cause" / "may be related" / "evidence
// suggests" language rather than asserting causation.
type RCAPossibleCauseResponse struct {
	Title      string   `json:"title"`
	Category   string   `json:"category"`
	Confidence string   `json:"confidence"`
	Evidence   []string `json:"evidence"`
}

// IncidentRCAResponse is the full Incident Intelligence / Root Cause
// Analysis payload for one incident.
type IncidentRCAResponse struct {
	IncidentID uint   `json:"incidentId"`
	Summary    string `json:"summary"`

	AffectedApplicationID   *string `json:"affectedApplicationId,omitempty"`
	AffectedApplicationName string  `json:"affectedApplicationName,omitempty"`
	ApplicationKnown        bool    `json:"applicationKnown"`

	WindowStart time.Time `json:"windowStart"`
	WindowEnd   time.Time `json:"windowEnd"`

	Timeline           []RCATimelineEventResponse   `json:"timeline"`
	RecentChanges      []RCAConfigChangeResponse    `json:"recentChanges"`
	CorrelatedAlerts   []RCAAlertResponse           `json:"correlatedAlerts"`
	RelevantMetrics    []RCAMetricSignalResponse    `json:"relevantMetrics"`
	RelevantLogs       []RCALogSignalResponse       `json:"relevantLogs"`
	KubernetesEvidence []RCAKubernetesEventResponse `json:"kubernetesEvidence"`
	PodEvidence        []RCAPodResponse             `json:"podEvidence"`

	PossibleCauses                []RCAPossibleCauseResponse `json:"possibleCauses"`
	RecommendedInvestigationSteps []string                   `json:"recommendedInvestigationSteps"`
	RecommendedRemediation        []string                   `json:"recommendedRemediation"`

	// EvidenceLevel is this analysis's confidence in its own findings
	// (INSUFFICIENT/WEAK/MODERATE/STRONG) - never "confirmed". Remediation
	// is always a recommendation; approving and applying it remains a
	// human engineer's responsibility.
	EvidenceLevel       string `json:"evidenceLevel"`
	EvidenceLevelReason string `json:"evidenceLevelReason"`

	GeneratedAt time.Time `json:"generatedAt"`
}
