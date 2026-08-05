package dto

import "time"

type SecretResponse struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Type              string            `json:"type"`
	DataKeyCount      int               `json:"dataKeyCount"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	Age               string            `json:"age"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
}

type SecretDetailResponse struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Type            string            `json:"type"`
	Keys            []string          `json:"keys"`
	Immutable       bool              `json:"immutable"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	UID             string            `json:"uid,omitempty"`
	ResourceVersion string            `json:"resourceVersion,omitempty"`
	Age             string            `json:"age"`
}

type SecretListResponse struct {
	Items []SecretResponse `json:"items"`
	Total int              `json:"total"`
}
