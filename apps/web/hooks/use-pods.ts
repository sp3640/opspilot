"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { podService } from "@/services/pod-service";
import type { PodQueryParams } from "@/types/pod-api";

const podKeys = {
  all: ["pods"] as const,
  list: (applicationId: string, params: PodQueryParams) =>
    ["pods", "list", applicationId, params] as const,
  clusterList: (clusterId: string, params: PodQueryParams) =>
    ["pods", "cluster-list", clusterId, params] as const,
};

export function usePodsByApplication(
  applicationId: string | null,
  params: PodQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: podKeys.list(applicationId ?? "", params),
    queryFn: () => podService.listPodsByApplication(applicationId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

export function usePodsByCluster(clusterId: string | null, params: PodQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: podKeys.clusterList(clusterId ?? "", params),
    queryFn: () => podService.listPodsByCluster(clusterId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(clusterId),
  });
}
