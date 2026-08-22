import { describe, expect, it } from "vitest";

import { buildIncidentTimeline, isRollbackHistoryEntry } from "./incident-timeline";

describe("buildIncidentTimeline", () => {
  it("always includes incident_created and orders every source chronologically", () => {
    const events = buildIncidentTimeline({
      incident: { createdAt: "2026-01-01T00:00:00Z" },
      alerts: [{ id: 1, title: "High CPU", firstSeenAt: "2026-01-01T01:00:00Z" }],
      comments: [{ id: 5, created_at: "2026-01-01T00:30:00Z" }],
    });

    expect(events.map((event) => event.type)).toEqual(["incident_created", "comment_added", "alert_fired"]);
  });

  it("omits incident_resolved entirely when the incident has no resolvedAt", () => {
    const events = buildIncidentTimeline({ incident: { createdAt: "2026-01-01T00:00:00Z" } });
    expect(events.some((event) => event.type === "incident_resolved")).toBe(false);
  });

  it("includes incident_resolved only when resolvedAt is set, at its real timestamp", () => {
    const events = buildIncidentTimeline({
      incident: { createdAt: "2026-01-01T00:00:00Z", resolvedAt: "2026-01-02T00:00:00Z" },
    });
    const resolved = events.find((event) => event.type === "incident_resolved");
    expect(resolved?.timestamp).toBe("2026-01-02T00:00:00Z");
  });

  it("emits alert_acknowledged only for alerts that have a real acknowledgedAt", () => {
    const events = buildIncidentTimeline({
      incident: { createdAt: "2026-01-01T00:00:00Z" },
      alerts: [
        { id: 1, title: "Acknowledged alert", firstSeenAt: "2026-01-01T01:00:00Z", acknowledgedAt: "2026-01-01T02:00:00Z" },
        { id: 2, title: "Unacknowledged alert", firstSeenAt: "2026-01-01T01:30:00Z" },
      ],
    });

    const acknowledged = events.filter((event) => event.type === "alert_acknowledged");
    expect(acknowledged).toHaveLength(1);
    expect(acknowledged[0]?.nav).toEqual({ kind: "alert", alertId: 1 });
  });

  it("emits deployment_occurred for every deployment with a deployment nav target", () => {
    const events = buildIncidentTimeline({
      incident: { createdAt: "2026-01-01T00:00:00Z" },
      deployments: [{ id: "dep-1", image: "web", imageTag: "v2", createdAt: "2026-01-01T03:00:00Z" }],
    });

    const deployed = events.find((event) => event.type === "deployment_occurred");
    expect(deployed?.title).toBe("Deployment occurred: web:v2");
    expect(deployed?.nav).toEqual({ kind: "deployment", deploymentId: "dep-1" });
  });

  it("detects rollback_performed via status OR changeSummary text, and ignores ordinary history rows", () => {
    const events = buildIncidentTimeline({
      incident: { createdAt: "2026-01-01T00:00:00Z" },
      deploymentHistories: [
        { id: "h1", deploymentId: "dep-1", revision: 1, status: "Succeeded", changeSummary: "Deployment created", createdAt: "2026-01-01T01:00:00Z" },
        { id: "h2", deploymentId: "dep-1", revision: 2, status: "Pending", changeSummary: "Deployment rolled back to revision 1", createdAt: "2026-01-01T02:00:00Z" },
        { id: "h3", deploymentId: "dep-1", revision: 3, status: "RolledBack", changeSummary: "Deployment status changed to RolledBack", createdAt: "2026-01-01T03:00:00Z" },
      ],
    });

    const rollbacks = events.filter((event) => event.type === "rollback_performed");
    expect(rollbacks.map((event) => event.id)).toEqual(["rollback-h2", "rollback-h3"]);
    expect(rollbacks[0]?.nav).toEqual({ kind: "deployment", deploymentId: "dep-1" });
  });

  it("emits pod_restarted only for containers with a real lastTerminationFinishedAt, never for a bare restartCount", () => {
    const events = buildIncidentTimeline({
      incident: { createdAt: "2026-01-01T00:00:00Z" },
      pods: [
        {
          name: "web-1",
          restartCount: 3,
          containerStatuses: [
            { name: "app", lastTerminationFinishedAt: "2026-01-01T04:00:00Z" },
            { name: "sidecar" },
          ],
        },
      ],
    });

    const restarts = events.filter((event) => event.type === "pod_restarted");
    expect(restarts).toHaveLength(1);
    expect(restarts[0]?.id).toBe("pod-restart-web-1-app");
  });

  it("emits comment_added for every comment at its own created_at", () => {
    const events = buildIncidentTimeline({
      incident: { createdAt: "2026-01-01T00:00:00Z" },
      comments: [{ id: 9, created_at: "2026-01-01T05:00:00Z" }],
    });

    expect(events.find((event) => event.type === "comment_added")?.timestamp).toBe("2026-01-01T05:00:00Z");
  });
});

describe("isRollbackHistoryEntry", () => {
  it("matches on status === RolledBack regardless of summary text", () => {
    expect(isRollbackHistoryEntry({ status: "RolledBack", changeSummary: "anything" })).toBe(true);
  });

  it("matches on changeSummary containing 'rolled back' case-insensitively", () => {
    expect(isRollbackHistoryEntry({ status: "Pending", changeSummary: "Deployment ROLLED BACK to revision 4" })).toBe(true);
  });

  it("does not match a normal creation/update entry", () => {
    expect(isRollbackHistoryEntry({ status: "Succeeded", changeSummary: "Deployment created" })).toBe(false);
  });
});
