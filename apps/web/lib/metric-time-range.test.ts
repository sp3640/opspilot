import { describe, expect, it } from "vitest";

import { findMetricTimeRangeOption, METRIC_TIME_RANGE_OPTIONS, resolveMetricTimeRange } from "./metric-time-range";

describe("METRIC_TIME_RANGE_OPTIONS", () => {
  it("exposes exactly the six required ranges in order", () => {
    expect(METRIC_TIME_RANGE_OPTIONS.map((option) => option.key)).toEqual(["15m", "1h", "6h", "24h", "7d", "30d"]);
  });
});

describe("findMetricTimeRangeOption", () => {
  it("finds the matching option", () => {
    expect(findMetricTimeRangeOption("7d").durationSeconds).toBe(7 * 24 * 60 * 60);
  });
});

describe("resolveMetricTimeRange", () => {
  it("computes start/end bounds relative to now", () => {
    const now = new Date("2026-01-15T12:00:00.000Z");
    const result = resolveMetricTimeRange("1h", now);
    expect(result.end).toBe("2026-01-15T12:00:00.000Z");
    expect(result.start).toBe("2026-01-15T11:00:00.000Z");
    expect(result.interval).toBe("minute");
  });

  it("picks a day interval for the 30d range", () => {
    const result = resolveMetricTimeRange("30d", new Date("2026-01-31T00:00:00.000Z"));
    expect(result.interval).toBe("day");
    expect(result.start).toBe("2026-01-01T00:00:00.000Z");
  });

  it("picks an hour interval for 6h/24h/7d", () => {
    expect(resolveMetricTimeRange("6h").interval).toBe("hour");
    expect(resolveMetricTimeRange("24h").interval).toBe("hour");
    expect(resolveMetricTimeRange("7d").interval).toBe("hour");
  });
});
