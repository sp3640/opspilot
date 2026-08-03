package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateMetricRequest is the validated input for metric ingestion.
type CreateMetricRequest struct {
	ProjectID    uuid.UUID       `json:"project_id" binding:"required"`
	ClusterID    uuid.UUID       `json:"cluster_id" binding:"required"`
	ResourceID   uuid.UUID       `json:"resource_id" binding:"required"`
	ResourceKind string          `json:"resource_kind" binding:"required"`
	MetricType   string          `json:"metric_type" binding:"required"`
	MetricName   string          `json:"metric_name" binding:"required,min=1,max=128"`
	Value        float64         `json:"value" binding:"required"`
	Unit         string          `json:"unit" binding:"max=32"`
	Timestamp    time.Time       `json:"timestamp" binding:"required"`
	Labels       json.RawMessage `json:"labels"`
	Metadata     json.RawMessage `json:"metadata"`
}

// MetricFilterRequest is the validated input for metric list filtering.
type MetricFilterRequest struct {
	Page         int       `json:"page"`
	Limit        int       `json:"limit"`
	Search       string    `json:"search"`
	Sort         string    `json:"sort"`
	Order        string    `json:"order"`
	ProjectID    uuid.UUID `json:"project_id"`
	ClusterID    uuid.UUID `json:"cluster_id"`
	ResourceID   uuid.UUID `json:"resource_id"`
	ResourceKind string    `json:"resource_kind"`
	MetricType   string    `json:"metric_type"`
	MetricName   string    `json:"metric_name"`
}

// MetricResponse is the canonical API representation of a stored metric.
type MetricResponse struct {
	ID           string          `json:"id"`
	ProjectID    string          `json:"projectId"`
	ClusterID    string          `json:"clusterId"`
	ResourceID   string          `json:"resourceId"`
	ResourceKind string          `json:"resourceKind"`
	MetricType   string          `json:"metricType"`
	MetricName   string          `json:"metricName"`
	Value        float64         `json:"value"`
	Unit         string          `json:"unit"`
	Timestamp    time.Time       `json:"timestamp"`
	Labels       json.RawMessage `json:"labels"`
	Metadata     json.RawMessage `json:"metadata"`
	CreatedAt    time.Time       `json:"createdAt"`
}

// MetricListResponse is the paginated metric list response.
type MetricListResponse struct {
	Items      []MetricResponse `json:"items"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	Total      int64            `json:"total"`
	TotalPages int              `json:"totalPages"`
}

// MetricAggregatePointResponse is a single bucket in a metric aggregate series.
type MetricAggregatePointResponse struct {
	Bucket time.Time `json:"bucket"`
	Value  float64   `json:"value"`
	Count  int64     `json:"count"`
}

// MetricAggregateResponse contains aggregate values and a bucketed time series.
type MetricAggregateResponse struct {
	ProjectID  string                         `json:"projectId"`
	MetricType string                         `json:"metricType"`
	MetricName string                         `json:"metricName"`
	Unit       string                         `json:"unit"`
	Interval   string                         `json:"interval"`
	StartTime  time.Time                      `json:"startTime"`
	EndTime    time.Time                      `json:"endTime"`
	Count      int64                          `json:"count"`
	Average    float64                        `json:"average"`
	Minimum    float64                        `json:"minimum"`
	Maximum    float64                        `json:"maximum"`
	Sum        float64                        `json:"sum"`
	Points     []MetricAggregatePointResponse `json:"points"`
}
