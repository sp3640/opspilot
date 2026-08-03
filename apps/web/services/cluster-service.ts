import { api } from "@/lib/api";
import type {
  ClusterListResponse,
  ClusterQueryParams,
  ClusterResponse,
  ClusterValidationResponse,
  CreateClusterRequest,
  UpdateClusterRequest,
} from "@/types/cluster-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const clusterService = {
  async listClusters(params: ClusterQueryParams): Promise<ClusterListResponse> {
    const response = await api.get<APIResponse<ClusterListResponse>>("/clusters", { params });
    return response.data.data;
  },

  async getCluster(id: string): Promise<ClusterResponse> {
    const response = await api.get<APIResponse<ClusterResponse>>(`/clusters/${id}`);
    return response.data.data;
  },

  async createCluster(payload: CreateClusterRequest): Promise<ClusterResponse> {
    const response = await api.post<APIResponse<ClusterResponse>>("/clusters", payload);
    return response.data.data;
  },

  async updateCluster(id: string, payload: UpdateClusterRequest): Promise<ClusterResponse> {
    const response = await api.put<APIResponse<ClusterResponse>>(`/clusters/${id}`, payload);
    return response.data.data;
  },

  async deleteCluster(id: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/clusters/${id}`);
  },

  async validateCluster(id: string): Promise<ClusterValidationResponse> {
    const response = await api.post<APIResponse<ClusterValidationResponse>>(`/clusters/${id}/validate`);
    return response.data.data;
  },

  async setDefaultCluster(id: string): Promise<ClusterResponse> {
    const response = await api.post<APIResponse<ClusterResponse>>(`/clusters/${id}/default`);
    return response.data.data;
  },
};
