package dto

import "time"

type ServicePortResponse struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort"`
	Protocol   string `json:"protocol"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

type ServiceResponse struct {
	Name                     string                   `json:"name"`
	Namespace                string                   `json:"namespace"`
	Type                     string                   `json:"type"`
	ClusterIP                string                   `json:"clusterIP"`
	ExternalIPs              []string                 `json:"externalIPs"`
	LoadBalancerIP           string                   `json:"loadBalancerIP,omitempty"`
	Ports                    []ServicePortResponse    `json:"ports"`
	TargetPorts              []string                 `json:"targetPorts"`
	Protocol                 string                   `json:"protocol"`
	Selector                 map[string]string        `json:"selector"`
	Labels                   map[string]string        `json:"labels"`
	Annotations              map[string]string        `json:"annotations"`
	Age                      string                   `json:"age"`
	CreationTimestamp        time.Time                `json:"creationTimestamp"`
	SessionAffinity          string                   `json:"sessionAffinity,omitempty"`
	InternalTrafficPolicy    string                   `json:"internalTrafficPolicy,omitempty"`
	ExternalTrafficPolicy    string                   `json:"externalTrafficPolicy,omitempty"`
	HealthCheckNodePort      int32                    `json:"healthCheckNodePort,omitempty"`
	PublishNotReadyAddresses bool                     `json:"publishNotReadyAddresses,omitempty"`
	IPFamilies               []string                 `json:"ipFamilies,omitempty"`
	IPFamilyPolicy           string                   `json:"ipFamilyPolicy,omitempty"`
	OwnerReferences          []OwnerReferenceResponse `json:"ownerReferences,omitempty"`
	ResourceVersion          string                   `json:"resourceVersion,omitempty"`
	UID                      string                   `json:"uid,omitempty"`
}

type ServiceListResponse struct {
	Items []ServiceResponse `json:"items"`
	Total int               `json:"total"`
}
