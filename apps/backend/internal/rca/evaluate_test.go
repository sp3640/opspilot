package rca

import (
	"strings"
	"testing"
	"time"
)

// TestEvaluateInsufficientEvidenceWhenNothingIsAvailable is the "honest
// zero" case: no evidence category could be queried at all, so Evaluate
// must never invent a possible cause or claim more confidence than it has.
func TestEvaluateInsufficientEvidenceWhenNothingIsAvailable(t *testing.T) {
	input := Input{
		IncidentID:        1,
		IncidentTitle:     "unexplained incident",
		IncidentSeverity:  "P2",
		IncidentStatus:    "OPEN",
		IncidentCreatedAt: baseIncidentTime(),
	}

	result := Evaluate(input)

	if result.EvidenceLevel != EvidenceLevelInsufficient {
		t.Fatalf("expected Insufficient evidence level, got %s", result.EvidenceLevel)
	}
	if len(result.PossibleCauses) != 0 {
		t.Fatalf("expected zero possible causes with no evidence available, got %d", len(result.PossibleCauses))
	}
	if len(result.Timeline) != 1 {
		t.Fatalf("expected only the incident-created timeline event, got %d", len(result.Timeline))
	}
}

// TestEvaluateDeploymentScenarioFromSpec reproduces the exact example given
// in the RCA spec: a deployment 7 minutes before degradation, followed by a
// rising error rate, rising memory, pod restarts, and OOM-flavored logs.
// The result must read as a hedged possible cause, never a confirmed one.
func TestEvaluateDeploymentScenarioFromSpec(t *testing.T) {
	incidentAt := baseIncidentTime()
	deployedAt := incidentAt.Add(-7 * time.Minute)

	input := Input{
		IncidentID:           7,
		IncidentTitle:        "checkout latency spike",
		IncidentSeverity:     "P1",
		IncidentStatus:       "INVESTIGATING",
		IncidentCreatedAt:    incidentAt,
		ApplicationAvailable: true,
		ApplicationName:      "checkout-service",

		AlertsAvailable: true,
		Alerts: []AlertRecord{
			{ID: 1, Title: "elevated 5xx rate", FirstSeenAt: deployedAt.Add(2 * time.Minute)},
			{ID: 2, Title: "elevated 5xx rate", FirstSeenAt: deployedAt.Add(4 * time.Minute)},
			{ID: 3, Title: "elevated 5xx rate", FirstSeenAt: deployedAt.Add(6 * time.Minute)},
		},

		MetricsAvailable: true,
		Metrics: []MetricPoint{
			{MetricType: "RESOURCE", MetricName: "memory_usage", Unit: "MiB", Value: 200, Timestamp: deployedAt.Add(-5 * time.Minute)},
			{MetricType: "RESOURCE", MetricName: "memory_usage", Unit: "MiB", Value: 210, Timestamp: deployedAt.Add(-2 * time.Minute)},
			{MetricType: "RESOURCE", MetricName: "memory_usage", Unit: "MiB", Value: 480, Timestamp: deployedAt.Add(3 * time.Minute)},
			{MetricType: "RESOURCE", MetricName: "memory_usage", Unit: "MiB", Value: 510, Timestamp: deployedAt.Add(6 * time.Minute)},
		},

		LogsAvailable: true,
		Logs: []LogLine{
			{PodName: "checkout-0", Line: "2026-01-01T11:56:00Z INFO handling request"},
			{PodName: "checkout-0", Line: "2026-01-01T11:57:00Z ERROR OOMKilled: container checkout exceeded its memory limit"},
		},

		DeploymentsAvailable: true,
		Deployments: []DeploymentRecord{
			{ID: "dep-abc", ImageTag: "v1.8.2", Status: "Succeeded", Environment: "production", CreatedAt: deployedAt},
		},

		PodsAvailable: true,
		Pods: []PodStatus{
			{Name: "checkout-0", Ready: true, RestartCount: 3, Reason: "OOMKilled"},
		},
	}

	result := Evaluate(input)

	if len(result.PossibleCauses) == 0 {
		t.Fatal("expected at least one possible cause")
	}

	leading := result.PossibleCauses[0]
	if leading.Category != "deployment" {
		t.Fatalf("expected the deployment cause to rank first, got category %s", leading.Category)
	}
	if leading.Confidence != EvidenceLevelStrong {
		t.Fatalf("expected Strong confidence given 3 corroborating signals, got %s", leading.Confidence)
	}
	if result.EvidenceLevel != EvidenceLevelStrong {
		t.Fatalf("expected overall evidence level Strong, got %s", result.EvidenceLevel)
	}

	assertNeverClaimsCertainty(t, result)
}

// TestEvaluateNeverClaimsCertaintyAcrossScenarios guards the spec's central
// requirement - "never claim certainty when the evidence only shows
// correlation" - across every text field Evaluate produces, for a handful
// of representative scenarios (not just the strongest one).
func TestEvaluateNeverClaimsCertaintyAcrossScenarios(t *testing.T) {
	incidentAt := baseIncidentTime()

	scenarios := map[string]Input{
		"no evidence": {
			IncidentID: 1, IncidentTitle: "mystery", IncidentSeverity: "P3", IncidentStatus: "OPEN", IncidentCreatedAt: incidentAt,
		},
		"deployment only": {
			IncidentID: 2, IncidentTitle: "deploy incident", IncidentSeverity: "P2", IncidentStatus: "OPEN", IncidentCreatedAt: incidentAt,
			DeploymentsAvailable: true,
			Deployments:          []DeploymentRecord{{ID: "d1", ImageTag: "v2.0.0", CreatedAt: incidentAt.Add(-5 * time.Minute)}},
		},
		"config change only": {
			IncidentID: 3, IncidentTitle: "config incident", IncidentSeverity: "P2", IncidentStatus: "OPEN", IncidentCreatedAt: incidentAt,
			ConfigChangesAvailable: true,
			ConfigChanges:          []ConfigChangeRecord{{EntityType: "cluster", EntityID: "c1", Action: "UPDATE", ChangedAt: incidentAt.Add(-2 * time.Minute)}},
		},
	}

	for name, input := range scenarios {
		t.Run(name, func(t *testing.T) {
			assertNeverClaimsCertainty(t, Evaluate(input))
		})
	}
}

func assertNeverClaimsCertainty(t *testing.T, result Result) {
	t.Helper()

	forbidden := []string{"root cause confirmed", "confirmed root cause", "definitely caused", "guaranteed"}
	texts := []string{result.Summary, result.EvidenceLevelReason}
	for _, cause := range result.PossibleCauses {
		texts = append(texts, cause.Title)
		texts = append(texts, cause.Evidence...)
	}
	texts = append(texts, result.RecommendedInvestigationSteps...)
	texts = append(texts, result.RecommendedRemediation...)

	for _, text := range texts {
		lower := strings.ToLower(text)
		for _, phrase := range forbidden {
			if strings.Contains(lower, phrase) {
				t.Fatalf("RCA output must never claim certainty; found forbidden phrase %q in %q", phrase, text)
			}
		}
	}
}

func TestEvaluateRecommendedRemediationAlwaysRequiresHumanApproval(t *testing.T) {
	result := Evaluate(Input{IncidentID: 1, IncidentTitle: "x", IncidentSeverity: "P3", IncidentStatus: "OPEN", IncidentCreatedAt: baseIncidentTime()})

	found := false
	for _, step := range result.RecommendedRemediation {
		if strings.Contains(step, "reviewed and applied by a human engineer") {
			found = true
		}
	}
	if !found {
		t.Fatal("expected remediation steps to always include the human-approval requirement")
	}
}
