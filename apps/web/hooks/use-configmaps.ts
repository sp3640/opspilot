"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { configMapService } from "@/services/configmap-service";
import type { ConfigMapQueryParams } from "@/types/configmap-api";

const configMapKeys = {
  all: ["configmaps"] as const,
  list: (applicationId: string, params: ConfigMapQueryParams) =>
    ["configmaps", "list", applicationId, params] as const,
  clusterList: (clusterId: string, params: ConfigMapQueryParams) =>
    ["configmaps", "cluster-list", clusterId, params] as const,
};

export function useConfigMapsByApplication(
  applicationId: string | null,
  params: ConfigMapQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: configMapKeys.list(applicationId ?? "", params),
    queryFn: () => configMapService.listConfigMapsByApplication(applicationId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

export function useConfigMapsByCluster(clusterId: string | null, params: ConfigMapQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: configMapKeys.clusterList(clusterId ?? "", params),
    queryFn: () => configMapService.listConfigMapsByCluster(clusterId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(clusterId),
  });
}
