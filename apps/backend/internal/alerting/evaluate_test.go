package alerting

import (
	"testing"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/metrics"
)

func findCondition(t *testing.T, conditions []EvaluatedCondition, resourceID, conditionKey string) *EvaluatedCondition {
	t.Helper()
	for i := range conditions {
		if conditions[i].ResourceID == resourceID && conditions[i].ConditionKey == conditionKey {
			return &conditions[i]
		}
	}
	return nil
}

func TestEvaluate_CPUThreshold(t *testing.T) {
	input := EvaluateInput{
		ClusterID: "cluster-1", ClusterName: "prod",
		Snapshot: metrics.CollectedMetrics{
			Cluster: metrics.ClusterMetrics{CPUCapacityMilliCores: 1000, CPUUsageMilliCores: 950},
		},
	}

	results := Evaluate(input)

	condition := findCondition(t, results, "cluster-1", ConditionCPUThreshold)
	if condition == nil {
		t.Fatalf("expected a CPU threshold condition, got %+v", results)
	}
	if condition.Severity != constants.AlertSeverityCritical {
		t.Fatalf("expected CRITICAL severity at 95%%, got %s", condition.Severity)
	}
	if condition.Title != "High CPU usage on cluster prod" {
		t.Fatalf("expected a stable title, got %q", condition.Title)
	}
}

func TestEvaluate_CPUBelowThresholdProducesNothing(t *testing.T) {
	input := EvaluateInput{
		ClusterID: "cluster-1", ClusterName: "prod",
		Snapshot: metrics.CollectedMetrics{
			Cluster: metrics.ClusterMetrics{CPUCapacityMilliCores: 1000, CPUUsageMilliCores: 100},
		},
	}

	results := Evaluate(input)
	if condition := findCondition(t, results, "cluster-1", ConditionCPUThreshold); condition != nil {
		t.Fatalf("expected no CPU condition at 10%% usage, got %+v", condition)
	}
}

func TestEvaluate_TitleStaysStableAcrossDifferentValues(t *testing.T) {
	// Two evaluations at different (but both HIGH-severity) usage levels must
	// produce an identical title - the title feeds Alert deduplication
	// upstream, so embedding the current value here would defeat dedup.
	first := Evaluate(EvaluateInput{ClusterID: "c1", ClusterName: "prod", Snapshot: metrics.CollectedMetrics{
		Cluster: metrics.ClusterMetrics{CPUCapacityMilliCores: 1000, CPUUsageMilliCores: 820},
	}})
	second := Evaluate(EvaluateInput{ClusterID: "c1", ClusterName: "prod", Snapshot: metrics.CollectedMetrics{
		Cluster: metrics.ClusterMetrics{CPUCapacityMilliCores: 1000, CPUUsageMilliCores: 880},
	}})

	a := findCondition(t, first, "c1", ConditionCPUThreshold)
	b := findCondition(t, second, "c1", ConditionCPUThreshold)
	if a == nil || b == nil {
		t.Fatalf("expected both evaluations to fire, got %+v / %+v", a, b)
	}
	if a.Title != b.Title {
		t.Fatalf("expected stable title, got %q vs %q", a.Title, b.Title)
	}
	if a.Description == b.Description {
		t.Fatalf("expected the description to reflect the different current values")
	}
}

func TestEvaluate_NodeNotReady(t *testing.T) {
	input := EvaluateInput{
		ClusterID: "cluster-1", ClusterName: "prod",
		Snapshot: metrics.CollectedMetrics{
			Nodes: []metrics.NodeMetrics{
				{Name: "node-a", Ready: true},
				{Name: "node-b", Ready: false},
			},
		},
	}

	results := Evaluate(input)

	if findCondition(t, results, "node-a", ConditionNodeNotReady) != nil {
		t.Fatalf("expected node-a (ready) not to fire")
	}
	condition := findCondition(t, results, "node-b", ConditionNodeNotReady)
	if condition == nil {
		t.Fatalf("expected node-b (not ready) to fire, got %+v", results)
	}
	if condition.Severity != constants.AlertSeverityHigh {
		t.Fatalf("expected HIGH severity, got %s", condition.Severity)
	}
}

func TestEvaluate_PodRestartCountAndCrashLoopBackOffAreDistinctConditions(t *testing.T) {
	input := EvaluateInput{
		ClusterID: "cluster-1", ClusterName: "prod",
		Snapshot: metrics.CollectedMetrics{
			Pods: []metrics.PodMetrics{
				{Name: "web-1", Namespace: "default", Restarts: 12, ContainerState: "CrashLoopBackOff"},
				{Name: "web-2", Namespace: "default", Restarts: 1, ContainerState: "Running"},
			},
		},
	}

	results := Evaluate(input)

	restartCondition := findCondition(t, results, "default/web-1", ConditionPodRestartCount)
	if restartCondition == nil || restartCondition.Severity != constants.AlertSeverityCritical {
		t.Fatalf("expected a CRITICAL restart-count condition for web-1, got %+v", restartCondition)
	}

	crashCondition := findCondition(t, results, "default/web-1", ConditionCrashLoopBackOff)
	if crashCondition == nil || crashCondition.Severity != constants.AlertSeverityCritical {
		t.Fatalf("expected a CrashLoopBackOff condition for web-1, got %+v", crashCondition)
	}

	if findCondition(t, results, "default/web-2", ConditionPodRestartCount) != nil {
		t.Fatalf("expected web-2 (1 restart, healthy) not to fire a restart condition")
	}
	if findCondition(t, results, "default/web-2", ConditionCrashLoopBackOff) != nil {
		t.Fatalf("expected web-2 not to fire a CrashLoopBackOff condition")
	}
}

func TestEvaluate_PropagatesApplicationIDWhenLabeled(t *testing.T) {
	input := EvaluateInput{
		ClusterID: "cluster-1", ClusterName: "prod",
		Snapshot: metrics.CollectedMetrics{
			Pods: []metrics.PodMetrics{
				{Name: "web-1", Namespace: "default", Restarts: 12, ContainerState: "CrashLoopBackOff", ApplicationID: "app-123"},
				{Name: "web-2", Namespace: "default", Restarts: 12, ContainerState: "CrashLoopBackOff"},
			},
			Deployments: []metrics.DeploymentMetrics{
				{Name: "api", Namespace: "default", DesiredReplicas: 3, AvailableReplicas: 1, UnavailableReplicas: 2, ApplicationID: "app-456"},
			},
		},
	}

	results := Evaluate(input)

	labeled := findCondition(t, results, "default/web-1", ConditionCrashLoopBackOff)
	if labeled == nil || labeled.ApplicationID != "app-123" {
		t.Fatalf("expected web-1's condition to carry ApplicationID app-123, got %+v", labeled)
	}

	unlabeled := findCondition(t, results, "default/web-2", ConditionCrashLoopBackOff)
	if unlabeled == nil || unlabeled.ApplicationID != "" {
		t.Fatalf("expected web-2's condition to have no ApplicationID, got %+v", unlabeled)
	}

	deployment := findCondition(t, results, "default/api", ConditionDeploymentFailure)
	if deployment == nil || deployment.ApplicationID != "app-456" {
		t.Fatalf("expected the deployment condition to carry ApplicationID app-456, got %+v", deployment)
	}
}

func TestEvaluate_DeploymentFailure(t *testing.T) {
	input := EvaluateInput{
		ClusterID: "cluster-1", ClusterName: "prod",
		Snapshot: metrics.CollectedMetrics{
			Deployments: []metrics.DeploymentMetrics{
				{Name: "api", Namespace: "default", DesiredReplicas: 3, AvailableReplicas: 1, UnavailableReplicas: 2},
				{Name: "worker", Namespace: "default", DesiredReplicas: 2, AvailableReplicas: 2, UnavailableReplicas: 0},
			},
		},
	}

	results := Evaluate(input)

	failing := findCondition(t, results, "default/api", ConditionDeploymentFailure)
	if failing == nil {
		t.Fatalf("expected default/api to fire a deployment failure condition, got %+v", results)
	}

	if findCondition(t, results, "default/worker", ConditionDeploymentFailure) != nil {
		t.Fatalf("expected default/worker (fully available) not to fire")
	}
}
