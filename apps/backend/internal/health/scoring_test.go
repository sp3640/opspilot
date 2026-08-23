package health

import (
	"reflect"
	"testing"
)

func strPtr(value string) *string {
	return &value
}

func factorByKey(t *testing.T, factors []Factor, key FactorKey) Factor {
	t.Helper()
	for _, factor := range factors {
		if factor.Key == key {
			return factor
		}
	}
	t.Fatalf("factor %s not found", key)
	return Factor{}
}

func TestEvaluate_EmptyInputIsEntirelyUnknown(t *testing.T) {
	result := Evaluate(Input{})

	if result.Score != nil {
		t.Fatalf("expected nil score with no data at all, got %v", *result.Score)
	}
	if result.OverallState != StateUnknown {
		t.Fatalf("expected overall state Unknown, got %s", result.OverallState)
	}
	if len(result.Factors) != 8 {
		t.Fatalf("expected all 8 factors to always be reported, got %d", len(result.Factors))
	}
	for _, factor := range result.Factors {
		if factor.State != StateUnknown {
			t.Fatalf("factor %s: expected Unknown with no input, got %s", factor.Key, factor.State)
		}
		if factor.Score != nil {
			t.Fatalf("factor %s: expected nil score when Unknown, got %v", factor.Key, *factor.Score)
		}
		if factor.Reason == "" {
			t.Fatalf("factor %s: expected a non-empty reason even when Unknown", factor.Key)
		}
	}
}

func TestEvaluate_ErrorRateLatencyResourceUtilizationAreAlwaysUnknown(t *testing.T) {
	// Even with every other signal fully healthy, these three factors have
	// no real data source anywhere in the system and must never be
	// fabricated as Healthy.
	input := Input{
		RuntimeDeploymentStatus: strPtr("Healthy"),
		DeploymentStatus:        strPtr("Succeeded"),
		PodsAvailable:           true,
		Pods:                    []PodInput{{Ready: true}, {Ready: true}},
		ActiveAlertCount:        intPtr(0),
		ActiveIncidentCount:     intPtr(0),
	}

	result := Evaluate(input)

	for _, key := range []FactorKey{FactorErrorRate, FactorLatency, FactorResourceUtilization} {
		factor := factorByKey(t, result.Factors, key)
		if factor.State != StateUnknown {
			t.Fatalf("factor %s: expected always-Unknown, got %s", key, factor.State)
		}
	}
}

func TestEvaluate_AllHealthySignalsProduceHighScore(t *testing.T) {
	input := Input{
		RuntimeDeploymentStatus: strPtr("Healthy"),
		DeploymentStatus:        strPtr("Succeeded"),
		PodsAvailable:           true,
		Pods:                    []PodInput{{Ready: true}, {Ready: true}, {Ready: true}},
		ActiveAlertCount:        intPtr(0),
		ActiveIncidentCount:     intPtr(0),
	}

	result := Evaluate(input)

	if result.Score == nil {
		t.Fatalf("expected a real score")
	}
	if *result.Score != 100 {
		t.Fatalf("expected a perfect 100 when every measurable factor is healthy, got %d", *result.Score)
	}
	if result.OverallState != StateHealthy {
		t.Fatalf("expected overall Healthy, got %s", result.OverallState)
	}
}

func TestEvaluate_UnknownFactorsAreExcludedNotPenalized(t *testing.T) {
	// Only availability is known (Healthy); everything else is Unknown.
	// The score must be computed purely from availability's own weight,
	// not diluted by treating the unknown factors as failing.
	input := Input{RuntimeDeploymentStatus: strPtr("Healthy")}

	result := Evaluate(input)

	if result.Score == nil {
		t.Fatalf("expected a real score from the one known factor")
	}
	if *result.Score != 100 {
		t.Fatalf("expected 100 (availability alone is healthy), got %d", *result.Score)
	}
	if result.OverallState != StateHealthy {
		t.Fatalf("expected overall Healthy, got %s", result.OverallState)
	}
}

func TestEvaluate_Availability(t *testing.T) {
	cases := []struct {
		name   string
		status *string
		state  State
	}{
		{"nil is unknown", nil, StateUnknown},
		{"healthy", strPtr("Healthy"), StateHealthy},
		{"progressing is degraded", strPtr("Progressing"), StateDegraded},
		{"scaling is degraded", strPtr("Scaling"), StateDegraded},
		{"unavailable is critical", strPtr("Unavailable"), StateCritical},
		{"failed is critical", strPtr("Failed"), StateCritical},
		{"unrecognized is unknown", strPtr("SomethingElse"), StateUnknown},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Evaluate(Input{RuntimeDeploymentStatus: tc.status})
			factor := factorByKey(t, result.Factors, FactorAvailability)
			if factor.State != tc.state {
				t.Fatalf("expected %s, got %s", tc.state, factor.State)
			}
		})
	}
}

func TestEvaluate_Deployment(t *testing.T) {
	cases := []struct {
		name   string
		status *string
		state  State
	}{
		{"never deployed is unknown", nil, StateUnknown},
		{"succeeded is healthy", strPtr("Succeeded"), StateHealthy},
		{"running is degraded", strPtr("Running"), StateDegraded},
		{"pending is degraded", strPtr("Pending"), StateDegraded},
		{"queued is degraded", strPtr("Queued"), StateDegraded},
		{"rolled back is degraded", strPtr("RolledBack"), StateDegraded},
		{"cancelled is degraded", strPtr("Cancelled"), StateDegraded},
		{"failed is critical", strPtr("Failed"), StateCritical},
		{"unrecognized is unknown", strPtr("Bogus"), StateUnknown},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Evaluate(Input{DeploymentStatus: tc.status})
			factor := factorByKey(t, result.Factors, FactorDeployment)
			if factor.State != tc.state {
				t.Fatalf("expected %s, got %s", tc.state, factor.State)
			}
		})
	}
}

func TestEvaluate_PodHealth(t *testing.T) {
	t.Run("unavailable when pods could not be queried", func(t *testing.T) {
		result := Evaluate(Input{PodsAvailable: false})
		factor := factorByKey(t, result.Factors, FactorPodHealth)
		if factor.State != StateUnknown {
			t.Fatalf("expected Unknown, got %s", factor.State)
		}
	})

	t.Run("unknown (not healthy) when zero pods exist", func(t *testing.T) {
		result := Evaluate(Input{PodsAvailable: true, Pods: []PodInput{}})
		factor := factorByKey(t, result.Factors, FactorPodHealth)
		if factor.State != StateUnknown {
			t.Fatalf("expected Unknown for zero pods (not a fabricated Healthy), got %s", factor.State)
		}
	})

	t.Run("all ready no restarts is healthy", func(t *testing.T) {
		result := Evaluate(Input{PodsAvailable: true, Pods: []PodInput{{Ready: true}, {Ready: true}}})
		factor := factorByKey(t, result.Factors, FactorPodHealth)
		if factor.State != StateHealthy {
			t.Fatalf("expected Healthy, got %s", factor.State)
		}
		if factor.Score == nil || *factor.Score != 100 {
			t.Fatalf("expected score 100, got %v", factor.Score)
		}
	})

	t.Run("all ready with restarts is degraded", func(t *testing.T) {
		result := Evaluate(Input{PodsAvailable: true, Pods: []PodInput{{Ready: true, RestartCount: 3}}})
		factor := factorByKey(t, result.Factors, FactorPodHealth)
		if factor.State != StateDegraded {
			t.Fatalf("expected Degraded, got %s", factor.State)
		}
	})

	t.Run("some unready is degraded with proportional score", func(t *testing.T) {
		result := Evaluate(Input{PodsAvailable: true, Pods: []PodInput{{Ready: true}, {Ready: true}, {Ready: false}, {Ready: false}}})
		factor := factorByKey(t, result.Factors, FactorPodHealth)
		if factor.State != StateDegraded {
			t.Fatalf("expected Degraded, got %s", factor.State)
		}
		if factor.Score == nil || *factor.Score != 50 {
			t.Fatalf("expected score 50 (2/4 ready), got %v", factor.Score)
		}
	})

	t.Run("zero ready is critical", func(t *testing.T) {
		result := Evaluate(Input{PodsAvailable: true, Pods: []PodInput{{Ready: false}, {Ready: false}}})
		factor := factorByKey(t, result.Factors, FactorPodHealth)
		if factor.State != StateCritical {
			t.Fatalf("expected Critical, got %s", factor.State)
		}
		if factor.Score == nil || *factor.Score != 0 {
			t.Fatalf("expected score 0, got %v", factor.Score)
		}
	})
}

func TestEvaluate_Alerts(t *testing.T) {
	cases := []struct {
		name  string
		count *int
		state State
	}{
		{"nil is unknown", nil, StateUnknown},
		{"zero is healthy", intPtr(0), StateHealthy},
		{"one is degraded", intPtr(1), StateDegraded},
		{"two is critical", intPtr(2), StateCritical},
		{"many is critical", intPtr(9), StateCritical},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Evaluate(Input{ActiveAlertCount: tc.count})
			factor := factorByKey(t, result.Factors, FactorAlerts)
			if factor.State != tc.state {
				t.Fatalf("expected %s, got %s", tc.state, factor.State)
			}
		})
	}
}

func TestEvaluate_Incidents(t *testing.T) {
	cases := []struct {
		name  string
		count *int
		state State
	}{
		{"nil is unknown", nil, StateUnknown},
		{"zero is healthy", intPtr(0), StateHealthy},
		{"one is critical", intPtr(1), StateCritical},
		{"many is critical", intPtr(4), StateCritical},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Evaluate(Input{ActiveIncidentCount: tc.count})
			factor := factorByKey(t, result.Factors, FactorIncidents)
			if factor.State != tc.state {
				t.Fatalf("expected %s, got %s", tc.state, factor.State)
			}
		})
	}
}

func TestEvaluate_IsDeterministic(t *testing.T) {
	input := Input{
		RuntimeDeploymentStatus: strPtr("Progressing"),
		DeploymentStatus:        strPtr("Running"),
		PodsAvailable:           true,
		Pods:                    []PodInput{{Ready: true}, {Ready: false}},
		ActiveAlertCount:        intPtr(1),
		ActiveIncidentCount:     intPtr(0),
	}

	first := Evaluate(input)
	second := Evaluate(input)

	if *first.Score != *second.Score {
		t.Fatalf("expected identical scores for identical input, got %d and %d", *first.Score, *second.Score)
	}
	if first.OverallState != second.OverallState {
		t.Fatalf("expected identical overall state for identical input")
	}
	for i := range first.Factors {
		if !reflect.DeepEqual(first.Factors[i], second.Factors[i]) {
			t.Fatalf("expected identical factor at index %d for identical input, got %+v vs %+v", i, first.Factors[i], second.Factors[i])
		}
	}
}

func TestEvaluate_OverallStateThresholds(t *testing.T) {
	// A single critical factor (incidents) alongside otherwise-healthy
	// signals should pull the weighted score down out of Healthy territory,
	// proving the weighting actually has teeth rather than always
	// rounding back up to Healthy.
	input := Input{
		RuntimeDeploymentStatus: strPtr("Healthy"),
		DeploymentStatus:        strPtr("Succeeded"),
		PodsAvailable:           true,
		Pods:                    []PodInput{{Ready: true}},
		ActiveAlertCount:        intPtr(0),
		ActiveIncidentCount:     intPtr(1),
	}

	result := Evaluate(input)
	if result.OverallState == StateHealthy {
		t.Fatalf("expected an active incident to prevent an overall Healthy state, got score %v", result.Score)
	}
}
