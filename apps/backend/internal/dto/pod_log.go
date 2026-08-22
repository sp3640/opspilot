package dto

import "time"

type PodLogResponse struct {
	Pod         string    `json:"pod"`
	Container   string    `json:"container"`
	Namespace   string    `json:"namespace"`
	Log         string    `json:"log"`
	Previous    bool      `json:"previous"`
	RetrievedAt time.Time `json:"retrievedAt"`
}
