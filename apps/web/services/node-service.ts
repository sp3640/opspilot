import { api } from "@/lib/api";
import type { NodeListResponse } from "@/types/node-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const nodeService = {
  async listNodesByCluster(clusterId: string): Promise<NodeListResponse> {
    const response = await api.get<APIResponse<NodeListResponse>>(`/clusters/${clusterId}/nodes`);
    return response.data.data;
  },
};
