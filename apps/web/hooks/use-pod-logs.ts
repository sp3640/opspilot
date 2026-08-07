"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { podLogService } from "@/services/pod-log-service";
import type { PodLogQueryParams } from "@/types/pod-log-api";

const podLogKeys = {
  all: ["pod-logs"] as const,
  detail: (namespace: string, name: string, params: PodLogQueryParams) =>
    ["pod-logs", namespace, name, params] as const,
};

export function usePodLogs(
  namespace: string | null,
  name: string | null,
  params: PodLogQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: podLogKeys.detail(namespace ?? "", name ?? "", params),
    queryFn: () => podLogService.getPodLogs(namespace ?? "", name ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(namespace) && Boolean(name) && Boolean(params.applicationId),
  });
}
