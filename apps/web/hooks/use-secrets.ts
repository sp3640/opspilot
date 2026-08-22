"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { secretService } from "@/services/secret-service";
import type { SecretQueryParams } from "@/types/secret-api";

const secretKeys = {
  all: ["secrets"] as const,
  list: (applicationId: string, params: SecretQueryParams) =>
    ["secrets", "list", applicationId, params] as const,
  clusterList: (clusterId: string, params: SecretQueryParams) =>
    ["secrets", "cluster-list", clusterId, params] as const,
};

export function useSecretsByApplication(
  applicationId: string | null,
  params: SecretQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: secretKeys.list(applicationId ?? "", params),
    queryFn: () => secretService.listSecretsByApplication(applicationId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

export function useSecretsByCluster(clusterId: string | null, params: SecretQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: secretKeys.clusterList(clusterId ?? "", params),
    queryFn: () => secretService.listSecretsByCluster(clusterId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(clusterId),
  });
}
