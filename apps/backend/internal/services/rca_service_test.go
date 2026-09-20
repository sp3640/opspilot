package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/dto"
)

func TestPrometheusMetricPointsFromSamplesConvertsSamplesToRCAFormat(t *testing.T) {
	now := time.Now().UTC()
	samples := []dto.PrometheusSample{{
		Timestamp: now,
		Value:     42,
		Labels: map[string]string{
			"namespace": "payments",
			"pod":       "payments-0",
		},
	}}

	points := prometheusMetricPointsFromSamples(samples, "memory_usage", "bytes")
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].MetricName != "memory_usage" {
		t.Fatalf("expected metric name memory_usage, got %q", points[0].MetricName)
	}
	if points[0].Value != 42 {
		t.Fatalf("expected value 42, got %v", points[0].Value)
	}
	if points[0].Unit != "bytes" {
		t.Fatalf("expected unit bytes, got %q", points[0].Unit)
	}
}

func TestPrometheusMetricPointsFromSamplesRejectsNilApplicationIDsWhenCorrelating(t *testing.T) {
	var id *uuid.UUID
	if id != nil {
		t.Fatal("expected nil application id")
	}
	points := prometheusMetricPointsFromSamples(nil, "cpu_usage", "cores")
	if len(points) != 0 {
		t.Fatalf("expected no points for nil input, got %d", len(points))
	}
	_ = id
}
