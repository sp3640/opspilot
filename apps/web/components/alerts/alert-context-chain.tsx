"use client";

import { useMemo } from "react";

import { AffectedResourceChain } from "@/components/shared/affected-resource-chain";
import { parseAlertMetadata } from "@/lib/alert-correlation";
import type { AlertResponse } from "@/types/alert-api";

/**
 * Alert -> Affected application -> Affected deployment -> Affected cluster
 * -> Affected namespace -> Affected pod/resource. The chain itself is shared
 * with incidents (see components/shared/affected-resource-chain.tsx); this
 * wrapper only extracts the real identifiers an alert carries (condition
 * metadata set by the alert-evaluation engine, see backend
 * internal/alerting) - a cluster-wide condition has no application/
 * deployment/pod, and the chain honestly stops at Cluster rather than
 * fabricating the rest.
 */
export function AlertContextChain({ alert }: { alert: AlertResponse }) {
  const meta = useMemo(() => parseAlertMetadata(alert.metadata), [alert.metadata]);

  return (
    <AffectedResourceChain
      applicationId={meta.applicationId ?? null}
      clusterIdFallback={meta.clusterId}
      clusterNameFallback={meta.clusterName}
      resourceType={alert.resourceType}
      resourceId={alert.resourceId}
    />
  );
}
