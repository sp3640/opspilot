package rca

import (
	"fmt"
	"sort"
)

// buildTimeline merges every already-real, timestamped evidence item into a
// single chronological story. Only items with a real timestamp are
// included; log lines have none from the Kubernetes API used here, so they
// never appear on the timeline (they still appear in RelevantLogs).
func buildTimeline(input Input) []TimelineEvent {
	events := make([]TimelineEvent, 0)

	events = append(events, TimelineEvent{
		Timestamp:   input.IncidentCreatedAt,
		Type:        TimelineEventIncident,
		Title:       "Incident created",
		Description: input.IncidentTitle,
		Source:      fmt.Sprintf("incident #%d", input.IncidentID),
	})

	if input.IncidentAcknowledgedAt != nil {
		events = append(events, TimelineEvent{
			Timestamp:   *input.IncidentAcknowledgedAt,
			Type:        TimelineEventIncident,
			Title:       "Incident acknowledged",
			Description: "An engineer acknowledged the incident.",
			Source:      fmt.Sprintf("incident #%d", input.IncidentID),
		})
	}

	if input.IncidentResolvedAt != nil {
		events = append(events, TimelineEvent{
			Timestamp:   *input.IncidentResolvedAt,
			Type:        TimelineEventIncident,
			Title:       "Incident resolved",
			Description: "The incident was marked resolved.",
			Source:      fmt.Sprintf("incident #%d", input.IncidentID),
		})
	}

	for _, alert := range input.Alerts {
		events = append(events, TimelineEvent{
			Timestamp:   alert.FirstSeenAt,
			Type:        TimelineEventAlert,
			Title:       "Alert fired: " + alert.Title,
			Description: fmt.Sprintf("Severity %s, status %s.", alert.Severity, alert.Status),
			Source:      fmt.Sprintf("alert #%d", alert.ID),
		})
	}

	for _, deployment := range input.Deployments {
		events = append(events, TimelineEvent{
			Timestamp:   deployment.CreatedAt,
			Type:        TimelineEventDeployment,
			Title:       "Deployment: " + deploymentLabel(deployment),
			Description: fmt.Sprintf("Status %s in %s.", deployment.Status, deployment.Environment),
			Source:      "deployment " + deployment.ID,
		})
	}

	for _, change := range input.ConfigChanges {
		events = append(events, TimelineEvent{
			Timestamp:   change.ChangedAt,
			Type:        TimelineEventConfigChange,
			Title:       fmt.Sprintf("%s: %s %s", change.Action, change.EntityType, change.EntityID),
			Description: fieldSuffix(change),
			Source:      "audit log",
		})
	}

	for _, event := range input.K8sEvents {
		if event.LastTimestamp == nil {
			continue
		}
		events = append(events, TimelineEvent{
			Timestamp:   *event.LastTimestamp,
			Type:        TimelineEventKubernetes,
			Title:       event.Reason + ": " + event.InvolvedObject,
			Description: event.Message,
			Source:      "kubernetes event",
		})
	}

	sort.Slice(events, func(i, j int) bool { return events[i].Timestamp.Before(events[j].Timestamp) })
	return events
}
