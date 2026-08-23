package rca

import (
	"strings"
	"testing"
	"time"
)

func baseIncidentTime() time.Time {
	return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
}

func TestDetectDeploymentCorrelationStrongEvidence(t *testing.T) {
	incidentAt := baseIncidentTime()
	deployedAt := incidentAt.Add(-7 * time.Minute)

	input := Input{
		IncidentCreatedAt:    incidentAt,
		DeploymentsAvailable: true,
		Deployments: []DeploymentRecord{
			{ID: "dep-1", ImageTag: "v1.8.2", Status: "Succeeded", CreatedAt: deployedAt},
		},
		AlertsAvailable: true,
		Alerts: []AlertRecord{
			{ID: 1, Title: "high error rate", FirstSeenAt: deployedAt.Add(2 * time.Minute)},
			{ID: 2, Title: "high error rate", FirstSeenAt: deployedAt.Add(3 * time.Minute)},
		},
		PodsAvailable: true,
		Pods: []PodStatus{
			{Name: "api-0", Ready: true, RestartCount: 2},
		},
	}

	metricSignals := []MetricSignal{
		{MetricName: "memory_usage", Unit: "MiB", Before: 100, After: 300, ChangePercent: 200, Direction: "increased"},
	}
	logSignals := []LogSignal{
		{PodName: "api-0", Line: "OOMKilled", Keyword: "oomkilled"},
	}

	cause := detectDeploymentCorrelation(input, metricSignals, logSignals)
	if cause == nil {
		t.Fatal("expected a deployment correlation cause, got nil")
	}
	if cause.Category != "deployment" {
		t.Fatalf("expected category deployment, got %s", cause.Category)
	}
	if cause.Confidence != EvidenceLevelStrong {
		t.Fatalf("expected Strong confidence with 3 corroborating signals, got %s", cause.Confidence)
	}
	if len(cause.Evidence) < 4 {
		t.Fatalf("expected at least 4 evidence bullets (timing + alerts + metric + restarts + logs), got %d: %v", len(cause.Evidence), cause.Evidence)
	}
	if cause.Title == "" || cause.Title == "Root cause confirmed" {
		t.Fatalf("cause title must be hedged, got %q", cause.Title)
	}
}

func TestDetectDeploymentCorrelationOutsideLookbackWindowIsIgnored(t *testing.T) {
	incidentAt := baseIncidentTime()
	input := Input{
		IncidentCreatedAt:    incidentAt,
		DeploymentsAvailable: true,
		Deployments: []DeploymentRecord{
			{ID: "dep-old", ImageTag: "v1.0.0", CreatedAt: incidentAt.Add(-24 * time.Hour)},
		},
	}

	cause := detectDeploymentCorrelation(input, nil, nil)
	if cause != nil {
		t.Fatalf("expected no correlation for a deployment far outside the lookback window, got %+v", cause)
	}
}

func TestDetectDeploymentCorrelationIgnoresDeploymentsAfterIncident(t *testing.T) {
	incidentAt := baseIncidentTime()
	input := Input{
		IncidentCreatedAt:    incidentAt,
		DeploymentsAvailable: true,
		Deployments: []DeploymentRecord{
			{ID: "dep-after", ImageTag: "v2.0.0", CreatedAt: incidentAt.Add(5 * time.Minute)},
		},
	}

	cause := detectDeploymentCorrelation(input, nil, nil)
	if cause != nil {
		t.Fatalf("expected no correlation for a deployment that happened after the incident, got %+v", cause)
	}
}

func TestDetectPodHealthCauseFlagsCrashLoopingPods(t *testing.T) {
	input := Input{
		PodsAvailable: true,
		Pods: []PodStatus{
			{Name: "worker-0", Ready: false, RestartCount: 5, Reason: "CrashLoopBackOff"},
			{Name: "worker-1", Ready: true, RestartCount: 0},
		},
	}

	cause := detectPodHealthCause(input, nil)
	if cause == nil {
		t.Fatal("expected a pod health cause, got nil")
	}
	if cause.Category != "pod_health" {
		t.Fatalf("expected category pod_health, got %s", cause.Category)
	}
}

func TestDetectPodHealthCauseIgnoresHealthyPods(t *testing.T) {
	input := Input{
		PodsAvailable: true,
		Pods: []PodStatus{
			{Name: "worker-0", Ready: true, RestartCount: 0},
		},
	}

	if cause := detectPodHealthCause(input, nil); cause != nil {
		t.Fatalf("expected no pod health cause for healthy pods, got %+v", cause)
	}
}

func TestDetectConfigChangeCauseExcludesDeploymentEntity(t *testing.T) {
	input := Input{
		ConfigChangesAvailable: true,
		ConfigChanges: []ConfigChangeRecord{
			{EntityType: "deployment", EntityID: "dep-1", Action: "UPDATE", ChangedAt: baseIncidentTime()},
			{EntityType: "cluster", EntityID: "cl-1", Action: "UPDATE", FieldName: "kubeconfig", ChangedAt: baseIncidentTime()},
		},
	}

	cause := detectConfigChangeCause(input)
	if cause == nil {
		t.Fatal("expected a config change cause for the non-deployment entity")
	}
	for _, evidence := range cause.Evidence {
		if strings.Contains(evidence, "deployment dep-1") {
			t.Fatalf("deployment entity changes must be excluded from config change cause (covered separately), got %q", evidence)
		}
	}
}

func TestDetectResourceExhaustionCauseThreshold(t *testing.T) {
	below := []MetricSignal{{MetricName: "memory_usage", Before: 100, After: 120, ChangePercent: 20, Direction: "increased"}}
	if cause := detectResourceExhaustionCause(below); cause != nil {
		t.Fatalf("expected no cause below threshold, got %+v", cause)
	}

	above := []MetricSignal{{MetricName: "memory_usage", Before: 100, After: 150, ChangePercent: 50, Direction: "increased"}}
	cause := detectResourceExhaustionCause(above)
	if cause == nil {
		t.Fatal("expected a resource exhaustion cause above threshold")
	}
	if cause.Category != "resource_exhaustion" {
		t.Fatalf("expected category resource_exhaustion, got %s", cause.Category)
	}
}

func TestDetectKubernetesEventCauseFlagsWarningReasons(t *testing.T) {
	input := Input{
		K8sEventsAvailable: true,
		K8sEvents: []K8sEvent{
			{Reason: "Scheduled", Type: "Normal", Message: "assigned to node-1", InvolvedObject: "pod/api-0"},
			{Reason: "OOMKilling", Type: "Warning", Message: "memory limit exceeded", InvolvedObject: "pod/api-0"},
		},
	}

	cause := detectKubernetesEventCause(input)
	if cause == nil {
		t.Fatal("expected a kubernetes event cause")
	}
	if len(cause.Evidence) != 1 {
		t.Fatalf("expected exactly one flagged event (Normal events excluded), got %d", len(cause.Evidence))
	}
}
