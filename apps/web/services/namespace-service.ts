import { api } from "@/lib/api";
import type { NamespaceListResponse } from "@/types/namespace-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const namespaceService = {
  async listNamespacesByCluster(clusterId: string): Promise<NamespaceListResponse> {
    const response = await api.get<APIResponse<NamespaceListResponse>>(`/clusters/${clusterId}/namespaces`);
    return response.data.data;
  },
};
