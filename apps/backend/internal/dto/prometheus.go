package dto

import (
	"time"

	"github.com/google/uuid"
)

// PrometheusSample is a single value observed at a specific timestamp.
type PrometheusSample struct {
	Timestamp time.Time         `json:"timestamp"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// PrometheusInstantQueryResult is the normalized response for a Prometheus instant query.
type PrometheusInstantQueryResult struct {
	Query   string             `json:"query"`
	Samples []PrometheusSample `json:"samples"`
}

// PrometheusResourceRef is the minimal resource identity surfaced by a
// Prometheus correlation match - just enough for a caller to display or link
// to the resource, never the full resource record.
type PrometheusResourceRef struct {
	ID        uuid.UUID `json:"id"`
	Kind      string    `json:"kind"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace,omitempty"`
	Cluster   string    `json:"cluster,omitempty"`
}

// PrometheusApplicationRef is the minimal application identity surfaced by a
// Prometheus correlation match.
type PrometheusApplicationRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// PrometheusCorrelationResult is the API-facing shape of a Prometheus
// sample-set correlation: whether it deterministically resolved to a
// resource/application, and if not, why. Mirrors
// services.PrometheusCorrelationResult but never carries internal-only
// wording verbatim - callers get the same conservative "unavailable rather
// than guessing" guarantee the underlying correlation functions provide.
type PrometheusCorrelationResult struct {
	Available   bool                      `json:"available"`
	Reason      string                    `json:"reason,omitempty"`
	Resource    *PrometheusResourceRef    `json:"resource,omitempty"`
	Application *PrometheusApplicationRef `json:"application,omitempty"`
}

// PrometheusQueryResponse is the API response for an ad hoc integration
// Prometheus instant query: the normalized result plus a best-effort
// resource/application correlation computed across the whole organization
// (an ad hoc query has no incident/project context to narrow the search the
// way RCA evidence gathering does).
type PrometheusQueryResponse struct {
	Query       string                      `json:"query"`
	Samples     []PrometheusSample          `json:"samples"`
	Correlation PrometheusCorrelationResult `json:"correlation"`
}

// PrometheusSeries is a normalized time series returned by a Prometheus range query.
type PrometheusSeries struct {
	Metric map[string]string  `json:"metric,omitempty"`
	Values []PrometheusSample `json:"values"`
}

// PrometheusRangeQueryResult is the normalized response for a Prometheus range query.
type PrometheusRangeQueryResult struct {
	Query  string             `json:"query"`
	Series []PrometheusSeries `json:"series"`
}
