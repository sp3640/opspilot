import { api } from "@/lib/api";
import type { CreateResourceRequest, ResourceListResponse, ResourceQueryParams, ResourceResponse, SyncResourcesRequest, SyncResourcesResponse, UpdateResourceRequest } from "@/types/resource-api";

type APIResponse<T> = { success: boolean; message: string; data: T };

export const resourceService = {
  async listResources(params: ResourceQueryParams): Promise<ResourceListResponse> {
    const response = await api.get<APIResponse<ResourceListResponse>>("/resources", { params });
    return response.data.data;
  },
  async getResource(id: string): Promise<ResourceResponse> {
    const response = await api.get<APIResponse<ResourceResponse>>(`/resources/${id}`);
    return response.data.data;
  },
  async createResource(payload: CreateResourceRequest): Promise<ResourceResponse> {
    const response = await api.post<APIResponse<ResourceResponse>>("/resources", payload);
    return response.data.data;
  },
  async updateResource(id: string, payload: UpdateResourceRequest): Promise<ResourceResponse> {
    const response = await api.put<APIResponse<ResourceResponse>>(`/resources/${id}`, payload);
    return response.data.data;
  },
  async deleteResource(id: string): Promise<void> { await api.delete(`/resources/${id}`); },
  async syncResources(payload: SyncResourcesRequest): Promise<SyncResourcesResponse> {
    const response = await api.post<APIResponse<SyncResourcesResponse & { Created?: number; Updated?: number; Deleted?: number; Restored?: number }>>("/resources/sync", payload);
    const result = response.data.data;
    return {
      ...result,
      created: result.created ?? result.Created ?? 0,
      updated: result.updated ?? result.Updated ?? 0,
      deleted: result.deleted ?? result.Deleted ?? 0,
      restored: result.restored ?? result.Restored ?? 0,
    };
  },
};
