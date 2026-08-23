package rca

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// deploymentLookback bounds how far before the incident a deployment can
// have occurred and still be considered for correlation - a deployment from
// three days ago is very unlikely to explain an incident that just started.
const deploymentLookback = 3 * time.Hour

// podRestartThreshold is the restart count above which a pod is considered
// crash-looping rather than having had an isolated restart.
const podRestartThreshold = int32(3)

// resourceMetricChangeThreshold is the minimum percent increase in a
// resource-utilization-shaped metric within the window before it is
// reported as a possible cause on its own (independent of any deployment).
const resourceMetricChangeThreshold = 30.0

// warningEventReasons is the fixed set of Kubernetes event reasons treated
// as evidence of a real problem, not routine scheduling noise.
var warningEventReasons = map[string]bool{
	"OOMKilling":       true,
	"OOMKilled":        true,
	"BackOff":          true,
	"CrashLoopBackOff": true,
	"Failed":           true,
	"FailedScheduling": true,
	"Unhealthy":        true,
	"FailedMount":      true,
	"NodeNotReady":     true,
	"Evicted":          true,
}

// resourceMetricNames identifies metric names this package treats as
// resource-utilization signals for the standalone resource-exhaustion cause.
var resourceMetricNames = map[string]bool{
	"memory_usage":       true,
	"memory_utilization": true,
	"cpu_usage":          true,
	"cpu_utilization":    true,
}

// detectDeploymentCorrelation looks for the deployment closest to (but
// before) the incident's creation time, within deploymentLookback, and
// gathers every corroborating signal already computed for the window: an
// alert-volume increase after it, a rising resource metric, pod restarts,
// and OOM/crash-flavored log lines. Mirrors the example in the RCA spec
// almost verbatim: "deployment occurred N minutes before degradation".
func detectDeploymentCorrelation(input Input, metricSignals []MetricSignal, logSignals []LogSignal) *PossibleCause {
	if !input.DeploymentsAvailable || len(input.Deployments) == 0 {
		return nil
	}

	var closest *DeploymentRecord
	var closestGap time.Duration
	for i := range input.Deployments {
		deployment := input.Deployments[i]
		if deployment.CreatedAt.After(input.IncidentCreatedAt) {
			continue
		}
		gap := input.IncidentCreatedAt.Sub(deployment.CreatedAt)
		if gap > deploymentLookback {
			continue
		}
		if closest == nil || gap < closestGap {
			closest = &input.Deployments[i]
			closestGap = gap
		}
	}

	if closest == nil {
		return nil
	}

	minutesBefore := closestGap.Minutes()
	evidence := []string{
		fmt.Sprintf("Deployment %s occurred %s before the incident was created.", deploymentLabel(*closest), formatMinutes(minutesBefore)),
	}
	corroborating := 0

	if increased, before, after := alertVolumeIncreasedAfter(input.Alerts, closest.CreatedAt); increased {
		evidence = append(evidence, fmt.Sprintf("Alert volume increased after this deployment (%d before vs %d after).", before, after))
		corroborating++
	}

	for _, signal := range metricSignals {
		if signal.Direction != "increased" {
			continue
		}
		evidence = append(evidence, fmt.Sprintf("%s increased after this window's midpoint (%.2f -> %.2f %s, %+.1f%%).", metricLabel(signal), signal.Before, signal.After, signal.Unit, signal.ChangePercent))
		corroborating++
	}

	if restarted, count := podsRestartedAfter(input.Pods); restarted {
		evidence = append(evidence, fmt.Sprintf("%d pod(s) have restart counts consistent with crashing after this deployment.", count))
		corroborating++
	}

	if len(logSignals) > 0 {
		evidence = append(evidence, fmt.Sprintf("Logs contain %s.", summarizeLogKeywords(logSignals)))
		corroborating++
	}

	return &PossibleCause{
		Title:      fmt.Sprintf("Deployment %s may be related to the incident.", deploymentLabel(*closest)),
		Category:   "deployment",
		Confidence: confidenceForCorroboration(corroborating),
		Evidence:   evidence,
	}
}

// detectPodHealthCause flags pods that are not ready or are crash-looping
// (restart count at/above podRestartThreshold) as a standalone possible
// cause, independent of whether a deployment correlates with the incident.
func detectPodHealthCause(input Input, logSignals []LogSignal) *PossibleCause {
	if !input.PodsAvailable || len(input.Pods) == 0 {
		return nil
	}

	var unhealthy []PodStatus
	for _, pod := range input.Pods {
		if !pod.Ready || pod.RestartCount >= podRestartThreshold {
			unhealthy = append(unhealthy, pod)
		}
	}
	if len(unhealthy) == 0 {
		return nil
	}

	evidence := make([]string, 0, len(unhealthy)+1)
	for _, pod := range unhealthy {
		switch {
		case !pod.Ready && pod.RestartCount > 0:
			evidence = append(evidence, fmt.Sprintf("Pod %s is not ready and has restarted %d time(s)%s.", pod.Name, pod.RestartCount, reasonSuffix(pod.Reason)))
		case !pod.Ready:
			evidence = append(evidence, fmt.Sprintf("Pod %s is not ready%s.", pod.Name, reasonSuffix(pod.Reason)))
		default:
			evidence = append(evidence, fmt.Sprintf("Pod %s has restarted %d times, suggesting a crash loop.", pod.Name, pod.RestartCount))
		}
	}

	corroborating := 0
	if len(logSignals) > 0 {
		evidence = append(evidence, fmt.Sprintf("Logs contain %s.", summarizeLogKeywords(logSignals)))
		corroborating++
	}

	confidence := EvidenceLevelWeak
	if len(unhealthy) > 1 {
		corroborating++
	}
	confidence = confidenceForCorroboration(corroborating)

	return &PossibleCause{
		Title:      "Repeated pod restarts or crash-looping may be contributing to the incident.",
		Category:   "pod_health",
		Confidence: confidence,
		Evidence:   evidence,
	}
}

// detectConfigChangeCause flags any audit-logged change (to any entity
// other than the deployment/incident themselves, which are already
// evaluated separately) that occurred inside the evidence window as a
// possible cause - "evidence suggests" language only, since an audit log
// entry only proves a change happened, not that it caused anything.
func detectConfigChangeCause(input Input) *PossibleCause {
	if !input.ConfigChangesAvailable || len(input.ConfigChanges) == 0 {
		return nil
	}

	relevant := make([]ConfigChangeRecord, 0, len(input.ConfigChanges))
	for _, change := range input.ConfigChanges {
		// Deployments are already evaluated as their own, more specific
		// possible cause; incident entity changes are the incident's own
		// lifecycle (status/severity/assignment), not a cause of it.
		if change.EntityType == "deployment" || change.EntityType == "incident" {
			continue
		}
		relevant = append(relevant, change)
	}
	if len(relevant) == 0 {
		return nil
	}

	sort.Slice(relevant, func(i, j int) bool { return relevant[i].ChangedAt.Before(relevant[j].ChangedAt) })

	evidence := make([]string, 0, len(relevant))
	for _, change := range relevant {
		evidence = append(evidence, fmt.Sprintf(
			"%s change to %s %s at %s%s.",
			strings.ToLower(change.Action), change.EntityType, change.EntityID,
			change.ChangedAt.Format(time.RFC3339), fieldSuffix(change),
		))
	}

	confidence := EvidenceLevelWeak
	if len(relevant) > 1 {
		confidence = EvidenceLevelModerate
	}

	return &PossibleCause{
		Title:      "A recent configuration change may be related to the incident.",
		Category:   "config_change",
		Confidence: confidence,
		Evidence:   evidence,
	}
}

// detectResourceExhaustionCause flags a sharp rise in a resource-shaped
// metric (memory/CPU usage or utilization) as a standalone possible cause
// when it isn't already captured by the deployment-correlation cause.
func detectResourceExhaustionCause(metricSignals []MetricSignal) *PossibleCause {
	var flagged []MetricSignal
	for _, signal := range metricSignals {
		if !resourceMetricNames[strings.ToLower(signal.MetricName)] {
			continue
		}
		if signal.Direction == "increased" && signal.ChangePercent >= resourceMetricChangeThreshold {
			flagged = append(flagged, signal)
		}
	}
	if len(flagged) == 0 {
		return nil
	}

	evidence := make([]string, 0, len(flagged))
	for _, signal := range flagged {
		evidence = append(evidence, fmt.Sprintf("%s rose sharply within the incident window (%.2f -> %.2f %s, %+.1f%%).", metricLabel(signal), signal.Before, signal.After, signal.Unit, signal.ChangePercent))
	}

	confidence := EvidenceLevelWeak
	if len(flagged) > 1 {
		confidence = EvidenceLevelModerate
	}

	return &PossibleCause{
		Title:      "Resource utilization increased sharply during the incident window.",
		Category:   "resource_exhaustion",
		Confidence: confidence,
		Evidence:   evidence,
	}
}

// detectKubernetesEventCause flags Kubernetes Warning events with a
// recognized problem reason (OOMKilled, BackOff, FailedScheduling, ...) as
// a possible cause.
func detectKubernetesEventCause(input Input) *PossibleCause {
	if !input.K8sEventsAvailable || len(input.K8sEvents) == 0 {
		return nil
	}

	var flagged []K8sEvent
	for _, event := range input.K8sEvents {
		if event.Type == "Warning" && warningEventReasons[event.Reason] {
			flagged = append(flagged, event)
		}
	}
	if len(flagged) == 0 {
		return nil
	}

	evidence := make([]string, 0, len(flagged))
	for _, event := range flagged {
		evidence = append(evidence, fmt.Sprintf("%s: %s (%s, seen %d time(s))%s.", event.Reason, event.Message, event.InvolvedObject, event.Count, lastSeenSuffix(event.LastTimestamp)))
	}

	confidence := EvidenceLevelWeak
	if len(flagged) > 1 {
		confidence = EvidenceLevelModerate
	}

	return &PossibleCause{
		Title:      "Kubernetes reported warning events consistent with a real problem in the cluster.",
		Category:   "kubernetes_event",
		Confidence: confidence,
		Evidence:   evidence,
	}
}

func confidenceForCorroboration(corroboratingSignals int) EvidenceLevel {
	switch {
	case corroboratingSignals >= 2:
		return EvidenceLevelStrong
	case corroboratingSignals == 1:
		return EvidenceLevelModerate
	default:
		return EvidenceLevelWeak
	}
}

func alertVolumeIncreasedAfter(alerts []AlertRecord, at time.Time) (bool, int, int) {
	before, after := 0, 0
	for _, alert := range alerts {
		if alert.FirstSeenAt.Before(at) {
			before++
		} else {
			after++
		}
	}
	return after > before, before, after
}

func podsRestartedAfter(pods []PodStatus) (bool, int) {
	count := 0
	for _, pod := range pods {
		if pod.RestartCount > 0 {
			count++
		}
	}
	return count > 0, count
}

func summarizeLogKeywords(signals []LogSignal) string {
	seen := make(map[string]bool)
	keywords := make([]string, 0, len(signals))
	for _, signal := range signals {
		if seen[signal.Keyword] {
			continue
		}
		seen[signal.Keyword] = true
		keywords = append(keywords, signal.Keyword)
	}
	sort.Strings(keywords)
	return strings.Join(keywords, ", ") + "-related errors"
}

func deploymentLabel(deployment DeploymentRecord) string {
	if deployment.ImageTag != "" {
		return deployment.ImageTag
	}
	return deployment.ID
}

func metricLabel(signal MetricSignal) string {
	if signal.MetricName != "" {
		return signal.MetricName
	}
	return signal.MetricType
}

func formatMinutes(minutes float64) string {
	if minutes < 1 {
		return "less than a minute"
	}
	rounded := int(minutes + 0.5)
	if rounded == 1 {
		return "1 minute"
	}
	if rounded < 60 {
		return fmt.Sprintf("%d minutes", rounded)
	}
	hours := rounded / 60
	remMinutes := rounded % 60
	if remMinutes == 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, remMinutes)
}

func reasonSuffix(reason string) string {
	if reason == "" {
		return ""
	}
	return fmt.Sprintf(" (%s)", reason)
}

func fieldSuffix(change ConfigChangeRecord) string {
	if change.FieldName == "" {
		return ""
	}
	return fmt.Sprintf(" (%s: %q -> %q)", change.FieldName, change.OldValue, change.NewValue)
}

func lastSeenSuffix(lastTimestamp *time.Time) string {
	if lastTimestamp == nil {
		return ""
	}
	return ", last seen " + lastTimestamp.Format(time.RFC3339)
}
