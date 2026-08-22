import type { AlertJson, AlertResponse } from "@/types/alert-api";

/**
 * The condition metadata OpsPilot's alert-evaluation engine attaches to
 * every alert it creates (see backend internal/alerting.ConditionMetadata).
 * Manually-created or third-party-sourced alerts may not have any of this -
 * every field is optional, and callers must treat its absence as "not
 * available" rather than assume it's always present.
 */
export type AlertConditionMetadata = {
  condition?: string;
  currentValue?: number;
  threshold?: number;
  unit?: string;
  clusterId?: string;
  clusterName?: string;
  applicationId?: string;
};

export function parseAlertMetadata(metadata: AlertJson | undefined): AlertConditionMetadata {
  if (!metadata || typeof metadata !== "object") return {};

  const record = metadata as Record<string, unknown>;
  const asString = (key: string): string | undefined => (typeof record[key] === "string" && record[key] ? (record[key] as string) : undefined);
  const asNumber = (key: string): number | undefined => (typeof record[key] === "number" ? (record[key] as number) : undefined);

  return {
    condition: asString("condition"),
    currentValue: asNumber("currentValue"),
    threshold: asNumber("threshold"),
    unit: asString("unit"),
    clusterId: asString("clusterId"),
    clusterName: asString("clusterName"),
    applicationId: asString("applicationId"),
  };
}

export type ResourceIdentifier = {
  namespace?: string;
  name: string;
};

/**
 * Alert.resourceId is a plain string whose shape depends on resourceType:
 * POD/DEPLOYMENT conditions use "namespace/name" (see internal/alerting),
 * while CLUSTER/NODE conditions use a bare cluster ID or node name. Parsing
 * this is inherently best-effort for alerts not created by the engine.
 */
export function parseResourceIdentifier(resourceType: string, resourceId: string): ResourceIdentifier {
  const trimmed = resourceId.trim();
  if ((resourceType === "POD" || resourceType === "DEPLOYMENT") && trimmed.includes("/")) {
    const [namespace, ...rest] = trimmed.split("/");
    return { namespace, name: rest.join("/") };
  }

  return { name: trimmed };
}

/** Human-readable elapsed time between two ISO timestamps (e.g. "2h 14m"). */
export function formatDuration(startISO: string, endISO?: string): string {
  const start = new Date(startISO).getTime();
  const end = endISO ? new Date(endISO).getTime() : Date.now();
  const totalSeconds = Math.max(0, Math.floor((end - start) / 1000));

  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  if (minutes > 0) return `${minutes}m ${seconds}s`;
  return `${seconds}s`;
}

/** Two alerts are "related" only via a real shared identifier: the same resource. */
export function findRelatedAlerts(current: AlertResponse, candidates: AlertResponse[]): AlertResponse[] {
  return candidates.filter(
    (candidate) =>
      candidate.id !== current.id &&
      candidate.resourceType === current.resourceType &&
      candidate.resourceId === current.resourceId
  );
}
