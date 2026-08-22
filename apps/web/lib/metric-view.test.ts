import { describe, expect, it } from "vitest";

import { formatBytes, formatChartTimestamp, formatMillicores, formatPercent, metricSource, pickLatestMetric } from "./metric-view";
import type { MetricResponse } from "@/types/metric-api";

function metric(overrides: Partial<MetricResponse> = {}): MetricResponse {
  return {
    id: "1",
    projectId: "project-1",
    clusterId: "cluster-1",
    resourceId: "cluster-1",
    resourceKind: "Cluster",
    metricType: "CPU",
    metricName: "cluster.cpu.usage.millicores",
    value: 250,
    unit: "m",
    timestamp: "2026-01-01T00:00:00Z",
    labels: { source: "metrics-server" },
    metadata: {},
    createdAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("pickLatestMetric", () => {
  it("finds the first entry matching the metric name (newest-first ordering)", () => {
    const items = [
      metric({ metricName: "cluster.cpu.usage.millicores", value: 250 }),
      metric({ metricName: "cluster.cpu.capacity.millicores", value: 2000 }),
    ];
    expect(pickLatestMetric(items, "cluster.cpu.capacity.millicores")?.value).toBe(2000);
  });

  it("returns null when no entry matches", () => {
    expect(pickLatestMetric([metric()], "does.not.exist")).toBeNull();
  });
});

describe("metricSource", () => {
  it("reads the source label", () => {
    expect(metricSource(metric({ labels: { source: "requested-capacity" } }))).toBe("requested-capacity");
  });

  it("returns null when there is no metric or no source label", () => {
    expect(metricSource(null)).toBeNull();
    expect(metricSource(metric({ labels: {} }))).toBeNull();
  });
});

describe("formatMillicores", () => {
  it("formats sub-core values as millicores", () => {
    expect(formatMillicores(250)).toBe("250m");
  });

  it("formats whole-and-above core values as cores", () => {
    expect(formatMillicores(2500)).toBe("2.50 cores");
  });

  it("reports not available for null/undefined", () => {
    expect(formatMillicores(null)).toBe("Not available");
    expect(formatMillicores(undefined)).toBe("Not available");
  });
});

describe("formatBytes", () => {
  it("scales into the appropriate unit", () => {
    expect(formatBytes(512)).toBe("512 B");
    expect(formatBytes(1024 * 1024 * 256)).toBe("256.0 MiB");
  });

  it("reports not available for null/undefined", () => {
    expect(formatBytes(null)).toBe("Not available");
  });
});

describe("formatChartTimestamp", () => {
  const iso = "2026-03-15T14:32:00.000Z";

  it("uses a time-only label when the chart span fits within a day", () => {
    const label = formatChartTimestamp(iso, 60 * 60 * 1000);
    expect(label).not.toMatch(/2026|Mar/);
  });

  it("uses a date label once the span exceeds a day", () => {
    const label = formatChartTimestamp(iso, 8 * 24 * 60 * 60 * 1000);
    expect(label).toMatch(/Mar/);
    expect(label).not.toMatch(/26|2026/);
  });

  it("includes the year for very wide spans", () => {
    const label = formatChartTimestamp(iso, 400 * 24 * 60 * 60 * 1000);
    expect(label).toMatch(/26/);
  });
});

describe("formatPercent", () => {
  it("computes a usage percentage", () => {
    expect(formatPercent(512, 1024)).toBe("50.0%");
  });

  it("reports not available when capacity is zero or missing", () => {
    expect(formatPercent(512, 0)).toBe("Not available");
    expect(formatPercent(512, null)).toBe("Not available");
    expect(formatPercent(null, 1024)).toBe("Not available");
  });
});
