"use client";

import { MetricTimeSeriesChart } from "@/components/metrics/metric-time-series-chart";

const UNAVAILABLE_REASON =
  "No application observability provider (e.g. Prometheus/APM) is connected yet, so this application has no per-application metric history.";

const APPLICATION_CHARTS = ["Request rate", "Error rate", "Latency", "CPU", "Memory"];

/**
 * Application: Request rate, Error rate, Latency, CPU, Memory.
 *
 * The Metric model has no per-application dimension today (Phase 9/12) - no
 * provider populates request rate, error rate, latency, or per-application
 * CPU/memory. Rather than omitting these charts or inventing data, they are
 * shown in an explicit "not available" state so the dashboard shape is
 * ready the moment a real provider (e.g. Prometheus) is connected.
 */
export function ApplicationMetricsDashboard() {
  return (
    <div className="space-y-4">
      <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        {UNAVAILABLE_REASON}
      </p>
      <div className="grid gap-4 sm:grid-cols-2">
        {APPLICATION_CHARTS.map((title) => (
          <MetricTimeSeriesChart key={title} title={title} points={[]} unavailableReason={UNAVAILABLE_REASON} />
        ))}
      </div>
    </div>
  );
}
