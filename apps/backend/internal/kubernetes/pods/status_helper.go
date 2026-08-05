package pods

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type PodStatusHelper struct{}

func NewPodStatusHelper() *PodStatusHelper {
	return &PodStatusHelper{}
}

func (h *PodStatusHelper) IsPodReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

func (h *PodStatusHelper) TotalRestarts(pod corev1.Pod) int32 {
	var count int32
	for _, container := range pod.Status.ContainerStatuses {
		count += container.RestartCount
	}

	return count
}

func (h *PodStatusHelper) ResolveContainerState(status corev1.ContainerStatus) string {
	if status.State.Running != nil {
		return "Running"
	}
	if status.State.Waiting != nil {
		if status.State.Waiting.Reason != "" {
			return status.State.Waiting.Reason
		}
		return "Waiting"
	}
	if status.State.Terminated != nil {
		if status.State.Terminated.Reason != "" {
			return status.State.Terminated.Reason
		}
		return "Terminated"
	}

	return "Unknown"
}

func (h *PodStatusHelper) ResolveAggregateContainerState(pod corev1.Pod) string {
	if len(pod.Status.ContainerStatuses) == 0 {
		return "Unknown"
	}

	state := "Running"
	for _, container := range pod.Status.ContainerStatuses {
		resolved := h.ResolveContainerState(container)
		if resolved != "Running" {
			state = resolved
			break
		}
	}

	return state
}

func (h *PodStatusHelper) Age(now time.Time, createdAt time.Time) string {
	if createdAt.IsZero() {
		return ""
	}

	delta := now.Sub(createdAt)
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
