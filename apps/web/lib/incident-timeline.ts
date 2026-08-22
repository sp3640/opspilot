export type TimelineEventType =
  | "incident_created"
  | "alert_fired"
  | "alert_acknowledged"
  | "deployment_occurred"
  | "rollback_performed"
  | "pod_restarted"
  | "comment_added"
  | "incident_resolved";

export type TimelineNav =
  | { kind: "alert"; alertId: number }
  | { kind: "deployment"; deploymentId: string };

export type TimelineEvent = {
  id: string;
  type: TimelineEventType;
  timestamp: string;
  title: string;
  description?: string;
  nav?: TimelineNav;
};

type IncidentLike = { createdAt: string; resolvedAt?: string };
type AlertLike = { id: number; title: string; firstSeenAt: string; acknowledgedAt?: string };
type DeploymentLike = { id: string; image: string; imageTag: string; createdAt: string };
type DeploymentHistoryLike = {
  id: string;
  deploymentId: string;
  revision: number;
  status: string;
  changeSummary: string;
  createdAt: string;
};
type CommentLike = { id: number; created_at: string };
type PodContainerLike = { name: string; lastTerminationFinishedAt?: string };
type PodLike = { name: string; restartCount: number; containerStatuses?: PodContainerLike[] };

/**
 * A rollback is only ever visible in DeploymentHistory - the rollback
 * endpoint resets the live Deployment's own status back to "Pending" and
 * the change is recorded solely in the new history row's status/summary.
 */
export function isRollbackHistoryEntry(entry: Pick<DeploymentHistoryLike, "status" | "changeSummary">): boolean {
  return entry.status === "RolledBack" || /rolled back/i.test(entry.changeSummary ?? "");
}

/**
 * Merges every real, already-fetched incident-related event source into one
 * chronological timeline. Nothing here is synthesized - each entry traces
 * back to a genuine timestamp already present on the incident, an attached
 * alert, a deployment/history row, a comment, or a live pod's own
 * lastTerminationFinishedAt. Callers omit a source entirely (undefined)
 * when that data isn't available rather than passing a fabricated value.
 */
export function buildIncidentTimeline(input: {
  incident: IncidentLike;
  alerts?: AlertLike[];
  deployments?: DeploymentLike[];
  deploymentHistories?: DeploymentHistoryLike[];
  comments?: CommentLike[];
  pods?: PodLike[];
}): TimelineEvent[] {
  const events: TimelineEvent[] = [];

  events.push({
    id: "incident-created",
    type: "incident_created",
    timestamp: input.incident.createdAt,
    title: "Incident created",
  });

  if (input.incident.resolvedAt) {
    events.push({
      id: "incident-resolved",
      type: "incident_resolved",
      timestamp: input.incident.resolvedAt,
      title: "Incident resolved",
    });
  }

  for (const alert of input.alerts ?? []) {
    events.push({
      id: `alert-fired-${alert.id}`,
      type: "alert_fired",
      timestamp: alert.firstSeenAt,
      title: `Alert fired: ${alert.title}`,
      nav: { kind: "alert", alertId: alert.id },
    });

    if (alert.acknowledgedAt) {
      events.push({
        id: `alert-acknowledged-${alert.id}`,
        type: "alert_acknowledged",
        timestamp: alert.acknowledgedAt,
        title: `Alert acknowledged: ${alert.title}`,
        nav: { kind: "alert", alertId: alert.id },
      });
    }
  }

  for (const deployment of input.deployments ?? []) {
    events.push({
      id: `deployment-${deployment.id}`,
      type: "deployment_occurred",
      timestamp: deployment.createdAt,
      title: `Deployment occurred: ${deployment.image}${deployment.imageTag ? `:${deployment.imageTag}` : ""}`,
      nav: { kind: "deployment", deploymentId: deployment.id },
    });
  }

  for (const entry of input.deploymentHistories ?? []) {
    if (!isRollbackHistoryEntry(entry)) continue;
    events.push({
      id: `rollback-${entry.id}`,
      type: "rollback_performed",
      timestamp: entry.createdAt,
      title: `Rollback performed (revision ${entry.revision})`,
      description: entry.changeSummary,
      nav: { kind: "deployment", deploymentId: entry.deploymentId },
    });
  }

  for (const comment of input.comments ?? []) {
    events.push({
      id: `comment-${comment.id}`,
      type: "comment_added",
      timestamp: comment.created_at,
      title: "Comment added",
    });
  }

  for (const pod of input.pods ?? []) {
    for (const container of pod.containerStatuses ?? []) {
      if (!container.lastTerminationFinishedAt) continue;
      events.push({
        id: `pod-restart-${pod.name}-${container.name}`,
        type: "pod_restarted",
        timestamp: container.lastTerminationFinishedAt,
        title: `Pod restarted: ${pod.name}`,
        description: `Container "${container.name}" restarted (restart count: ${pod.restartCount})`,
      });
    }
  }

  return events.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
}
