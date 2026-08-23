package rca

import (
	"fmt"
	"sort"
	"strings"
)

// Evaluate computes the full RCA for one incident from already-fetched,
// real evidence. It is a pure function - deterministic for the same Input,
// safe to unit test without any I/O, and it never invents a signal that
// wasn't actually observed.
func Evaluate(input Input) Result {
	splitAt := input.IncidentCreatedAt
	for _, deployment := range input.Deployments {
		if deployment.CreatedAt.Before(input.IncidentCreatedAt) {
			splitAt = deployment.CreatedAt
			break
		}
	}

	var metricSignals []MetricSignal
	if input.MetricsAvailable {
		metricSignals = computeMetricSignals(input.Metrics, splitAt)
	}

	var logSignals []LogSignal
	if input.LogsAvailable {
		logSignals = extractLogSignals(input.Logs)
	}

	causes := collectCauses(input, metricSignals, logSignals)
	timeline := buildTimeline(input)

	evidenceLevel, evidenceLevelReason := aggregateEvidenceLevel(input, causes)

	return Result{
		Summary:                       buildSummary(input, causes),
		AffectedApplication:           input.ApplicationName,
		ApplicationKnown:              input.ApplicationAvailable,
		Timeline:                      timeline,
		RecentChanges:                 input.ConfigChanges,
		CorrelatedAlerts:              input.Alerts,
		RelevantMetrics:               metricSignals,
		RelevantLogs:                  logSignals,
		KubernetesEvidence:            input.K8sEvents,
		PodEvidence:                   input.Pods,
		PossibleCauses:                causes,
		RecommendedInvestigationSteps: buildInvestigationSteps(input, causes),
		RecommendedRemediation:        buildRemediationSteps(causes),
		EvidenceLevel:                 evidenceLevel,
		EvidenceLevelReason:           evidenceLevelReason,
	}
}

func collectCauses(input Input, metricSignals []MetricSignal, logSignals []LogSignal) []PossibleCause {
	causes := make([]PossibleCause, 0, 5)

	if cause := detectDeploymentCorrelation(input, metricSignals, logSignals); cause != nil {
		causes = append(causes, *cause)
	}
	if cause := detectPodHealthCause(input, logSignals); cause != nil {
		causes = append(causes, *cause)
	}
	if cause := detectConfigChangeCause(input); cause != nil {
		causes = append(causes, *cause)
	}
	if cause := detectResourceExhaustionCause(metricSignals); cause != nil {
		causes = append(causes, *cause)
	}
	if cause := detectKubernetesEventCause(input); cause != nil {
		causes = append(causes, *cause)
	}

	sort.SliceStable(causes, func(i, j int) bool {
		return confidenceRank(causes[i].Confidence) > confidenceRank(causes[j].Confidence)
	})

	return causes
}

func confidenceRank(level EvidenceLevel) int {
	switch level {
	case EvidenceLevelStrong:
		return 3
	case EvidenceLevelModerate:
		return 2
	case EvidenceLevelWeak:
		return 1
	default:
		return 0
	}
}

// aggregateEvidenceLevel reports how much real evidence backs this
// analysis overall: the strongest individual cause's confidence when any
// cause was found, or Insufficient (with an honest reason) when none was -
// never inflated by the mere presence of unrelated data.
func aggregateEvidenceLevel(input Input, causes []PossibleCause) (EvidenceLevel, string) {
	availableCategories := countAvailable(input)

	if len(causes) == 0 {
		if availableCategories < 2 {
			return EvidenceLevelInsufficient, "Too few evidence categories were available (alerts, metrics, logs, deployments, Kubernetes events, pod health, config changes) to identify a possible cause."
		}
		return EvidenceLevelInsufficient, "Evidence was available but no signals in the incident window correlate with each other; no possible cause could be identified."
	}

	strongest := causes[0].Confidence
	switch strongest {
	case EvidenceLevelStrong:
		return EvidenceLevelStrong, "Multiple independent evidence categories corroborate the same possible cause."
	case EvidenceLevelModerate:
		return EvidenceLevelModerate, "At least one evidence category corroborates the leading possible cause."
	default:
		return EvidenceLevelWeak, "A signal was found, but nothing else in the available evidence corroborates it."
	}
}

func countAvailable(input Input) int {
	count := 0
	for _, available := range []bool{
		input.AlertsAvailable,
		input.MetricsAvailable,
		input.LogsAvailable,
		input.DeploymentsAvailable,
		input.K8sEventsAvailable,
		input.PodsAvailable,
		input.ConfigChangesAvailable,
	} {
		if available {
			count++
		}
	}
	return count
}

func buildSummary(input Input, causes []PossibleCause) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "%s severity incident %q is currently %s.", input.IncidentSeverity, input.IncidentTitle, strings.ToLower(input.IncidentStatus))

	if input.ApplicationAvailable && input.ApplicationName != "" {
		fmt.Fprintf(&builder, " It affects the %s application.", input.ApplicationName)
	}

	switch len(causes) {
	case 0:
		builder.WriteString(" No possible cause could be identified from the available evidence yet.")
	case 1:
		builder.WriteString(" Evidence suggests one possible cause; see below.")
	default:
		fmt.Fprintf(&builder, " Evidence suggests %d possible causes, ranked by how well they are corroborated.", len(causes))
	}

	return builder.String()
}

func buildInvestigationSteps(input Input, causes []PossibleCause) []string {
	steps := make([]string, 0, 6)
	steps = append(steps, "Review the timeline to confirm the order in which alerts, deployments, and changes occurred relative to the incident.")

	categories := make(map[string]bool)
	for _, cause := range causes {
		categories[cause.Category] = true
	}

	if categories["deployment"] {
		steps = append(steps, "Compare application behavior immediately before and after the correlated deployment.")
	}
	if categories["pod_health"] {
		steps = append(steps, "Inspect pod logs and run `kubectl describe pod` on the affected pod(s) to confirm the crash/restart reason.")
	}
	if categories["config_change"] {
		steps = append(steps, "Review the flagged configuration change with whoever made it to confirm whether it was intentional and related.")
	}
	if categories["resource_exhaustion"] {
		steps = append(steps, "Check resource requests/limits and current utilization for the affected workload.")
	}
	if categories["kubernetes_event"] {
		steps = append(steps, "Review the flagged Kubernetes events for the affected object(s) in more detail.")
	}

	if !input.MetricsAvailable {
		steps = append(steps, "No metrics were available for this window; connect metrics collection for this application to improve future analysis.")
	}
	if !input.LogsAvailable {
		steps = append(steps, "No pod logs were available for this window; verify the cluster connection and pod status for log collection.")
	}

	return steps
}

func buildRemediationSteps(causes []PossibleCause) []string {
	steps := make([]string, 0, len(causes)+1)

	for _, cause := range causes {
		switch cause.Category {
		case "deployment":
			steps = append(steps, "Investigate the correlated deployment, and consider rolling it back if the correlation is confirmed.")
		case "pod_health":
			steps = append(steps, "Investigate the crash-looping/unhealthy pod(s); consider a rollout restart once the underlying cause is confirmed.")
		case "config_change":
			steps = append(steps, "Investigate the flagged configuration change; consider reverting it if it is confirmed to be related.")
		case "resource_exhaustion":
			steps = append(steps, "Investigate resource limits/requests for the affected workload; consider scaling or adjusting limits if confirmed.")
		case "kubernetes_event":
			steps = append(steps, "Investigate the flagged Kubernetes events; remediation depends on the specific reason reported.")
		}
	}

	if len(steps) == 0 {
		steps = append(steps, "No possible cause was identified yet, so no remediation can be recommended - continue investigating using the evidence above.")
	}

	steps = append(steps, "Any remediation must be reviewed and applied by a human engineer; this analysis does not authorize automated action.")

	return steps
}
