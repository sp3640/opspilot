package dto

import "time"

type DeploymentRuntimeResponse struct {
	Name                string            `json:"name"`
	Namespace           string            `json:"namespace"`
	Replicas            int32             `json:"replicas"`
	ReadyReplicas       int32             `json:"readyReplicas"`
	UpdatedReplicas     int32             `json:"updatedReplicas"`
	AvailableReplicas   int32             `json:"availableReplicas"`
	UnavailableReplicas int32             `json:"unavailableReplicas"`
	Strategy            string            `json:"strategy"`
	Labels              map[string]string `json:"labels"`
	Age                 string            `json:"age"`
	Status              string            `json:"status"`
}

type DeploymentRuntimeConditionResponse struct {
	Type               string     `json:"type"`
	Status             string     `json:"status"`
	Reason             string     `json:"reason,omitempty"`
	Message            string     `json:"message,omitempty"`
	LastUpdateTime     *time.Time `json:"lastUpdateTime,omitempty"`
	LastTransitionTime *time.Time `json:"lastTransitionTime,omitempty"`
}

type DeploymentRuntimeDetailResponse struct {
	Metadata            map[string]any                       `json:"metadata"`
	Spec                map[string]any                       `json:"spec"`
	Selector            map[string]any                       `json:"selector"`
	PodTemplate         map[string]any                       `json:"podTemplate"`
	Strategy            map[string]any                       `json:"strategy"`
	Conditions          []DeploymentRuntimeConditionResponse `json:"conditions"`
	Replicas            int32                                `json:"replicas"`
	ReadyReplicas       int32                                `json:"readyReplicas"`
	UpdatedReplicas     int32                                `json:"updatedReplicas"`
	AvailableReplicas   int32                                `json:"availableReplicas"`
	UnavailableReplicas int32                                `json:"unavailableReplicas"`
	ObservedGeneration  int64                                `json:"observedGeneration"`
	Revision            string                               `json:"revision,omitempty"`
	Labels              map[string]string                    `json:"labels"`
	Annotations         map[string]string                    `json:"annotations"`
	Age                 string                               `json:"age"`
	Status              string                               `json:"status"`
}

type DeploymentRuntimeListResponse struct {
	Items []DeploymentRuntimeResponse `json:"items"`
	Total int                         `json:"total"`
}
