"use client";

import {
  NOT_AVAILABLE_APPLICATION_HEALTH_METRICS,
  type ApplicationHealthMetrics,
} from "@/types/application-health-api";

/**
 * Placeholder for the future Prometheus-backed metrics phase. There is no
 * `GET /applications/:id/metrics` endpoint yet, so this always resolves to
 * "not available" values shaped exactly like a real query result. When a
 * real endpoint exists, swap the body for a `useQuery` returning the same
 * `ApplicationHealthMetrics` shape — no consuming component needs to change.
 */
export function useApplicationHealthMetrics(_applicationId: string | null): {
  data: ApplicationHealthMetrics;
  isLoading: boolean;
  isError: boolean;
} {
  return { data: NOT_AVAILABLE_APPLICATION_HEALTH_METRICS, isLoading: false, isError: false };
}
