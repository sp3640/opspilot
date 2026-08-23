package dto

import "time"

// ApplicationSLOConfigRequest is the validated input for configuring an
// application's SLO target and rolling measurement window.
type ApplicationSLOConfigRequest struct {
	TargetPercentage float64 `json:"target_percentage" binding:"required"`
	WindowDays       int     `json:"window_days" binding:"required"`
}

type ApplicationSLOConfigResponse struct {
	ApplicationID    string    `json:"applicationId"`
	TargetPercentage float64   `json:"targetPercentage"`
	WindowDays       int       `json:"windowDays"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// MetricValue represents one computed SRE metric. Available is false when
// there is not enough real historical data to compute it honestly -
// Formatted then reads "Insufficient data" and Reason explains why; Value
// is always zero in that case, never a fabricated number. Formula is a
// short, human-readable description of exactly how Value was derived (see
// docs/sre-metrics.md for the full methodology), returned so the frontend
// never has to hardcode an explanation that could drift from the backend.
type MetricValue struct {
	Available bool    `json:"available"`
	Value     float64 `json:"value,omitempty"`
	Unit      string  `json:"unit,omitempty"`
	Formatted string  `json:"formatted"`
	Reason    string  `json:"reason,omitempty"`
	Formula   string  `json:"formula"`
}

// ApplicationSREMetricsResponse is the full SRE metrics view for one
// application: availability/uptime, SLO compliance and error budget (both
// only meaningful once an SLO is configured), and the incident-response
// timing metrics MTTR/MTTA - all computed from real Incident history, never
// invented. See SREMetricsService.GetMetrics for the exact formulas.
type ApplicationSREMetricsResponse struct {
	ApplicationID string    `json:"applicationId"`
	WindowDays    int       `json:"windowDays"`
	WindowStart   time.Time `json:"windowStart"`
	WindowEnd     time.Time `json:"windowEnd"`
	// ObservedDays is the actual amount of history available within the
	// window - never larger than WindowDays, and smaller whenever the
	// application is younger than the configured window.
	ObservedDays  float64     `json:"observedDays"`
	SLOConfigured bool        `json:"sloConfigured"`
	CurrentSLO    *float64    `json:"currentSlo,omitempty"`
	Availability  MetricValue `json:"availability"`
	Uptime        MetricValue `json:"uptime"`
	SLOCompliance MetricValue `json:"sloCompliance"`
	ErrorBudget   MetricValue `json:"errorBudget"`
	MTTR          MetricValue `json:"mttr"`
	MTTA          MetricValue `json:"mtta"`
	IncidentCount int         `json:"incidentCount"`
}
