package services

import (
	"testing"
	"time"

	"github.com/sp3640/opspilot/backend/internal/constants"
	"github.com/sp3640/opspilot/backend/internal/models"
)

func TestComputeDowntimeSecondsMergesOverlappingIntervals(t *testing.T) {
	windowStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	resolvedA := windowStart.Add(2 * time.Hour)
	resolvedB := windowStart.Add(3 * time.Hour)

	incidents := []models.Incident{
		{Severity: constants.SeverityP0, CreatedAt: windowStart.Add(1 * time.Hour), ResolvedAt: &resolvedA},
		// Overlaps the first interval entirely - must not double-count.
		{Severity: constants.SeverityP1, CreatedAt: windowStart.Add(90 * time.Minute), ResolvedAt: &resolvedB},
		// A P3 incident never counts as downtime.
		{Severity: constants.SeverityP3, CreatedAt: windowStart.Add(4 * time.Hour), ResolvedAt: &resolvedB},
	}

	got := computeDowntimeSeconds(incidents, windowStart, now)
	want := 2 * time.Hour.Seconds() // 1h00 -> 3h00 merged, not 1h+1.5h separately
	if got != want {
		t.Fatalf("expected merged downtime of %v seconds, got %v", want, got)
	}
}

func TestComputeDowntimeSecondsClipsToWindow(t *testing.T) {
	windowStart := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

	// Started well before the window, still unresolved (open-ended -> now).
	incidents := []models.Incident{
		{Severity: constants.SeverityP0, CreatedAt: windowStart.Add(-48 * time.Hour), ResolvedAt: nil},
	}

	got := computeDowntimeSeconds(incidents, windowStart, now)
	want := now.Sub(windowStart).Seconds()
	if got != want {
		t.Fatalf("expected downtime clipped to the full window (%v), got %v", want, got)
	}
}

func TestComputeMTTRAveragesResolvedIncidentsOnly(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	resolved1 := base.Add(30 * time.Minute)
	resolved2 := base.Add(90 * time.Minute)

	incidents := []models.Incident{
		{CreatedAt: base, ResolvedAt: &resolved1},
		{CreatedAt: base, ResolvedAt: &resolved2},
		{CreatedAt: base, ResolvedAt: nil}, // unresolved - excluded, not treated as zero
	}

	result := computeMTTR(incidents)
	if !result.Available {
		t.Fatalf("expected MTTR to be available with 2 resolved incidents")
	}
	if result.Value != 60 {
		t.Fatalf("expected average MTTR of 60 minutes, got %v", result.Value)
	}
}

func TestComputeMTTRInsufficientDataWhenNoneResolved(t *testing.T) {
	incidents := []models.Incident{{CreatedAt: time.Now(), ResolvedAt: nil}}
	result := computeMTTR(incidents)
	if result.Available {
		t.Fatalf("expected MTTR to be unavailable when no incidents are resolved")
	}
	if result.Formatted != "Insufficient data" {
		t.Fatalf("expected 'Insufficient data' formatted value, got %q", result.Formatted)
	}
}

func TestComputeMTTAAveragesAcknowledgedIncidentsOnly(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ack := base.Add(10 * time.Minute)

	incidents := []models.Incident{
		{CreatedAt: base, AcknowledgedAt: &ack},
		{CreatedAt: base, AcknowledgedAt: nil}, // never acknowledged - excluded
	}

	result := computeMTTA(incidents)
	if !result.Available {
		t.Fatalf("expected MTTA to be available with 1 acknowledged incident")
	}
	if result.Value != 10 {
		t.Fatalf("expected MTTA of 10 minutes, got %v", result.Value)
	}
}

func TestComputeMTTAInsufficientDataWhenNoneAcknowledged(t *testing.T) {
	incidents := []models.Incident{{CreatedAt: time.Now(), AcknowledgedAt: nil}}
	result := computeMTTA(incidents)
	if result.Available {
		t.Fatalf("expected MTTA to be unavailable when nothing is acknowledged")
	}
}

func TestFormatMinutes(t *testing.T) {
	cases := []struct {
		minutes float64
		want    string
	}{
		{30, "30m"},
		{90, "1h 30m"},
		{60 * 30, "1d 6h"},
	}
	for _, tc := range cases {
		if got := formatMinutes(tc.minutes); got != tc.want {
			t.Fatalf("formatMinutes(%v) = %q, want %q", tc.minutes, got, tc.want)
		}
	}
}
