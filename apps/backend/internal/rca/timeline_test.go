package rca

import (
	"testing"
	"time"
)

func TestBuildTimelineIsChronological(t *testing.T) {
	incidentAt := baseIncidentTime()
	resolvedAt := incidentAt.Add(30 * time.Minute)

	input := Input{
		IncidentID:         42,
		IncidentTitle:      "checkout errors",
		IncidentCreatedAt:  incidentAt,
		IncidentResolvedAt: &resolvedAt,
		Alerts: []AlertRecord{
			{ID: 1, Title: "high error rate", FirstSeenAt: incidentAt.Add(-5 * time.Minute)},
		},
		Deployments: []DeploymentRecord{
			{ID: "dep-1", ImageTag: "v1.8.2", CreatedAt: incidentAt.Add(-10 * time.Minute), Status: "Succeeded", Environment: "production"},
		},
	}

	timeline := buildTimeline(input)

	if len(timeline) != 4 {
		t.Fatalf("expected 4 timeline events, got %d: %+v", len(timeline), timeline)
	}

	for i := 1; i < len(timeline); i++ {
		if timeline[i].Timestamp.Before(timeline[i-1].Timestamp) {
			t.Fatalf("timeline is not chronologically ordered at index %d: %+v", i, timeline)
		}
	}

	if timeline[0].Type != TimelineEventDeployment {
		t.Fatalf("expected the earliest event to be the deployment, got %s", timeline[0].Type)
	}
	if timeline[len(timeline)-1].Type != TimelineEventIncident {
		t.Fatalf("expected the latest event to be incident resolution, got %s", timeline[len(timeline)-1].Type)
	}
}
