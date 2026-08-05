package dto

import "time"

type ConfigMapBinaryDataMetadataResponse struct {
	Key       string `json:"key"`
	SizeBytes int    `json:"sizeBytes"`
}

type ConfigMapResponse struct {
	Name              string                                `json:"name"`
	Namespace         string                                `json:"namespace"`
	DataKeyCount      int                                   `json:"dataKeyCount"`
	Labels            map[string]string                     `json:"labels"`
	Annotations       map[string]string                     `json:"annotations"`
	Age               string                                `json:"age"`
	CreationTimestamp time.Time                             `json:"creationTimestamp"`
	Data              map[string]string                     `json:"data,omitempty"`
	BinaryData        []ConfigMapBinaryDataMetadataResponse `json:"binaryData,omitempty"`
	Immutable         bool                                  `json:"immutable,omitempty"`
	UID               string                                `json:"uid,omitempty"`
	ResourceVersion   string                                `json:"resourceVersion,omitempty"`
}

type ConfigMapListResponse struct {
	Items []ConfigMapResponse `json:"items"`
	Total int                 `json:"total"`
}
