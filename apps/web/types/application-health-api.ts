/**
 * Clean seam for the future Prometheus-backed metrics phase. No backend
 * endpoint provides these yet, so every field is `null` today and every
 * consumer must render "Not available" rather than a fabricated number.
 * When a real `GET /applications/:id/metrics` endpoint exists, only the
 * data source needs to change (see `useApplicationHealthMetrics`) — this
 * shape and the components that render it stay the same.
 */
export type ApplicationHealthMetric = {
  value: number | null;
  unit: string | null;
};

export type ApplicationHealthMetrics = {
  availability: ApplicationHealthMetric;
  errorRate: ApplicationHealthMetric;
  latency: ApplicationHealthMetric;
  cpu: ApplicationHealthMetric;
  memory: ApplicationHealthMetric;
};

export const NOT_AVAILABLE_METRIC: ApplicationHealthMetric = { value: null, unit: null };

export const NOT_AVAILABLE_APPLICATION_HEALTH_METRICS: ApplicationHealthMetrics = {
  availability: NOT_AVAILABLE_METRIC,
  errorRate: NOT_AVAILABLE_METRIC,
  latency: NOT_AVAILABLE_METRIC,
  cpu: NOT_AVAILABLE_METRIC,
  memory: NOT_AVAILABLE_METRIC,
};
