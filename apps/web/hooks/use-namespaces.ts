"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { namespaceService } from "@/services/namespace-service";

const namespaceKeys = {
  all: ["namespaces"] as const,
  clusterList: (clusterId: string) => ["namespaces", "cluster-list", clusterId] as const,
};

export function useNamespacesByCluster(clusterId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: namespaceKeys.clusterList(clusterId ?? ""),
    queryFn: () => namespaceService.listNamespacesByCluster(clusterId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(clusterId),
  });
}
