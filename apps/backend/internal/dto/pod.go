package dto

import "time"

type ContainerStatusResponse struct {
	Name                      string     `json:"name"`
	Image                     string     `json:"image"`
	Ready                     bool       `json:"ready"`
	Started                   bool       `json:"started"`
	RestartCount              int32      `json:"restartCount"`
	State                     string     `json:"state"`
	StateReason               string     `json:"stateReason,omitempty"`
	StateMessage              string     `json:"stateMessage,omitempty"`
	ExitCode                  *int32     `json:"exitCode,omitempty"`
	LastTerminationReason     string     `json:"lastTerminationReason,omitempty"`
	LastTerminationExitCode   *int32     `json:"lastTerminationExitCode,omitempty"`
	LastTerminationFinishedAt *time.Time `json:"lastTerminationFinishedAt,omitempty"`
	HasReadinessProbe         bool       `json:"hasReadinessProbe"`
	HasLivenessProbe          bool       `json:"hasLivenessProbe"`
	CPURequest                string     `json:"cpuRequest,omitempty"`
	CPULimit                  string     `json:"cpuLimit,omitempty"`
	MemoryRequest             string     `json:"memoryRequest,omitempty"`
	MemoryLimit               string     `json:"memoryLimit,omitempty"`
}

// ContainerMetricsResponse holds live resource usage for a single container,
// sourced from the metrics.k8s.io API (metrics-server). Nil/omitted when
// metrics-server is not installed on the target cluster.
type ContainerMetricsResponse struct {
	Name   string `json:"name"`
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// PodMetricsResponse is nil unless the target cluster has metrics-server
// installed. Callers must render "Not available" rather than fabricate a
// value when this is absent.
type PodMetricsResponse struct {
	Timestamp  time.Time                  `json:"timestamp"`
	Window     string                     `json:"window"`
	Containers []ContainerMetricsResponse `json:"containers"`
}

type PodConditionResponse struct {
	Type               string     `json:"type"`
	Status             string     `json:"status"`
	Reason             string     `json:"reason,omitempty"`
	Message            string     `json:"message,omitempty"`
	LastTransitionTime *time.Time `json:"lastTransitionTime,omitempty"`
}

type OwnerReferenceResponse struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	UID        string `json:"uid"`
}

type PodResponse struct {
	Name                string                    `json:"name"`
	Namespace           string                    `json:"namespace"`
	Phase               string                    `json:"phase"`
	Reason              string                    `json:"reason,omitempty"`
	Message             string                    `json:"message,omitempty"`
	Ready               bool                      `json:"ready"`
	ReadyContainerCount int                       `json:"readyContainerCount"`
	RestartCount        int32                     `json:"restartCount"`
	NodeName            string                    `json:"nodeName"`
	PodIP               string                    `json:"podIP"`
	HostIP              string                    `json:"hostIP"`
	CreationTimestamp   time.Time                 `json:"creationTimestamp"`
	Age                 string                    `json:"age"`
	ContainerImages     []string                  `json:"containerImages"`
	ContainerState      string                    `json:"containerState"`
	ContainerCount      int                       `json:"containerCount"`
	Labels              map[string]string         `json:"labels"`
	OwnerReferences     []OwnerReferenceResponse  `json:"ownerReferences"`
	Conditions          []PodConditionResponse    `json:"conditions,omitempty"`
	Events              []string                  `json:"events,omitempty"`
	ContainerStatuses   []ContainerStatusResponse `json:"containerStatuses,omitempty"`
	StartTime           *time.Time                `json:"startTime,omitempty"`
	QoSClass            string                    `json:"qosClass,omitempty"`
	Volumes             []string                  `json:"volumes,omitempty"`
	ServiceAccount      string                    `json:"serviceAccount,omitempty"`
	Metrics             *PodMetricsResponse       `json:"metrics,omitempty"`
}

type PodListResponse struct {
	Items []PodResponse `json:"items"`
	Total int           `json:"total"`
}
