"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { replicaSetService } from "@/services/replicaset-service";
import type { ReplicaSetQueryParams } from "@/types/replicaset-api";

const replicaSetKeys = {
  all: ["replicasets"] as const,
  list: (applicationId: string, params: ReplicaSetQueryParams) =>
    ["replicasets", "list", applicationId, params] as const,
  clusterList: (clusterId: string, params: ReplicaSetQueryParams) =>
    ["replicasets", "cluster-list", clusterId, params] as const,
};

export function useReplicaSetsByApplication(
  applicationId: string | null,
  params: ReplicaSetQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: replicaSetKeys.list(applicationId ?? "", params),
    queryFn: () => replicaSetService.listReplicaSetsByApplication(applicationId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

export function useReplicaSetsByCluster(clusterId: string | null, params: ReplicaSetQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: replicaSetKeys.clusterList(clusterId ?? "", params),
    queryFn: () => replicaSetService.listReplicaSetsByCluster(clusterId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(clusterId),
  });
}
