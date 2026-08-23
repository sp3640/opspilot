"use client";

import { useQuery } from "@tanstack/react-query";

import { incidentRCAService } from "@/services/incident-rca-service";

const incidentRCAKeys = {
  detail: (id: number) => ["incidents", "rca", id] as const,
};

/** Incident Intelligence / Root Cause Analysis for one incident. Read-only
 * and deterministic on the backend - refetch to recompute against the
 * latest evidence, never to poll for an AI job. */
export function useIncidentRCA(id: number | null) {
  return useQuery({
    queryKey: incidentRCAKeys.detail(id ?? 0),
    queryFn: () => incidentRCAService.getIncidentRCA(id ?? 0),
    enabled: Boolean(id),
  });
}
