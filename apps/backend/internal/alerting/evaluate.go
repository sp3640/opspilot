package alerting

import (
	"fmt"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/metrics"
)

// EvaluateInput is one cluster's freshly-collected snapshot, ready to be
// checked against the conditions this engine supports.
type EvaluateInput struct {
	ClusterID   string
	ClusterName string
	Snapshot    metrics.CollectedMetrics
}

// Evaluate is a pure function: given a snapshot, it returns every condition
// currently firing. It performs no I/O and makes no decisions about
// creating/updating/resolving alerts - that correlation logic lives in
// AlertService.ReconcileConditions, which this package's caller drives.
func Evaluate(input EvaluateInput) []EvaluatedCondition {
	var results []EvaluatedCondition

	results = append(results, evaluateClusterResourceThresholds(input)...)
	results = append(results, evaluateNodes(input)...)
	results = append(results, evaluatePods(input)...)
	results = append(results, evaluateDeployments(input)...)

	return results
}

func evaluateClusterResourceThresholds(input EvaluateInput) []EvaluatedCondition {
	var results []EvaluatedCondition
	cluster := input.Snapshot.Cluster

	if cluster.CPUCapacityMilliCores > 0 {
		percent := percentOf(cluster.CPUUsageMilliCores, cluster.CPUCapacityMilliCores)
		if severity, threshold, ok := thresholdSeverity(percent, CPUWarningPercent, CPUCriticalPercent); ok {
			results = append(results, EvaluatedCondition{
				ResourceType: constants.AlertResourceTypeCluster,
				ResourceID:   input.ClusterID,
				ResourceName: input.ClusterName,
				ConditionKey: ConditionCPUThreshold,
				Title:        fmt.Sprintf("High CPU usage on cluster %s", input.ClusterName),
				Description:  fmt.Sprintf("Cluster %s CPU usage is %.1f%% (threshold %.0f%%)", input.ClusterName, percent, threshold),
				Severity:     severity,
				CurrentValue: percent,
				Threshold:    threshold,
				Unit:         "%",
				ClusterID:    input.ClusterID,
				ClusterName:  input.ClusterName,
			})
		}
	}

	if cluster.MemoryCapacityBytes > 0 {
		percent := percentOf(cluster.MemoryUsageBytes, cluster.MemoryCapacityBytes)
		if severity, threshold, ok := thresholdSeverity(percent, MemoryWarningPercent, MemoryCriticalPercent); ok {
			results = append(results, EvaluatedCondition{
				ResourceType: constants.AlertResourceTypeCluster,
				ResourceID:   input.ClusterID,
				ResourceName: input.ClusterName,
				ConditionKey: ConditionMemoryThreshold,
				Title:        fmt.Sprintf("High memory usage on cluster %s", input.ClusterName),
				Description:  fmt.Sprintf("Cluster %s memory usage is %.1f%% (threshold %.0f%%)", input.ClusterName, percent, threshold),
				Severity:     severity,
				CurrentValue: percent,
				Threshold:    threshold,
				Unit:         "%",
				ClusterID:    input.ClusterID,
				ClusterName:  input.ClusterName,
			})
		}
	}

	return results
}

func evaluateNodes(input EvaluateInput) []EvaluatedCondition {
	var results []EvaluatedCondition

	for _, node := range input.Snapshot.Nodes {
		if node.Ready {
			continue
		}

		results = append(results, EvaluatedCondition{
			ResourceType: constants.AlertResourceTypeNode,
			ResourceID:   node.Name,
			ResourceName: node.Name,
			ConditionKey: ConditionNodeNotReady,
			Title:        fmt.Sprintf("Node %s is not ready", node.Name),
			Description:  fmt.Sprintf("Node %s is reporting NotReady", node.Name),
			Severity:     constants.AlertSeverityHigh,
			CurrentValue: 0,
			Threshold:    1,
			Unit:         "ready",
			ClusterID:    input.ClusterID,
			ClusterName:  input.ClusterName,
		})
	}

	return results
}

func evaluatePods(input EvaluateInput) []EvaluatedCondition {
	var results []EvaluatedCondition

	for _, pod := range input.Snapshot.Pods {
		resourceID := pod.Namespace + "/" + pod.Name

		if severity, threshold, ok := thresholdSeverity(float64(pod.Restarts), PodRestartWarningCount, PodRestartCriticalCount); ok {
			results = append(results, EvaluatedCondition{
				ResourceType:  constants.AlertResourceTypePod,
				ResourceID:    resourceID,
				ResourceName:  resourceID,
				ConditionKey:  ConditionPodRestartCount,
				Title:         fmt.Sprintf("Pod %s is restarting frequently", resourceID),
				Description:   fmt.Sprintf("Pod %s has restarted %d times (threshold %.0f)", resourceID, pod.Restarts, threshold),
				Severity:      severity,
				CurrentValue:  float64(pod.Restarts),
				Threshold:     threshold,
				Unit:          "restarts",
				ClusterID:     input.ClusterID,
				ClusterName:   input.ClusterName,
				ApplicationID: pod.ApplicationID,
			})
		}

		if pod.ContainerState == "CrashLoopBackOff" {
			results = append(results, EvaluatedCondition{
				ResourceType:  constants.AlertResourceTypePod,
				ResourceID:    resourceID,
				ResourceName:  resourceID,
				ConditionKey:  ConditionCrashLoopBackOff,
				Title:         fmt.Sprintf("Pod %s is in CrashLoopBackOff", resourceID),
				Description:   fmt.Sprintf("Pod %s is stuck in CrashLoopBackOff (%d restarts so far)", resourceID, pod.Restarts),
				Severity:      constants.AlertSeverityCritical,
				CurrentValue:  float64(pod.Restarts),
				Threshold:     0,
				Unit:          "restarts",
				ClusterID:     input.ClusterID,
				ClusterName:   input.ClusterName,
				ApplicationID: pod.ApplicationID,
			})
		}
	}

	return results
}

func evaluateDeployments(input EvaluateInput) []EvaluatedCondition {
	var results []EvaluatedCondition

	for _, deployment := range input.Snapshot.Deployments {
		if deployment.UnavailableReplicas <= 0 && deployment.AvailableReplicas >= deployment.DesiredReplicas {
			continue
		}

		resourceID := deployment.Namespace + "/" + deployment.Name
		results = append(results, EvaluatedCondition{
			ResourceType: constants.AlertResourceTypeDeployment,
			ResourceID:   resourceID,
			ResourceName: resourceID,
			ConditionKey: ConditionDeploymentFailure,
			Title:        fmt.Sprintf("Deployment %s has unavailable replicas", resourceID),
			Description: fmt.Sprintf(
				"Deployment %s has %d/%d replicas available (%d unavailable)",
				resourceID, deployment.AvailableReplicas, deployment.DesiredReplicas, deployment.UnavailableReplicas,
			),
			Severity:      constants.AlertSeverityHigh,
			CurrentValue:  float64(deployment.AvailableReplicas),
			Threshold:     float64(deployment.DesiredReplicas),
			Unit:          "replicas",
			ClusterID:     input.ClusterID,
			ClusterName:   input.ClusterName,
			ApplicationID: deployment.ApplicationID,
		})
	}

	return results
}

func percentOf(usage, capacity int64) float64 {
	if capacity <= 0 {
		return 0
	}

	return (float64(usage) / float64(capacity)) * 100
}

// thresholdSeverity returns (severity, threshold-crossed, true) when value
// has crossed the critical or warning threshold, else (_, _, false).
func thresholdSeverity(value, warn, crit float64) (string, float64, bool) {
	if value >= crit {
		return constants.AlertSeverityCritical, crit, true
	}
	if value >= warn {
		return constants.AlertSeverityHigh, warn, true
	}

	return "", 0, false
}
