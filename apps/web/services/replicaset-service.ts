import { api } from "@/lib/api";
import type { ReplicaSetListResponse, ReplicaSetQueryParams } from "@/types/replicaset-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const replicaSetService = {
  async listReplicaSetsByApplication(
    applicationId: string,
    params: ReplicaSetQueryParams
  ): Promise<ReplicaSetListResponse> {
    const response = await api.get<APIResponse<ReplicaSetListResponse>>(
      `/applications/${applicationId}/replicasets`,
      { params }
    );
    return response.data.data;
  },

  async listReplicaSetsByCluster(clusterId: string, params: ReplicaSetQueryParams): Promise<ReplicaSetListResponse> {
    const response = await api.get<APIResponse<ReplicaSetListResponse>>(`/clusters/${clusterId}/replicasets`, { params });
    return response.data.data;
  },
};
