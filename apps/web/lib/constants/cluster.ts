export const CLUSTER_STATUS_VALUES = [
  "CONNECTED",
  "DISCONNECTED",
  "FAILED",
  "PENDING",
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

export const CLUSTER_DEFAULT_STATUS: ClusterStatus = "PENDING";
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

export const CLUSTER_STATUS_LABELS: Record<ClusterStatus, string> = {
  CONNECTED: "Connected",
  DISCONNECTED: "Disconnected",
  FAILED: "Failed",
  PENDING: "Pending",
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
