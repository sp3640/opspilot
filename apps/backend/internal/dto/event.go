package dto

import "time"

type EventSeriesResponse struct {
	Count            int32      `json:"count"`
	LastObservedTime *time.Time `json:"lastObservedTime,omitempty"`
}

type EventResponse struct {
	Name           string     `json:"name"`
	Namespace      string     `json:"namespace"`
	Reason         string     `json:"reason"`
	Message        string     `json:"message"`
	Type           string     `json:"type"`
	Count          int32      `json:"count"`
	FirstTimestamp *time.Time `json:"firstTimestamp,omitempty"`
	LastTimestamp  *time.Time `json:"lastTimestamp,omitempty"`
	Age            string     `json:"age"`
	InvolvedObject string     `json:"involvedObject"`
	Component      string     `json:"component"`
	Source         string     `json:"source"`
}

type EventDetailResponse struct {
	EventResponse
	UID                 string               `json:"uid,omitempty"`
	ResourceVersion     string               `json:"resourceVersion,omitempty"`
	ReportingController string               `json:"reportingController,omitempty"`
	ReportingInstance   string               `json:"reportingInstance,omitempty"`
	Action              string               `json:"action,omitempty"`
	Series              *EventSeriesResponse `json:"series,omitempty"`
	Labels              map[string]string    `json:"labels"`
	Annotations         map[string]string    `json:"annotations"`
}

type EventListResponse struct {
	Items []EventResponse `json:"items"`
	Total int             `json:"total"`
}
