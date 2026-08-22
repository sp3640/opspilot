"use client";

import { AffectedResourceChain } from "@/components/shared/affected-resource-chain";
import type { IncidentResponse } from "@/types/incident-api";

/**
 * Incident -> Affected application -> Affected deployment -> Affected
 * cluster -> Affected namespace. Incidents (unlike alerts) have no
 * resourceType/resourceId of their own, so there is no pod/resource link
 * here - only what Incident.applicationId (a real, explicitly-set field)
 * actually supports.
 */
export function IncidentContextChain({ incident }: { incident: IncidentResponse }) {
  return <AffectedResourceChain applicationId={incident.applicationId ?? null} />;
}
