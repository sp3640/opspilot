package dto

import "time"

// NodeResponse is a cluster-level Kubernetes Node — never scoped to a
// project or application, since Nodes are a property of the cluster itself.
type NodeResponse struct {
	Name              string            `json:"name"`
	Ready             bool              `json:"ready"`
	Unschedulable     bool              `json:"unschedulable"`
	KubeletVersion    string            `json:"kubeletVersion"`
	OperatingSystem   string            `json:"operatingSystem"`
	Architecture      string            `json:"architecture"`
	InternalIP        string            `json:"internalIP,omitempty"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	Age               string            `json:"age"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
}

type NodeListResponse struct {
	Items []NodeResponse `json:"items"`
	Total int            `json:"total"`
}
