import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { computeDeploymentWindows, evaluateDeploymentHealthCorrelation } from "./deployment-health-correlation";

const NOW = new Date("2026-01-01T12:00:00Z");

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(NOW);
});

afterEach(() => {
  vi.useRealTimers();
});

// Deployment started an hour ago and completed 50 minutes ago, well before
// "now" - so with the default 30-minute window, both before/after windows
// have fully elapsed by the time these tests run.
const ELAPSED_DEPLOYMENT = {
  startedAt: "2026-01-01T11:00:00Z",
  completedAt: "2026-01-01T11:10:00Z",
  createdAt: "2026-01-01T10:59:00Z",
};

describe("computeDeploymentWindows", () => {
  it("derives the before window from startedAt and the after window from completedAt", () => {
    const { beforeWindow, afterWindow } = computeDeploymentWindows(ELAPSED_DEPLOYMENT, 30);
    expect(beforeWindow).toEqual({ start: "2026-01-01T10:30:00.000Z", end: "2026-01-01T11:00:00.000Z" });
    expect(afterWindow).toEqual({ start: "2026-01-01T11:10:00.000Z", end: "2026-01-01T11:40:00.000Z" });
  });

  it("falls back to createdAt/startedAt when startedAt/completedAt are absent", () => {
    const { beforeWindow, afterWindow } = computeDeploymentWindows({ createdAt: "2026-01-01T11:00:00Z" }, 30);
    expect(beforeWindow.end).toBe("2026-01-01T11:00:00.000Z");
    expect(afterWindow.start).toBe("2026-01-01T11:00:00.000Z");
  });
});

describe("evaluateDeploymentHealthCorrelation", () => {
  it("reports no_degradation_detected when the after window has elapsed and no signal worsened", () => {
    const result = evaluateDeploymentHealthCorrelation({ deployment: ELAPSED_DEPLOYMENT });
    expect(result.afterWindowElapsed).toBe(true);
    expect(result.alerts).toEqual({ status: "available", before: 0, after: 0, degraded: false });
    expect(result.incidents).toEqual({ status: "available", before: 0, after: 0, degraded: false });
    expect(result.pods).toEqual({ status: "unavailable" });
    expect(result.verdict).toBe("no_degradation_detected");
    expect(result.reasons).toEqual([]);
  });

  it("reports insufficient_data when the after window has not fully elapsed yet, even with zero findings", () => {
    const recentDeployment = { startedAt: "2026-01-01T11:55:00Z", completedAt: "2026-01-01T11:56:00Z", createdAt: "2026-01-01T11:54:00Z" };
    const result = evaluateDeploymentHealthCorrelation({ deployment: recentDeployment });
    expect(result.afterWindowElapsed).toBe(false);
    expect(result.verdict).toBe("insufficient_data");
  });

  it("detects potential_degradation when more alerts fired after deployment than before", () => {
    const result = evaluateDeploymentHealthCorrelation({
      deployment: ELAPSED_DEPLOYMENT,
      alerts: [{ firstSeenAt: "2026-01-01T10:45:00Z" }, { firstSeenAt: "2026-01-01T11:15:00Z" }, { firstSeenAt: "2026-01-01T11:20:00Z" }],
    });
    expect(result.alerts).toEqual({ status: "available", before: 1, after: 2, degraded: true });
    expect(result.verdict).toBe("potential_degradation");
    expect(result.reasons[0]).toContain("2 alert(s) fired");
  });

  it("does not flag degradation when alert counts stay flat or improve after deployment", () => {
    const result = evaluateDeploymentHealthCorrelation({
      deployment: ELAPSED_DEPLOYMENT,
      alerts: [{ firstSeenAt: "2026-01-01T10:45:00Z" }, { firstSeenAt: "2026-01-01T10:50:00Z" }],
    });
    expect(result.alerts).toEqual({ status: "available", before: 2, after: 0, degraded: false });
    expect(result.verdict).toBe("no_degradation_detected");
  });

  it("detects potential_degradation when more incidents opened after deployment than before", () => {
    const result = evaluateDeploymentHealthCorrelation({
      deployment: ELAPSED_DEPLOYMENT,
      incidents: [{ createdAt: "2026-01-01T11:12:00Z" }],
    });
    expect(result.incidents).toEqual({ status: "available", before: 0, after: 1, degraded: true });
    expect(result.verdict).toBe("potential_degradation");
  });

  it("reports pods as post_only and flags degradation for unready pods", () => {
    const result = evaluateDeploymentHealthCorrelation({
      deployment: ELAPSED_DEPLOYMENT,
      pods: [{ ready: true }, { ready: false }],
    });
    expect(result.pods).toEqual({ status: "post_only", totalPods: 2, unreadyPods: 1, restartedSinceDeployment: 0, degraded: true });
    expect(result.verdict).toBe("potential_degradation");
    expect(result.reasons[0]).toContain("not ready");
  });

  it("flags degradation for a pod that restarted after the deployment completed", () => {
    const result = evaluateDeploymentHealthCorrelation({
      deployment: ELAPSED_DEPLOYMENT,
      pods: [{ ready: true, containerStatuses: [{ lastTerminationFinishedAt: "2026-01-01T11:20:00Z" }] }],
    });
    expect(result.pods).toMatchObject({ status: "post_only", restartedSinceDeployment: 1, degraded: true });
    expect(result.reasons.some((reason) => reason.includes("restarted"))).toBe(true);
  });

  it("does not count a restart that happened before the deployment completed", () => {
    const result = evaluateDeploymentHealthCorrelation({
      deployment: ELAPSED_DEPLOYMENT,
      pods: [{ ready: true, containerStatuses: [{ lastTerminationFinishedAt: "2026-01-01T10:00:00Z" }] }],
    });
    expect(result.pods).toMatchObject({ restartedSinceDeployment: 0, degraded: false });
    expect(result.verdict).toBe("no_degradation_detected");
  });

  it("reports cluster CPU/memory as unavailable when neither window's aggregate is supplied", () => {
    const result = evaluateDeploymentHealthCorrelation({ deployment: ELAPSED_DEPLOYMENT });
    expect(result.cpu).toEqual({ status: "unavailable" });
    expect(result.memory).toEqual({ status: "unavailable" });
  });

  it("computes a cluster CPU change percentage from real aggregate averages, without affecting the verdict", () => {
    const result = evaluateDeploymentHealthCorrelation({
      deployment: ELAPSED_DEPLOYMENT,
      cpuBefore: { count: 10, average: 100, unit: "millicores" },
      cpuAfter: { count: 10, average: 150, unit: "millicores" },
    });
    expect(result.cpu).toEqual({ status: "available", scope: "cluster", unit: "millicores", before: 100, after: 150, changePercent: 50 });
    // Cluster-wide metrics never drive the verdict on their own - only
    // application-scoped signals (alerts/incidents/pods) do.
    expect(result.verdict).toBe("no_degradation_detected");
  });

  it("treats a window with zero collected metric points as null rather than a fabricated zero", () => {
    const result = evaluateDeploymentHealthCorrelation({
      deployment: ELAPSED_DEPLOYMENT,
      cpuBefore: { count: 0, average: 0, unit: "millicores" },
      cpuAfter: { count: 5, average: 200, unit: "millicores" },
    });
    expect(result.cpu).toEqual({ status: "available", scope: "cluster", unit: "millicores", before: null, after: 200, changePercent: null });
  });
});
