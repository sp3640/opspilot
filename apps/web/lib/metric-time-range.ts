export type MetricTimeRangeKey = "15m" | "1h" | "6h" | "24h" | "7d" | "30d";

export type MetricTimeRangeOption = {
  key: MetricTimeRangeKey;
  label: string;
  durationSeconds: number;
  interval: "minute" | "hour" | "day";
};

export const METRIC_TIME_RANGE_OPTIONS: readonly MetricTimeRangeOption[] = [
  { key: "15m", label: "Last 15 minutes", durationSeconds: 15 * 60, interval: "minute" },
  { key: "1h", label: "Last hour", durationSeconds: 60 * 60, interval: "minute" },
  { key: "6h", label: "Last 6 hours", durationSeconds: 6 * 60 * 60, interval: "hour" },
  { key: "24h", label: "Last 24 hours", durationSeconds: 24 * 60 * 60, interval: "hour" },
  { key: "7d", label: "Last 7 days", durationSeconds: 7 * 24 * 60 * 60, interval: "hour" },
  { key: "30d", label: "Last 30 days", durationSeconds: 30 * 24 * 60 * 60, interval: "day" },
];

export const DEFAULT_METRIC_TIME_RANGE: MetricTimeRangeKey = "24h";

export function findMetricTimeRangeOption(key: MetricTimeRangeKey): MetricTimeRangeOption {
  return METRIC_TIME_RANGE_OPTIONS.find((option) => option.key === key) ?? METRIC_TIME_RANGE_OPTIONS[3]!;
}

/** Resolves a time range key into concrete ISO start/end bounds and the bucketing interval to request. */
export function resolveMetricTimeRange(
  key: MetricTimeRangeKey,
  now: Date = new Date()
): { start: string; end: string; interval: "minute" | "hour" | "day" } {
  const option = findMetricTimeRangeOption(key);
  const end = now;
  const start = new Date(end.getTime() - option.durationSeconds * 1000);

  return { start: start.toISOString(), end: end.toISOString(), interval: option.interval };
}
