package dto

import "time"

type ContainerStatusResponse struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restartCount"`
	State        string `json:"state"`
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
	Name              string                    `json:"name"`
	Namespace         string                    `json:"namespace"`
	Phase             string                    `json:"phase"`
	Ready             bool                      `json:"ready"`
	RestartCount      int32                     `json:"restartCount"`
	NodeName          string                    `json:"nodeName"`
	PodIP             string                    `json:"podIP"`
	HostIP            string                    `json:"hostIP"`
	CreationTimestamp time.Time                 `json:"creationTimestamp"`
	Age               string                    `json:"age"`
	ContainerImages   []string                  `json:"containerImages"`
	ContainerState    string                    `json:"containerState"`
	ContainerCount    int                       `json:"containerCount"`
	Labels            map[string]string         `json:"labels"`
	OwnerReferences   []OwnerReferenceResponse  `json:"ownerReferences"`
	Conditions        []PodConditionResponse    `json:"conditions,omitempty"`
	Events            []string                  `json:"events,omitempty"`
	ContainerStatuses []ContainerStatusResponse `json:"containerStatuses,omitempty"`
	StartTime         *time.Time                `json:"startTime,omitempty"`
	QoSClass          string                    `json:"qosClass,omitempty"`
	Volumes           []string                  `json:"volumes,omitempty"`
	ServiceAccount    string                    `json:"serviceAccount,omitempty"`
}

type PodListResponse struct {
	Items []PodResponse `json:"items"`
	Total int           `json:"total"`
}
