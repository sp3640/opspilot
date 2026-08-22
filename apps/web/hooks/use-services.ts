"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { serviceService } from "@/services/service-service";
import type { ServiceQueryParams } from "@/types/service-api";

const serviceKeys = {
  all: ["services"] as const,
  list: (applicationId: string, params: ServiceQueryParams) =>
    ["services", "list", applicationId, params] as const,
  clusterList: (clusterId: string, params: ServiceQueryParams) =>
    ["services", "cluster-list", clusterId, params] as const,
};

export function useServicesByApplication(
  applicationId: string | null,
  params: ServiceQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: serviceKeys.list(applicationId ?? "", params),
    queryFn: () => serviceService.listServicesByApplication(applicationId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

export function useServicesByCluster(clusterId: string | null, params: ServiceQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: serviceKeys.clusterList(clusterId ?? "", params),
    queryFn: () => serviceService.listServicesByCluster(clusterId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(clusterId),
  });
}
