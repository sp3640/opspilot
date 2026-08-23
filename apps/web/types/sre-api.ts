export type ApplicationSLOResponse = {
  applicationId: string;
  targetPercentage: number;
  windowDays: number;
  createdAt: string;
  updatedAt: string;
};

export type ConfigureApplicationSLORequest = {
  target_percentage: number;
  window_days: number;
};

// MetricValue mirrors the backend's dto.MetricValue exactly (see
// docs/monitoring/SRE_METRICS.md): available=false means there isn't enough
// real historical data, never a fabricated number - render `formatted`
// ("Insufficient data") and `reason`, not `value`. `formula` is the backend's
// own description of how the number was derived, shown verbatim so this
// explanation never has to be duplicated (and risk drifting) in the frontend.
export type MetricValue = {
  available: boolean;
  value?: number;
  unit?: string;
  formatted: string;
  reason?: string;
  formula: string;
};

export type ApplicationSREMetricsResponse = {
  applicationId: string;
  windowDays: number;
  windowStart: string;
  windowEnd: string;
  observedDays: number;
  sloConfigured: boolean;
  currentSlo?: number;
  availability: MetricValue;
  uptime: MetricValue;
  sloCompliance: MetricValue;
  errorBudget: MetricValue;
  mttr: MetricValue;
  mtta: MetricValue;
  incidentCount: number;
};
