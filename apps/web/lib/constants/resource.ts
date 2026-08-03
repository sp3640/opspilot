export const RESOURCE_KIND_VALUES = ["Cluster", "Namespace", "Node", "Deployment", "ReplicaSet", "Pod", "Container", "Service", "Ingress", "ConfigMap", "Secret", "PersistentVolume", "PersistentVolumeClaim", "Database", "VM", "Application", "Queue", "Storage", "Cache", "LoadBalancer"] as const;
export const RESOURCE_STATUS_VALUES = ["ACTIVE", "PENDING", "UPDATING", "TERMINATING", "DELETED", "UNKNOWN"] as const;
export const RESOURCE_HEALTH_VALUES = ["HEALTHY", "DEGRADED", "UNHEALTHY", "UNKNOWN"] as const;
export const RESOURCE_SORT_FIELDS = { CREATED_AT: "created_at", UPDATED_AT: "updated_at", NAME: "name", KIND: "kind", STATUS: "status", HEALTH: "health" } as const;
