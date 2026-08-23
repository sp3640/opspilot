// Package health computes a deterministic, documented application health
// score from already-fetched, real signals only. Evaluate is a pure
// function: it never queries anything itself, so every factor it reports is
// traceable to a value the caller actually observed. A signal the caller
// could not obtain (cluster unreachable, application never deployed, no
// data source exists at all) is represented as a nil/zero field on Input,
// which always produces StateUnknown for that factor - never a fabricated
// StateHealthy.
package health

import (
	"math"
	"strconv"
)

// State is one of exactly four values, deliberately excluding anything
// finer-grained: a health score is only as trustworthy as its simplest
// honest summary.
type State string

const (
	StateHealthy  State = "HEALTHY"
	StateDegraded State = "DEGRADED"
	StateCritical State = "CRITICAL"
	// StateUnknown means insufficient real data existed to evaluate this
	// factor - not "assumed fine" and not counted toward the overall score.
	StateUnknown State = "UNKNOWN"
)

// FactorKey identifies one of the fixed set of factors this package always
// evaluates and reports, regardless of whether real data was available for
// it (an unavailable factor is still returned, with StateUnknown).
type FactorKey string

const (
	FactorAvailability        FactorKey = "availability"
	FactorErrorRate           FactorKey = "error_rate"
	FactorLatency             FactorKey = "latency"
	FactorPodHealth           FactorKey = "pod_health"
	FactorDeployment          FactorKey = "deployment"
	FactorAlerts              FactorKey = "alerts"
	FactorIncidents           FactorKey = "incidents"
	FactorResourceUtilization FactorKey = "resource_utilization"
)

// Factor weights sum to 100 and are fixed/documented here as the single
// source of truth. Error rate, latency, and resource utilization currently
// have no real data source anywhere in OpsPilot (no APM/Prometheus
// integration, no bulk pod-metrics collection) and are therefore always
// StateUnknown - their weight is reserved rather than zeroed, so that if a
// future phase adds a real source for one of them, it starts contributing
// to the score without any weight redesign.
const (
	weightAvailability        = 20
	weightErrorRate           = 5
	weightLatency             = 5
	weightPodHealth           = 20
	weightDeployment          = 15
	weightAlerts              = 15
	weightIncidents           = 15
	weightResourceUtilization = 5
)

// Alert/incident count thresholds. Fixed and documented here, not
// configurable, so the score is reproducible from the same input every time.
const (
	alertsDegradedAt = 1 // 1 active alert -> Degraded
	alertsCriticalAt = 2 // 2+ active alerts -> Critical

	incidentsCriticalAt = 1 // any active incident on this application -> Critical
)

// Factor is one row of the health breakdown: a fixed key/label, the state
// this evaluation reached, a 0-100 score (nil when State is Unknown - there
// is no meaningful score for a factor with no data), the factor's fixed
// weight, and a short human-readable reason a caller can render directly.
type Factor struct {
	Key    FactorKey
	Label  string
	State  State
	Score  *int
	Weight int
	Reason string
}

// PodInput is the minimal, already-fetched real state of one pod needed to
// evaluate the Pod Health factor.
type PodInput struct {
	Ready        bool
	RestartCount int32
}

// Input carries every already-fetched real signal for one application at
// evaluation time. Every field is a pointer/optional or an explicit
// "available" flag; leaving a field unset is how a caller honestly reports
// "I don't have this data" rather than a fabricated zero value.
type Input struct {
	// RuntimeDeploymentStatus is the live Kubernetes Deployment's derived
	// status (Healthy/Progressing/Scaling/Unavailable/Failed/Unknown, from
	// internal/kubernetes/deployments.DetermineDeploymentStatus). Nil means
	// no live deployment could be matched (cluster unreachable, application
	// never deployed, or no runtime Deployment object exists yet).
	RuntimeDeploymentStatus *string

	// DeploymentStatus is the most recent database deployment record's
	// lifecycle status (Pending/Queued/Running/Succeeded/Failed/Cancelled/
	// RolledBack). Nil means the application has never been deployed.
	DeploymentStatus *string

	// PodsAvailable reports whether pod data could be fetched at all
	// (false = cluster unreachable / query failed, distinct from "queried
	// successfully and found zero pods"). Pods is only meaningful when true.
	PodsAvailable bool
	Pods          []PodInput

	// ActiveAlertCount is the number of currently-active alerts
	// attributable to this specific application. Nil means the query
	// itself could not be performed; 0 is a real, positive finding (not
	// "unknown"), since the underlying query always succeeds or fails
	// cleanly.
	ActiveAlertCount *int

	// ActiveIncidentCount is the number of currently-active incidents
	// scoped to this application (Incident.ApplicationID is a real,
	// precise column - not a best-effort match). Same nil/0 semantics as
	// ActiveAlertCount.
	ActiveIncidentCount *int
}

// Result is the full, deterministic evaluation output: every factor (even
// ones that are always Unknown today) plus the overall weighted score.
type Result struct {
	// Score is the weighted average of every non-Unknown factor's score,
	// rounded to the nearest integer, or nil when every factor is Unknown
	// (no real data existed for this application at all).
	Score        *int
	OverallState State
	Factors      []Factor
}

// Evaluate computes the health score for one application from already
// fetched, real data. It is a pure function - deterministic for the same
// Input, and safe to unit test without any I/O.
func Evaluate(input Input) Result {
	factors := []Factor{
		evaluateAvailability(input),
		evaluateErrorRate(),
		evaluateLatency(),
		evaluatePodHealth(input),
		evaluateDeployment(input),
		evaluateAlerts(input),
		evaluateIncidents(input),
		evaluateResourceUtilization(),
	}

	score, overallState := aggregate(factors)

	return Result{Score: score, OverallState: overallState, Factors: factors}
}

func evaluateAvailability(input Input) Factor {
	factor := Factor{Key: FactorAvailability, Label: "Availability", Weight: weightAvailability}

	if input.RuntimeDeploymentStatus == nil {
		factor.State = StateUnknown
		factor.Reason = "No live deployment status is available (cluster unreachable, or the application has never been deployed)."
		return factor
	}

	switch *input.RuntimeDeploymentStatus {
	case "Healthy":
		factor.State = StateHealthy
		factor.Score = intPtr(100)
		factor.Reason = "The live deployment reports all replicas ready and available."
	case "Progressing", "Scaling":
		factor.State = StateDegraded
		factor.Score = intPtr(60)
		factor.Reason = "The live deployment is currently progressing or scaling toward its desired replica count."
	case "Unavailable", "Failed":
		factor.State = StateCritical
		factor.Score = intPtr(0)
		factor.Reason = "The live deployment has no available replicas or has failed to progress."
	default:
		factor.State = StateUnknown
		factor.Reason = "The live deployment's status could not be determined."
	}

	return factor
}

func evaluateErrorRate() Factor {
	return Factor{
		Key:    FactorErrorRate,
		Label:  "Error Rate",
		Weight: weightErrorRate,
		State:  StateUnknown,
		Reason: "No APM or error-tracking integration is connected - error rate is not measured anywhere in OpsPilot today.",
	}
}

func evaluateLatency() Factor {
	return Factor{
		Key:    FactorLatency,
		Label:  "Latency",
		Weight: weightLatency,
		State:  StateUnknown,
		Reason: "No APM integration is connected - request latency is not measured anywhere in OpsPilot today.",
	}
}

func evaluateResourceUtilization() Factor {
	return Factor{
		Key:    FactorResourceUtilization,
		Label:  "Resource Utilization",
		Weight: weightResourceUtilization,
		State:  StateUnknown,
		Reason: "No per-application resource usage data source is available today (only point-in-time single-pod metrics exist, not an aggregate).",
	}
}

func evaluatePodHealth(input Input) Factor {
	factor := Factor{Key: FactorPodHealth, Label: "Pod Health", Weight: weightPodHealth}

	if !input.PodsAvailable {
		factor.State = StateUnknown
		factor.Reason = "Pod status could not be retrieved (cluster unreachable or no kubeconfig configured)."
		return factor
	}

	total := len(input.Pods)
	if total == 0 {
		factor.State = StateUnknown
		factor.Reason = "No pods were found for this application - this may mean it isn't running, not that it's healthy."
		return factor
	}

	ready := 0
	restarts := int32(0)
	for _, pod := range input.Pods {
		if pod.Ready {
			ready++
		}
		restarts += pod.RestartCount
	}

	switch {
	case ready == 0:
		factor.State = StateCritical
		factor.Score = intPtr(0)
		factor.Reason = "No pods are ready."
	case ready < total:
		factor.State = StateDegraded
		factor.Score = intPtr(roundRatio(ready, total))
		factor.Reason = plural(total-ready, "pod is", "pods are") + " not ready."
	case restarts > 0:
		factor.State = StateDegraded
		factor.Score = intPtr(70)
		factor.Reason = plural(int(restarts), "restart has", "restarts have") + " occurred, though all pods are currently ready."
	default:
		factor.State = StateHealthy
		factor.Score = intPtr(100)
		factor.Reason = "All pods are ready with no restarts."
	}

	return factor
}

func evaluateDeployment(input Input) Factor {
	factor := Factor{Key: FactorDeployment, Label: "Deployment", Weight: weightDeployment}

	if input.DeploymentStatus == nil {
		factor.State = StateUnknown
		factor.Reason = "This application has never been deployed."
		return factor
	}

	switch *input.DeploymentStatus {
	case "Succeeded":
		factor.State = StateHealthy
		factor.Score = intPtr(100)
		factor.Reason = "The most recent deployment completed successfully."
	case "Running", "Pending", "Queued", "RolledBack":
		factor.State = StateDegraded
		factor.Score = intPtr(60)
		factor.Reason = "The most recent deployment is still in progress or has not yet been confirmed successful."
	case "Cancelled":
		factor.State = StateDegraded
		factor.Score = intPtr(40)
		factor.Reason = "The most recent deployment was cancelled before completing."
	case "Failed":
		factor.State = StateCritical
		factor.Score = intPtr(0)
		factor.Reason = "The most recent deployment failed."
	default:
		factor.State = StateUnknown
		factor.Reason = "The most recent deployment's status is not recognized."
	}

	return factor
}

func evaluateAlerts(input Input) Factor {
	factor := Factor{Key: FactorAlerts, Label: "Active Alerts", Weight: weightAlerts}

	if input.ActiveAlertCount == nil {
		factor.State = StateUnknown
		factor.Reason = "Active alerts could not be determined."
		return factor
	}

	count := *input.ActiveAlertCount
	switch {
	case count >= alertsCriticalAt:
		factor.State = StateCritical
		factor.Score = intPtr(0)
		factor.Reason = plural(count, "active alert is", "active alerts are") + " currently attributed to this application."
	case count >= alertsDegradedAt:
		factor.State = StateDegraded
		factor.Score = intPtr(60)
		factor.Reason = plural(count, "active alert is", "active alerts are") + " currently attributed to this application."
	default:
		factor.State = StateHealthy
		factor.Score = intPtr(100)
		factor.Reason = "No active alerts are attributed to this application."
	}

	return factor
}

func evaluateIncidents(input Input) Factor {
	factor := Factor{Key: FactorIncidents, Label: "Incident Status", Weight: weightIncidents}

	if input.ActiveIncidentCount == nil {
		factor.State = StateUnknown
		factor.Reason = "Active incidents could not be determined."
		return factor
	}

	count := *input.ActiveIncidentCount
	if count >= incidentsCriticalAt {
		factor.State = StateCritical
		factor.Score = intPtr(0)
		factor.Reason = plural(count, "active incident is", "active incidents are") + " open for this application."
		return factor
	}

	factor.State = StateHealthy
	factor.Score = intPtr(100)
	factor.Reason = "No active incidents are open for this application."
	return factor
}

// aggregate computes the weighted average over every non-Unknown factor.
// Unknown factors are excluded from both the numerator and denominator -
// they contribute no evidence in either direction, rather than being
// treated as a worst-case 0 or a best-case 100.
func aggregate(factors []Factor) (*int, State) {
	totalWeight := 0
	weightedSum := 0

	for _, factor := range factors {
		if factor.State == StateUnknown || factor.Score == nil {
			continue
		}
		totalWeight += factor.Weight
		weightedSum += factor.Weight * (*factor.Score)
	}

	if totalWeight == 0 {
		return nil, StateUnknown
	}

	score := int(math.Round(float64(weightedSum) / float64(totalWeight)))
	return &score, stateForScore(score)
}

// stateForScore maps the overall 0-100 score to a State. Thresholds are
// fixed and documented here: 85+ Healthy, 60-84 Degraded, below 60 Critical.
func stateForScore(score int) State {
	switch {
	case score >= 85:
		return StateHealthy
	case score >= 60:
		return StateDegraded
	default:
		return StateCritical
	}
}

func roundRatio(numerator, denominator int) int {
	if denominator == 0 {
		return 0
	}
	return int(math.Round(float64(numerator) / float64(denominator) * 100))
}

func intPtr(value int) *int {
	return &value
}

func plural(count int, singular, pluralForm string) string {
	if count == 1 {
		return "1 " + singular
	}
	return strconv.Itoa(count) + " " + pluralForm
}
