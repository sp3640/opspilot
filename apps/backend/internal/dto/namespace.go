package dto

import "time"

// NamespaceResponse is a cluster-level Kubernetes Namespace — never scoped
// to a project or application, since Namespaces are a property of the
// cluster itself.
type NamespaceResponse struct {
	Name              string            `json:"name"`
	Phase             string            `json:"phase"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	Age               string            `json:"age"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
}

type NamespaceListResponse struct {
	Items []NamespaceResponse `json:"items"`
	Total int                 `json:"total"`
}
