package mapper

import (
	"encoding/json"

	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/models"
)

// MapMetric converts a Metric model to a MetricResponse DTO.
func MapMetric(metric models.Metric) dto.MetricResponse {
	return dto.MetricResponse{
		ID:           metric.ID.String(),
		ProjectID:    metric.ProjectID.String(),
		ClusterID:    metric.ClusterID.String(),
		ResourceID:   metric.ResourceID.String(),
		ResourceKind: metric.ResourceKind,
		MetricType:   metric.MetricType,
		MetricName:   metric.MetricName,
		Value:        metric.Value,
		Unit:         metric.Unit,
		Timestamp:    metric.Timestamp,
		Labels:       json.RawMessage(metric.Labels),
		Metadata:     json.RawMessage(metric.Metadata),
		CreatedAt:    metric.CreatedAt,
	}
}

// MapMetrics converts a slice of Metric models to a slice of MetricResponse DTOs.
func MapMetrics(metrics []models.Metric) []dto.MetricResponse {
	responses := make([]dto.MetricResponse, 0, len(metrics))
	for _, metric := range metrics {
		responses = append(responses, MapMetric(metric))
	}

	return responses
}

// MapMetricAggregatePoints converts aggregate model points into response points.
func MapMetricAggregatePoints(points []models.MetricAggregatePoint) []dto.MetricAggregatePointResponse {
	responses := make([]dto.MetricAggregatePointResponse, 0, len(points))
	for _, point := range points {
		responses = append(responses, dto.MetricAggregatePointResponse{
			Bucket: point.Bucket,
			Value:  point.Value,
			Count:  point.Count,
		})
	}

	return responses
}
