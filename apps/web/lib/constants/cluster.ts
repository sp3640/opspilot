// Mirrors the backend's canonical cluster status values exactly
// (internal/constants/cluster.go) — status is server-owned, set only by
// cluster creation (PENDING_VALIDATION) and real connectivity validation
// (HEALTHY / INVALID), never by client input.
export const CLUSTER_STATUS_VALUES = [
  "PENDING_VALIDATION",
  "HEALTHY",
  "INVALID",
] as const;

export const CLUSTER_PROVIDER_VALUES = [
  "KUBERNETES",
  "DOCKER",
  "AZURE",
  "VM",
  "AWS",
  "CUSTOM",
] as const;

export const CLUSTER_CONNECTION_TYPE_VALUES = [
  "KUBECONFIG",
  "API",
  "SOCKET",
  "CUSTOM",
] as const;

export type ClusterStatus = (typeof CLUSTER_STATUS_VALUES)[number];
export type ClusterProvider = (typeof CLUSTER_PROVIDER_VALUES)[number];
export type ClusterConnectionType = (typeof CLUSTER_CONNECTION_TYPE_VALUES)[number];

export const CLUSTER_DEFAULT_PROVIDER: ClusterProvider = "KUBERNETES";
export const CLUSTER_DEFAULT_CONNECTION_TYPE: ClusterConnectionType = "KUBECONFIG";

export const CLUSTER_PROVIDER_LABELS: Record<ClusterProvider, string> = {
  KUBERNETES: "Kubernetes",
  DOCKER: "Docker",
  AZURE: "Azure",
  VM: "VM",
  AWS: "AWS",
  CUSTOM: "Custom",
};

// Friendly labels for the server-owned status values — "Connected" /
// "Disconnected" match how requirement #7 talks about connection status,
// while still being byte-compatible with the raw HEALTHY/INVALID/
// PENDING_VALIDATION strings the backend actually sends.
export const CLUSTER_STATUS_LABELS: Record<ClusterStatus, string> = {
  HEALTHY: "Connected",
  INVALID: "Disconnected",
  PENDING_VALIDATION: "Pending validation",
};

export const CLUSTER_SORT_FIELDS = {
  CREATED_AT: "created_at",
  UPDATED_AT: "updated_at",
  NAME: "name",
  PROVIDER: "provider",
  STATUS: "status",
  LAST_VALIDATED_AT: "last_validated_at",
  LAST_DISCOVERY_AT: "last_discovery_at",
} as const;
