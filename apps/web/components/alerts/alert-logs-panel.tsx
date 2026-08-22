"use client";

import { useMemo } from "react";
import { FileText } from "lucide-react";

import { EmptyState, ListSkeleton } from "@/components/common";
import { PodLogs } from "@/components/pods/pod-logs";
import { usePodsByApplication } from "@/hooks/use-pods";
import { parseAlertMetadata, parseResourceIdentifier } from "@/lib/alert-correlation";
import type { AlertResponse } from "@/types/alert-api";

/**
 * Relevant logs, reusing the Live Kubernetes Pod Logs viewer (Phase 11).
 * Only available for pod-scoped alerts where the pod can actually be
 * resolved (known application + namespace, and the pod still exists) -
 * every other case is an honest "not available", not a fabricated log
 * stream.
 */
export function AlertLogsPanel({ alert }: { alert: AlertResponse }) {
  const meta = useMemo(() => parseAlertMetadata(alert.metadata), [alert.metadata]);
  const resource = useMemo(() => parseResourceIdentifier(alert.resourceType, alert.resourceId), [alert.resourceType, alert.resourceId]);

  const enabled = alert.resourceType === "POD" && Boolean(meta.applicationId) && Boolean(resource.namespace);
  const podsParams = useMemo(() => ({ namespace: resource.namespace ?? "" }), [resource.namespace]);
  const { data, isLoading } = usePodsByApplication(meta.applicationId ?? null, podsParams, enabled);

  if (!enabled) {
    return (
      <EmptyState
        icon={FileText}
        title="Logs not available"
        description="Logs are only available for pod-scoped alerts linked to a known application and namespace."
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  const pod = (data?.items ?? []).find((item) => item.name === resource.name) ?? null;
  if (!pod) {
    return (
      <EmptyState
        icon={FileText}
        title="Pod not found"
        description="This pod could not be found - it may have been deleted or rescheduled since the alert fired."
      />
    );
  }

  return <PodLogs applicationId={meta.applicationId ?? ""} pod={pod} />;
}
