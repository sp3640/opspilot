// Mirrors apps/backend/internal/dto/prometheus.go (PrometheusQueryResponse
// and friends) - the response shape of GET
// /integrations/:id/prometheus/query. Only fields the backend actually
// returns are declared here; correlation is a best-effort, organization-wide
// enrichment computed by IntegrationService.CorrelatePrometheusSamples, not
// something this file should ever expand on its own.

export type PrometheusSample = {
  timestamp: string;
  value: number;
  labels?: Record<string, string>;
};

export type PrometheusResourceRef = {
  id: string;
  kind: string;
  name: string;
  namespace?: string;
  cluster?: string;
};

export type PrometheusApplicationRef = {
  id: string;
  name: string;
};

export type PrometheusCorrelationResult = {
  available: boolean;
  reason?: string;
  resource?: PrometheusResourceRef;
  application?: PrometheusApplicationRef;
};

export type PrometheusQueryResponse = {
  query: string;
  samples: PrometheusSample[];
  correlation: PrometheusCorrelationResult;
};
