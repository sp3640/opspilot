"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { nodeService } from "@/services/node-service";

const nodeKeys = {
  all: ["nodes"] as const,
  clusterList: (clusterId: string) => ["nodes", "cluster-list", clusterId] as const,
};

export function useNodesByCluster(clusterId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: nodeKeys.clusterList(clusterId ?? ""),
    queryFn: () => nodeService.listNodesByCluster(clusterId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(clusterId),
  });
}
