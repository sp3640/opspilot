package dto

import "time"

type IngressPathResponse struct {
	Path           string `json:"path"`
	PathType       string `json:"pathType,omitempty"`
	BackendService string `json:"backendService,omitempty"`
	ServicePort    string `json:"servicePort,omitempty"`
}

type IngressRuleResponse struct {
	Host  string                `json:"host"`
	Paths []IngressPathResponse `json:"paths"`
}

type IngressTLSResponse struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName,omitempty"`
}

type IngressResponse struct {
	Name                 string                   `json:"name"`
	Namespace            string                   `json:"namespace"`
	IngressClass         string                   `json:"ingressClass,omitempty"`
	Hosts                []string                 `json:"hosts"`
	Paths                []string                 `json:"paths"`
	BackendServices      []string                 `json:"backendServices"`
	TLSEnabled           bool                     `json:"tlsEnabled"`
	TLSSecret            string                   `json:"tlsSecret,omitempty"`
	LoadBalancerIP       string                   `json:"loadBalancerIP,omitempty"`
	LoadBalancerHostname string                   `json:"loadBalancerHostname,omitempty"`
	Labels               map[string]string        `json:"labels"`
	Annotations          map[string]string        `json:"annotations"`
	Age                  string                   `json:"age"`
	CreationTimestamp    time.Time                `json:"creationTimestamp"`
	Rules                []IngressRuleResponse    `json:"rules,omitempty"`
	HTTPPaths            []IngressPathResponse    `json:"httpPaths,omitempty"`
	DefaultBackend       string                   `json:"defaultBackend,omitempty"`
	DefaultBackendPort   string                   `json:"defaultBackendPort,omitempty"`
	TLSConfig            []IngressTLSResponse     `json:"tlsConfig,omitempty"`
	Status               string                   `json:"status,omitempty"`
	OwnerReferences      []OwnerReferenceResponse `json:"ownerReferences,omitempty"`
	UID                  string                   `json:"uid,omitempty"`
	ResourceVersion      string                   `json:"resourceVersion,omitempty"`
}

type IngressListResponse struct {
	Items []IngressResponse `json:"items"`
	Total int               `json:"total"`
}
