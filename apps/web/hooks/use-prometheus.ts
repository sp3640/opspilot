"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { prometheusService } from "@/services/prometheus-service";

// Keyed by integration + the exact submitted query text, so changing the
// query never serves a stale result from a different query's cache entry
// (React Query's default "no data while a new key loads" behavior applies -
// no keepPreviousData/placeholderData here on purpose). No polling: the
// query only ever runs when the caller explicitly enables it (the "Run
// query" action), never on an interval.
const prometheusKeys = {
  query: (integrationId: string, query: string) => ["prometheus", "query", integrationId, query] as const,
};

export function usePrometheusQuery(integrationId: string, query: string, enabled: boolean) {
  const accessToken = useAuthStore((state) => state.accessToken);
  const trimmedQuery = query.trim();

  return useQuery({
    queryKey: prometheusKeys.query(integrationId, trimmedQuery),
    queryFn: () => prometheusService.queryInstant(integrationId, trimmedQuery),
    enabled: enabled && Boolean(accessToken) && Boolean(integrationId) && Boolean(trimmedQuery),
    retry: false,
  });
}
