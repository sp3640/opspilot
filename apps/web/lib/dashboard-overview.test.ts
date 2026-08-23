import { describe, expect, it } from "vitest";

import { buildDashboardOverview } from "./dashboard-overview";

describe("buildDashboardOverview", () => {
  it("reports UNKNOWN overall health when no real data was loaded at all", () => {
    const result = buildDashboardOverview({});
    expect(result.overallHealth).toBe("UNKNOWN");
    expect(result.counts).toEqual({
      applications: null,
      clusters: null,
      activeAlerts: null,
      criticalAlerts: null,
      activeIncidents: null,
      recentDeployments: null,
      failedDeployments: null,
      unhealthyPods: null,
    });
    expect(result.criticalProblems).toEqual([]);
    expect(result.warnings).toEqual([]);
  });

  it("reports HEALTHY when real data loaded and nothing is wrong", () => {
    const result = buildDashboardOverview({
      applications: { total: 5 },
      clusters: [{ id: "c1", name: "prod", status: "HEALTHY" }],
      alerts: [],
      incidents: [],
      recentDeployments: [{ id: "d1", image: "web", imageTag: "v1", status: "Succeeded", createdAt: "2026-01-01T00:00:00Z" }],
    });
    expect(result.overallHealth).toBe("HEALTHY");
    expect(result.counts.applications).toBe(5);
    expect(result.counts.clusters).toBe(1);
    expect(result.counts.recentDeployments).toBe(1);
    expect(result.counts.failedDeployments).toBe(0);
  });

  it("promotes a disconnected cluster to a critical problem and CRITICAL overall health", () => {
    const result = buildDashboardOverview({
      clusters: [
        { id: "c1", name: "prod", status: "HEALTHY" },
        { id: "c2", name: "staging", status: "INVALID" },
      ],
    });
    expect(result.overallHealth).toBe("CRITICAL");
    expect(result.criticalProblems).toHaveLength(1);
    expect(result.criticalProblems[0]).toMatchObject({ title: "staging disconnected", href: "/clusters" });
  });

  it("splits alerts into critical problems vs warnings by severity, counting only active statuses", () => {
    const result = buildDashboardOverview({
      alerts: [
        { id: 1, title: "Payment API degraded", severity: "CRITICAL", status: "OPEN", firstSeenAt: "2026-01-01T00:00:00Z" },
        { id: 2, title: "High latency", severity: "MEDIUM", status: "ACKNOWLEDGED", firstSeenAt: "2026-01-01T01:00:00Z" },
        { id: 3, title: "Resolved noise", severity: "CRITICAL", status: "RESOLVED", firstSeenAt: "2026-01-01T02:00:00Z" },
      ],
    });
    expect(result.counts.activeAlerts).toBe(2);
    expect(result.counts.criticalAlerts).toBe(1);
    expect(result.criticalProblems.map((p) => p.id)).toEqual(["alert-1"]);
    expect(result.warnings.map((p) => p.id)).toEqual(["alert-2"]);
    expect(result.overallHealth).toBe("CRITICAL");
  });

  it("routes incidents to critical vs warning by severity, and links to the incident's own page", () => {
    const result = buildDashboardOverview({
      incidents: [
        { id: 10, title: "Database outage", severity: "P0", status: "OPEN", createdAt: "2026-01-01T00:00:00Z" },
        { id: 11, title: "Minor blip", severity: "P3", status: "INVESTIGATING", createdAt: "2026-01-01T01:00:00Z" },
        { id: 12, title: "Old resolved incident", severity: "P0", status: "RESOLVED", createdAt: "2026-01-01T02:00:00Z" },
      ],
    });
    expect(result.counts.activeIncidents).toBe(2);
    expect(result.criticalProblems).toEqual([
      { id: "incident-10", title: "Database outage", description: "Incident — P0", href: "/incidents/10" },
    ]);
    expect(result.warnings).toEqual([
      { id: "incident-11", title: "Minor blip", description: "Incident — P3", href: "/incidents/11" },
    ]);
  });

  it("flags failed deployments as critical problems with a readable image:tag label", () => {
    const result = buildDashboardOverview({
      recentDeployments: [
        { id: "d1", image: "ghcr.io/opspilot/api", imageTag: "v1.8.2", status: "Failed", createdAt: "2026-01-01T00:00:00Z" },
        { id: "d2", image: "ghcr.io/opspilot/api", imageTag: "v1.8.1", status: "Succeeded", createdAt: "2026-01-01T01:00:00Z" },
      ],
    });
    expect(result.counts.failedDeployments).toBe(1);
    expect(result.criticalProblems[0]?.title).toBe("Deployment ghcr.io/opspilot/api:v1.8.2 failed");
    expect(result.overallHealth).toBe("CRITICAL");
  });

  it("only counts unhealthy pods for clusters that were actually checked", () => {
    const result = buildDashboardOverview({
      unhealthyPodsByCluster: [
        { clusterId: "c1", clusterName: "prod", unhealthyCount: 3, checked: true },
        { clusterId: "c2", clusterName: "staging", unhealthyCount: 0, checked: false },
      ],
    });
    expect(result.counts.unhealthyPods).toBe(3);
    expect(result.warnings).toEqual([
      {
        id: "pods-c1",
        title: "3 unhealthy pod(s) in prod",
        description: "Pods that are not Running/Succeeded, or have restarted.",
        href: "/clusters",
      },
    ]);
    expect(result.overallHealth).toBe("DEGRADED");
  });

  it("leaves unhealthyPods null when no cluster could actually be checked", () => {
    const result = buildDashboardOverview({
      unhealthyPodsByCluster: [{ clusterId: "c1", clusterName: "prod", unhealthyCount: 0, checked: false }],
    });
    expect(result.counts.unhealthyPods).toBeNull();
  });

  it("merges alerts, incidents, and deployments into one chronologically sorted recent activity feed", () => {
    const result = buildDashboardOverview({
      alerts: [{ id: 1, title: "Alert A", severity: "LOW", status: "RESOLVED", firstSeenAt: "2026-01-01T02:00:00Z" }],
      incidents: [{ id: 5, title: "Incident A", severity: "P3", status: "RESOLVED", createdAt: "2026-01-01T03:00:00Z" }],
      recentDeployments: [{ id: "d1", image: "web", imageTag: "v1", status: "Succeeded", createdAt: "2026-01-01T01:00:00Z" }],
    });
    expect(result.recentActivity.map((item) => item.id)).toEqual(["incident-5", "alert-1", "deployment-d1"]);
  });

  it("caps recent activity to the requested limit", () => {
    const alerts = Array.from({ length: 5 }, (_, index) => ({
      id: index,
      title: `Alert ${index}`,
      severity: "LOW",
      status: "OPEN",
      firstSeenAt: new Date(2026, 0, index + 1).toISOString(),
    }));
    const result = buildDashboardOverview({ alerts, recentActivityLimit: 2 });
    expect(result.recentActivity).toHaveLength(2);
  });

  it("does not fabricate a warning/critical problem from data it never received", () => {
    const result = buildDashboardOverview({ applications: { total: 3 } });
    expect(result.criticalProblems).toEqual([]);
    expect(result.warnings).toEqual([]);
    // Only applications loaded; health is HEALTHY (real data, no problems
    // found in it) rather than UNKNOWN, since applications alone is real.
    expect(result.overallHealth).toBe("HEALTHY");
  });
});
