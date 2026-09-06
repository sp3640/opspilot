// Integration type identifiers (Sprint 27: Integration Foundation). No real
// connector exists yet for any of these - see the backend's
// internal/connector package - so every integration created today reports
// connectorImplemented=false until a future sprint wires one up.
export type IntegrationType =
  | "github"
  | "slack"
  | "email"
  | "prometheus"
  | "loki"
  | "opentelemetry"
  | "azure"
  | "teams"
  | "webhook";

export const INTEGRATION_TYPES: IntegrationType[] = [
  "github",
  "slack",
  "email",
  "prometheus",
  "loki",
  "opentelemetry",
  "azure",
  "teams",
  "webhook",
];

export const INTEGRATION_TYPE_LABELS: Record<IntegrationType, string> = {
  github: "GitHub",
  slack: "Slack",
  email: "Email",
  prometheus: "Prometheus",
  loki: "Loki",
  opentelemetry: "OpenTelemetry",
  azure: "Azure",
  teams: "Microsoft Teams",
  webhook: "Webhook",
};

// The integration record's own connect/disconnect lifecycle state - never
// the health of the external service itself.
export type IntegrationStatus = "PENDING" | "CONNECTED" | "DISCONNECTED" | "ERROR";

export type IntegrationResponse = {
  id: string;
  organizationId: string;
  type: IntegrationType;
  name: string;
  status: IntegrationStatus;
  metadata?: Record<string, string>;
  hasCredentials: boolean;
  connectorImplemented: boolean;
  connectorDescription?: string;
  lastCheckedAt?: string;
  lastError?: string;
  createdAt: string;
  updatedAt: string;
};

export type IntegrationListResponse = {
  items: IntegrationResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type IntegrationCheckResponse = {
  success: boolean;
  message: string;
  checkedAt: string;
  integration: IntegrationResponse;
};

// Credentials are write-only: accepted on create/update, never returned by
// any response type above.
export type CreateIntegrationRequest = {
  type: IntegrationType;
  name: string;
  metadata?: Record<string, string>;
  credentials?: Record<string, string>;
};

export type UpdateIntegrationRequest = {
  name: string;
  metadata?: Record<string, string>;
  credentials?: Record<string, string>;
};
