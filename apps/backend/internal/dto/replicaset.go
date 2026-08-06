package dto

import "time"

type ReplicaSetSummary struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	DesiredReplicas   int32             `json:"desiredReplicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	OwnerDeployment   string            `json:"ownerDeployment"`
	Revision          string            `json:"revision"`
	Status            string            `json:"status"`
	Labels            map[string]string `json:"labels"`
	Age               string            `json:"age"`
}

type ReplicaSetDetail struct {
	ID                   string             `json:"id"`
	Name                 string             `json:"name"`
	Namespace            string             `json:"namespace"`
	DesiredReplicas      int32              `json:"desiredReplicas"`
	ReadyReplicas        int32              `json:"readyReplicas"`
	AvailableReplicas    int32              `json:"availableReplicas"`
	FullyLabeledReplicas int32              `json:"fullyLabeledReplicas"`
	ObservedGeneration   int64              `json:"observedGeneration"`
	Revision             string             `json:"revision"`
	Status               string             `json:"status"`
	Labels               map[string]string  `json:"labels"`
	Annotations          map[string]string  `json:"annotations"`
	Selector             map[string]string  `json:"selector"`
	OwnerDeployment      string             `json:"ownerDeployment"`
	Conditions           []ReplicaCondition `json:"conditions"`
	CreatedAt            time.Time          `json:"createdAt"`
	Age                  string             `json:"age"`
}

type ReplicaCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
	LastTransitionTime string `json:"lastTransitionTime"`
}
