package deployments

import (
	"fmt"
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	DeploymentStatusHealthy     = "Healthy"
	DeploymentStatusProgressing = "Progressing"
	DeploymentStatusScaling     = "Scaling"
	DeploymentStatusUnavailable = "Unavailable"
	DeploymentStatusFailed      = "Failed"
	DeploymentStatusUnknown     = "Unknown"
)

type DeploymentRuntimeMapper struct {
	now func() time.Time
}

func NewDeploymentRuntimeMapper() *DeploymentRuntimeMapper {
	return &DeploymentRuntimeMapper{now: time.Now}
}

func (m *DeploymentRuntimeMapper) CalculateDeploymentAge(createdAt time.Time) string {
	if createdAt.IsZero() {
		return ""
	}

	delta := m.now().Sub(createdAt)
	if delta < time.Minute {
		return fmt.Sprintf("%ds", int(delta.Seconds()))
	}
	if delta < time.Hour {
		return fmt.Sprintf("%dm", int(delta.Minutes()))
	}
	if delta < 24*time.Hour {
		return fmt.Sprintf("%dh", int(delta.Hours()))
	}

	return fmt.Sprintf("%dd", int(delta.Hours()/24))
}

func (m *DeploymentRuntimeMapper) DetermineDeploymentStatus(deployment appsv1.Deployment) string {
	desired := int32(1)
	if deployment.Spec.Replicas != nil {
		desired = *deployment.Spec.Replicas
	}

	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing && condition.Status == "False" && condition.Reason == "ProgressDeadlineExceeded" {
			return DeploymentStatusFailed
		}
	}

	if deployment.Status.Replicas != desired {
		return DeploymentStatusScaling
	}

	if deployment.Status.UnavailableReplicas > 0 && deployment.Status.AvailableReplicas == 0 && desired > 0 {
		return DeploymentStatusUnavailable
	}

	if deployment.Status.UpdatedReplicas < desired || deployment.Status.ReadyReplicas < desired || deployment.Status.AvailableReplicas < desired {
		return DeploymentStatusProgressing
	}

	if deployment.Status.ReadyReplicas >= desired && deployment.Status.UpdatedReplicas >= desired && deployment.Status.AvailableReplicas >= desired {
		return DeploymentStatusHealthy
	}

	return DeploymentStatusUnknown
}

func (m *DeploymentRuntimeMapper) MapDeployment(deployment appsv1.Deployment, includeDetails bool) dto.DeploymentRuntimeResponse {
	status := m.DetermineDeploymentStatus(deployment)
	strategyType := string(deployment.Spec.Strategy.Type)
	if strategyType == "" {
		strategyType = string(appsv1.RollingUpdateDeploymentStrategyType)
	}

	response := dto.DeploymentRuntimeResponse{
		Name:                deployment.Name,
		Namespace:           deployment.Namespace,
		Replicas:            deployment.Status.Replicas,
		ReadyReplicas:       deployment.Status.ReadyReplicas,
		UpdatedReplicas:     deployment.Status.UpdatedReplicas,
		AvailableReplicas:   deployment.Status.AvailableReplicas,
		UnavailableReplicas: deployment.Status.UnavailableReplicas,
		Strategy:            strategyType,
		Labels:              deployment.Labels,
		Age:                 m.CalculateDeploymentAge(deployment.CreationTimestamp.Time),
		Status:              status,
	}

	return response
}

func (m *DeploymentRuntimeMapper) MapDeploymentDetail(deployment appsv1.Deployment) dto.DeploymentRuntimeDetailResponse {
	selector := map[string]any{}
	if deployment.Spec.Selector != nil {
		selector["matchLabels"] = deployment.Spec.Selector.MatchLabels
		selector["matchExpressions"] = deployment.Spec.Selector.MatchExpressions
	}

	podTemplate := map[string]any{
		"metadata": map[string]any{
			"labels":      deployment.Spec.Template.Labels,
			"annotations": deployment.Spec.Template.Annotations,
		},
		"spec": deployment.Spec.Template.Spec,
	}

	strategy := map[string]any{
		"type": deployment.Spec.Strategy.Type,
	}
	if deployment.Spec.Strategy.RollingUpdate != nil {
		strategy["rollingUpdate"] = deployment.Spec.Strategy.RollingUpdate
	}

	conditions := make([]dto.DeploymentRuntimeConditionResponse, 0, len(deployment.Status.Conditions))
	for _, cond := range deployment.Status.Conditions {
		conditions = append(conditions, dto.DeploymentRuntimeConditionResponse{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			Reason:             cond.Reason,
			Message:            cond.Message,
			LastUpdateTime:     ptrTime(cond.LastUpdateTime.Time),
			LastTransitionTime: ptrTime(cond.LastTransitionTime.Time),
		})
	}

	revision := ""
	if deployment.Annotations != nil {
		revision = deployment.Annotations["deployment.kubernetes.io/revision"]
	}

	replicas := int32(1)
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}

	return dto.DeploymentRuntimeDetailResponse{
		Metadata: map[string]any{
			"name":              deployment.Name,
			"namespace":         deployment.Namespace,
			"uid":               string(deployment.UID),
			"resourceVersion":   deployment.ResourceVersion,
			"generation":        deployment.Generation,
			"creationTimestamp": deployment.CreationTimestamp.Time,
			"labels":            deployment.Labels,
			"annotations":       deployment.Annotations,
		},
		Spec: map[string]any{
			"replicas":                replicas,
			"minReadySeconds":         deployment.Spec.MinReadySeconds,
			"paused":                  deployment.Spec.Paused,
			"progressDeadlineSeconds": deployment.Spec.ProgressDeadlineSeconds,
			"revisionHistoryLimit":    deployment.Spec.RevisionHistoryLimit,
			"strategy":                deployment.Spec.Strategy,
			"selector":                deployment.Spec.Selector,
			"template":                deployment.Spec.Template,
		},
		Selector:            selector,
		PodTemplate:         podTemplate,
		Strategy:            strategy,
		Conditions:          conditions,
		Replicas:            replicas,
		ReadyReplicas:       deployment.Status.ReadyReplicas,
		UpdatedReplicas:     deployment.Status.UpdatedReplicas,
		AvailableReplicas:   deployment.Status.AvailableReplicas,
		UnavailableReplicas: deployment.Status.UnavailableReplicas,
		ObservedGeneration:  deployment.Status.ObservedGeneration,
		Revision:            revision,
		Labels:              deployment.Labels,
		Annotations:         deployment.Annotations,
		Age:                 m.CalculateDeploymentAge(deployment.CreationTimestamp.Time),
		Status:              m.DetermineDeploymentStatus(deployment),
	}
}

func ptrTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}

	copied := value
	return &copied
}
