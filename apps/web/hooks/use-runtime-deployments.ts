"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { runtimeDeploymentService } from "@/services/runtime-deployment-service";
import type { RuntimeDeploymentQueryParams } from "@/types/runtime-deployment-api";

const runtimeDeploymentKeys = {
  all: ["runtime-deployments"] as const,
  list: (applicationId: string, params: RuntimeDeploymentQueryParams) =>
    ["runtime-deployments", "list", applicationId, params] as const,
  clusterList: (clusterId: string, params: RuntimeDeploymentQueryParams) =>
    ["runtime-deployments", "cluster-list", clusterId, params] as const,
};

export function useRuntimeDeploymentsByApplication(
  applicationId: string | null,
  params: RuntimeDeploymentQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: runtimeDeploymentKeys.list(applicationId ?? "", params),
    queryFn: () => runtimeDeploymentService.listRuntimeDeploymentsByApplication(applicationId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

export function useRuntimeDeploymentsByCluster(
  clusterId: string | null,
  params: RuntimeDeploymentQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: runtimeDeploymentKeys.clusterList(clusterId ?? "", params),
    queryFn: () => runtimeDeploymentService.listRuntimeDeploymentsByCluster(clusterId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(clusterId),
  });
}
