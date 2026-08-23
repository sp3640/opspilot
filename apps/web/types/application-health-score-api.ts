// GET /applications/:id/health — the unified, deterministic health score
// (Phase 21). Distinct from types/application-health-api.ts's raw-metric
// placeholder (availability/errorRate/latency/cpu/memory numeric values for
// a future APM integration): this is the already-real, computed score and
// its factor breakdown, documented entirely by the `reason` on each factor.
export type ApplicationHealthState = "HEALTHY" | "DEGRADED" | "CRITICAL" | "UNKNOWN";

export type ApplicationHealthFactorResponse = {
  key: string;
  label: string;
  state: ApplicationHealthState;
  score?: number;
  weight: number;
  reason: string;
};

export type ApplicationHealthScoreResponse = {
  applicationId: string;
  score?: number;
  state: ApplicationHealthState;
  factors: ApplicationHealthFactorResponse[];
};
