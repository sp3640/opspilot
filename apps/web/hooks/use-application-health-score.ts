"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { applicationHealthService } from "@/services/application-health-service";

const applicationHealthKeys = {
  detail: (applicationId: string) => ["application-health", applicationId] as const,
};

export function useApplicationHealthScore(applicationId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: applicationHealthKeys.detail(applicationId ?? ""),
    queryFn: () => applicationHealthService.getApplicationHealth(applicationId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}
