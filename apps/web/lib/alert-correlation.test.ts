import { describe, expect, it } from "vitest";

import { findRelatedAlerts, formatDuration, parseAlertMetadata, parseResourceIdentifier } from "./alert-correlation";
import type { AlertResponse } from "@/types/alert-api";

function alert(overrides: Partial<AlertResponse> = {}): AlertResponse {
  return {
    id: 1,
    projectId: "project-1",
    title: "Test alert",
    description: "",
    severity: "HIGH",
    status: "OPEN",
    source: "KUBERNETES",
    resourceType: "POD",
    resourceId: "default/web-1",
    fingerprint: "abc",
    occurrenceCount: 1,
    labels: {},
    metadata: {},
    firstSeenAt: "2026-01-01T00:00:00Z",
    lastSeenAt: "2026-01-01T00:00:00Z",
    createdBy: 1,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("parseAlertMetadata", () => {
  it("extracts known fields and ignores unknown ones", () => {
    const result = parseAlertMetadata({ condition: "cpu_threshold", currentValue: 91.2, threshold: 90, unit: "%", clusterId: "c1", clusterName: "prod", applicationId: "app-1", extra: "ignored" });
    expect(result).toEqual({ condition: "cpu_threshold", currentValue: 91.2, threshold: 90, unit: "%", clusterId: "c1", clusterName: "prod", applicationId: "app-1" });
  });

  it("returns an empty object for missing/malformed metadata", () => {
    expect(parseAlertMetadata(undefined)).toEqual({});
    expect(parseAlertMetadata({})).toEqual({});
    expect(parseAlertMetadata({ currentValue: "not-a-number" })).toEqual({ currentValue: undefined, condition: undefined, threshold: undefined, unit: undefined, clusterId: undefined, clusterName: undefined, applicationId: undefined });
  });
});

describe("parseResourceIdentifier", () => {
  it("splits namespace/name for POD and DEPLOYMENT", () => {
    expect(parseResourceIdentifier("POD", "default/web-1")).toEqual({ namespace: "default", name: "web-1" });
    expect(parseResourceIdentifier("DEPLOYMENT", "default/api")).toEqual({ namespace: "default", name: "api" });
  });

  it("treats CLUSTER/NODE resourceIds as a bare name with no namespace", () => {
    expect(parseResourceIdentifier("CLUSTER", "cluster-abc")).toEqual({ name: "cluster-abc" });
    expect(parseResourceIdentifier("NODE", "node-1")).toEqual({ name: "node-1" });
  });

  it("falls back to the raw string when a POD id has no slash", () => {
    expect(parseResourceIdentifier("POD", "web-1")).toEqual({ name: "web-1" });
  });
});

describe("formatDuration", () => {
  it("formats seconds/minutes/hours/days appropriately", () => {
    expect(formatDuration("2026-01-01T00:00:00Z", "2026-01-01T00:00:45Z")).toBe("45s");
    expect(formatDuration("2026-01-01T00:00:00Z", "2026-01-01T00:05:30Z")).toBe("5m 30s");
    expect(formatDuration("2026-01-01T00:00:00Z", "2026-01-01T02:15:00Z")).toBe("2h 15m");
    expect(formatDuration("2026-01-01T00:00:00Z", "2026-01-03T01:00:00Z")).toBe("2d 1h");
  });
});

describe("findRelatedAlerts", () => {
  it("matches only alerts sharing the same resourceType+resourceId", () => {
    const current = alert({ id: 1, resourceType: "POD", resourceId: "default/web-1" });
    const candidates = [
      alert({ id: 2, resourceType: "POD", resourceId: "default/web-1" }),
      alert({ id: 3, resourceType: "POD", resourceId: "default/web-2" }),
      alert({ id: 1, resourceType: "POD", resourceId: "default/web-1" }), // itself
    ];

    const related = findRelatedAlerts(current, candidates);
    expect(related.map((a) => a.id)).toEqual([2]);
  });
});
