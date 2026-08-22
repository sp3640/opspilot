"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { ingressService } from "@/services/ingress-service";
import type { IngressQueryParams } from "@/types/ingress-api";

const ingressKeys = {
  all: ["ingresses"] as const,
  list: (applicationId: string, params: IngressQueryParams) =>
    ["ingresses", "list", applicationId, params] as const,
  clusterList: (clusterId: string, params: IngressQueryParams) =>
    ["ingresses", "cluster-list", clusterId, params] as const,
};

export function useIngressesByApplication(
  applicationId: string | null,
  params: IngressQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: ingressKeys.list(applicationId ?? "", params),
    queryFn: () => ingressService.listIngressesByApplication(applicationId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

export function useIngressesByCluster(clusterId: string | null, params: IngressQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: ingressKeys.clusterList(clusterId ?? "", params),
    queryFn: () => ingressService.listIngressesByCluster(clusterId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(clusterId),
  });
}
