package rca

import (
	"testing"
	"time"
)

func TestComputeMetricSignalsSplitsBeforeAfterAndDirection(t *testing.T) {
	splitAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	points := []MetricPoint{
		{MetricType: "RESOURCE", MetricName: "memory_usage", Unit: "MiB", Value: 100, Timestamp: splitAt.Add(-20 * time.Minute)},
		{MetricType: "RESOURCE", MetricName: "memory_usage", Unit: "MiB", Value: 120, Timestamp: splitAt.Add(-10 * time.Minute)},
		{MetricType: "RESOURCE", MetricName: "memory_usage", Unit: "MiB", Value: 300, Timestamp: splitAt.Add(10 * time.Minute)},
		{MetricType: "RESOURCE", MetricName: "memory_usage", Unit: "MiB", Value: 320, Timestamp: splitAt.Add(20 * time.Minute)},
		// A series with data on only one side of the split must be excluded.
		{MetricType: "RESOURCE", MetricName: "cpu_usage", Unit: "cores", Value: 1, Timestamp: splitAt.Add(-10 * time.Minute)},
	}

	signals := computeMetricSignals(points, splitAt)

	if len(signals) != 1 {
		t.Fatalf("expected exactly one signal (cpu_usage must be excluded, no after-split data), got %d: %+v", len(signals), signals)
	}

	signal := signals[0]
	if signal.MetricName != "memory_usage" {
		t.Fatalf("expected memory_usage signal, got %s", signal.MetricName)
	}
	if signal.Before != 110 {
		t.Fatalf("expected before average 110, got %v", signal.Before)
	}
	if signal.After != 310 {
		t.Fatalf("expected after average 310, got %v", signal.After)
	}
	if signal.Direction != "increased" {
		t.Fatalf("expected direction increased, got %s", signal.Direction)
	}
}

func TestComputeMetricSignalsStableWithinThreshold(t *testing.T) {
	splitAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	points := []MetricPoint{
		{MetricType: "RESOURCE", MetricName: "cpu_usage", Value: 100, Timestamp: splitAt.Add(-10 * time.Minute)},
		{MetricType: "RESOURCE", MetricName: "cpu_usage", Value: 105, Timestamp: splitAt.Add(10 * time.Minute)},
	}

	signals := computeMetricSignals(points, splitAt)
	if len(signals) != 1 {
		t.Fatalf("expected one signal, got %d", len(signals))
	}
	if signals[0].Direction != "stable" {
		t.Fatalf("expected a 5%% change to be reported stable, got %s (%.1f%%)", signals[0].Direction, signals[0].ChangePercent)
	}
}

func TestExtractLogSignalsFlagsKnownKeywordsOnly(t *testing.T) {
	logs := []LogLine{
		{PodName: "api-0", Line: "2026-01-01T12:00:00Z INFO request completed in 12ms"},
		{PodName: "api-0", Line: "2026-01-01T12:05:00Z ERROR OOMKilled: container exceeded memory limit"},
		{PodName: "api-1", Line: "panic: runtime error: index out of range"},
	}

	signals := extractLogSignals(logs)

	if len(signals) != 2 {
		t.Fatalf("expected 2 flagged log lines, got %d: %+v", len(signals), signals)
	}
	if signals[0].Keyword != "oomkilled" {
		t.Fatalf("expected first flagged line to match 'oomkilled', got %q", signals[0].Keyword)
	}
	if signals[1].Keyword != "panic" {
		t.Fatalf("expected second flagged line to match 'panic', got %q", signals[1].Keyword)
	}
}
