package pods

import (
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
	corev1 "k8s.io/api/core/v1"
)

type PodMapper struct {
	helper *PodStatusHelper
	now    func() time.Time
}

func NewPodMapper(helper *PodStatusHelper) *PodMapper {
	if helper == nil {
		helper = NewPodStatusHelper()
	}

	return &PodMapper{
		helper: helper,
		now:    time.Now,
	}
}

func (m *PodMapper) MapPod(pod corev1.Pod) dto.PodResponse {
	containerSpecs := make(map[string]corev1.Container, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		containerSpecs[container.Name] = container
	}

	readyContainerCount := 0
	containerStatuses := make([]dto.ContainerStatusResponse, 0, len(pod.Status.ContainerStatuses))
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready {
			readyContainerCount++
		}

		mapped := dto.ContainerStatusResponse{
			Name:         containerStatus.Name,
			Image:        containerStatus.Image,
			Ready:        containerStatus.Ready,
			Started:      containerStatus.Started != nil && *containerStatus.Started,
			RestartCount: containerStatus.RestartCount,
			State:        m.helper.ResolveContainerState(containerStatus),
		}

		switch {
		case containerStatus.State.Waiting != nil:
			mapped.StateReason = containerStatus.State.Waiting.Reason
			mapped.StateMessage = containerStatus.State.Waiting.Message
		case containerStatus.State.Terminated != nil:
			mapped.StateReason = containerStatus.State.Terminated.Reason
			mapped.StateMessage = containerStatus.State.Terminated.Message
			exitCode := containerStatus.State.Terminated.ExitCode
			mapped.ExitCode = &exitCode
		}

		if containerStatus.LastTerminationState.Terminated != nil {
			terminated := containerStatus.LastTerminationState.Terminated
			mapped.LastTerminationReason = terminated.Reason
			exitCode := terminated.ExitCode
			mapped.LastTerminationExitCode = &exitCode
			if !terminated.FinishedAt.IsZero() {
				finishedAt := terminated.FinishedAt.Time
				mapped.LastTerminationFinishedAt = &finishedAt
			}
		}

		if spec, ok := containerSpecs[containerStatus.Name]; ok {
			mapped.HasReadinessProbe = spec.ReadinessProbe != nil
			mapped.HasLivenessProbe = spec.LivenessProbe != nil
			mapped.CPURequest = quantityString(spec.Resources.Requests, corev1.ResourceCPU)
			mapped.CPULimit = quantityString(spec.Resources.Limits, corev1.ResourceCPU)
			mapped.MemoryRequest = quantityString(spec.Resources.Requests, corev1.ResourceMemory)
			mapped.MemoryLimit = quantityString(spec.Resources.Limits, corev1.ResourceMemory)
		}

		containerStatuses = append(containerStatuses, mapped)
	}

	conditions := make([]dto.PodConditionResponse, 0, len(pod.Status.Conditions))
	for _, condition := range pod.Status.Conditions {
		transition := condition.LastTransitionTime.Time
		if transition.IsZero() {
			conditions = append(conditions, dto.PodConditionResponse{
				Type:    string(condition.Type),
				Status:  string(condition.Status),
				Reason:  condition.Reason,
				Message: condition.Message,
			})
			continue
		}

		conditions = append(conditions, dto.PodConditionResponse{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: &transition,
		})
	}

	ownerReferences := make([]dto.OwnerReferenceResponse, 0, len(pod.OwnerReferences))
	for _, ownerReference := range pod.OwnerReferences {
		ownerReferences = append(ownerReferences, dto.OwnerReferenceResponse{
			APIVersion: ownerReference.APIVersion,
			Kind:       ownerReference.Kind,
			Name:       ownerReference.Name,
			UID:        string(ownerReference.UID),
		})
	}

	containerImages := make([]string, 0, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		containerImages = append(containerImages, container.Image)
	}

	volumes := make([]string, 0, len(pod.Spec.Volumes))
	for _, volume := range pod.Spec.Volumes {
		volumes = append(volumes, volume.Name)
	}

	var startTime *time.Time
	if pod.Status.StartTime != nil {
		startedAt := pod.Status.StartTime.Time
		startTime = &startedAt
	}

	return dto.PodResponse{
		Name:                pod.Name,
		Namespace:           pod.Namespace,
		Phase:               string(pod.Status.Phase),
		Reason:              pod.Status.Reason,
		Message:             pod.Status.Message,
		Ready:               m.helper.IsPodReady(pod),
		ReadyContainerCount: readyContainerCount,
		RestartCount:        m.helper.TotalRestarts(pod),
		NodeName:            pod.Spec.NodeName,
		PodIP:               pod.Status.PodIP,
		HostIP:              pod.Status.HostIP,
		CreationTimestamp:   pod.CreationTimestamp.Time,
		Age:                 m.helper.Age(m.now(), pod.CreationTimestamp.Time),
		ContainerImages:     containerImages,
		ContainerState:      m.helper.ResolveAggregateContainerState(pod),
		ContainerCount:      len(pod.Spec.Containers),
		Labels:              pod.Labels,
		OwnerReferences:     ownerReferences,
		Conditions:          conditions,
		ContainerStatuses:   containerStatuses,
		StartTime:           startTime,
		QoSClass:            string(pod.Status.QOSClass),
		Volumes:             volumes,
		ServiceAccount:      pod.Spec.ServiceAccountName,
	}
}

func quantityString(list corev1.ResourceList, name corev1.ResourceName) string {
	quantity, ok := list[name]
	if !ok {
		return ""
	}

	return quantity.String()
}

func (m *PodMapper) MapPodList(items []corev1.Pod) []dto.PodResponse {
	responses := make([]dto.PodResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, m.MapPod(item))
	}

	return responses
}
