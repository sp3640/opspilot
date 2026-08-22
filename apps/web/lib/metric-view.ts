import type { MetricResponse } from "@/types/metric-api";

/**
 * Metric history is returned newest-first (ORDER BY timestamp DESC) and can
 * contain several distinct metric names under one type (e.g. both
 * "cluster.cpu.usage.millicores" and "cluster.cpu.capacity.millicores" are
 * MetricTypeCPU). This picks the latest value for one specific name.
 */
export function pickLatestMetric(items: MetricResponse[], metricName: string): MetricResponse | null {
  return items.find((item) => item.metricName === metricName) ?? null;
}

/**
 * Reads the `source` label the backend attaches to every persisted metric
 * (e.g. "metrics-server", "requested-capacity", "kubernetes-api"), used to
 * honestly distinguish live usage from an estimate rather than presenting
 * both the same way.
 */
export function metricSource(metric: MetricResponse | null): string | null {
  if (!metric) return null;
  const labels = metric.labels as Record<string, unknown> | undefined;
  const source = labels?.source;
  return typeof source === "string" ? source : null;
}

export function formatMillicores(value: number | null | undefined): string {
  if (value === null || value === undefined) return "Not available";
  if (value >= 1000) return `${(value / 1000).toFixed(2)} cores`;
  return `${Math.round(value)}m`;
}

const BYTE_UNITS = ["B", "KiB", "MiB", "GiB", "TiB"];

export function formatBytes(value: number | null | undefined): string {
  if (value === null || value === undefined) return "Not available";
  if (value <= 0) return "0 B";

  let unitIndex = 0;
  let remaining = value;
  while (remaining >= 1024 && unitIndex < BYTE_UNITS.length - 1) {
    remaining /= 1024;
    unitIndex++;
  }

  return `${remaining.toFixed(unitIndex === 0 ? 0 : 1)} ${BYTE_UNITS[unitIndex]}`;
}

export function formatPercent(usage: number | null | undefined, capacity: number | null | undefined): string {
  if (usage === null || usage === undefined || !capacity) return "Not available";
  return `${((usage / capacity) * 100).toFixed(1)}%`;
}

export function formatCount(value: number | null | undefined): string {
  if (value === null || value === undefined) return "Not available";
  return String(Math.round(value));
}
