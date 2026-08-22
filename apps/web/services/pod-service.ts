import { api } from "@/lib/api";
import type { PodListResponse, PodQueryParams } from "@/types/pod-api";
import type { EventListResponse } from "@/types/event-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const podService = {
  async listPodsByApplication(
    applicationId: string,
    params: PodQueryParams
  ): Promise<PodListResponse> {
    const response = await api.get<APIResponse<PodListResponse>>(
      `/applications/${applicationId}/pods`,
      { params }
    );
    return response.data.data;
  },

  async listPodsByCluster(clusterId: string, params: PodQueryParams): Promise<PodListResponse> {
    const response = await api.get<APIResponse<PodListResponse>>(`/clusters/${clusterId}/pods`, { params });
    return response.data.data;
  },

  async getPodEvents(namespace: string, name: string, applicationId: string): Promise<EventListResponse> {
    const response = await api.get<APIResponse<EventListResponse>>(
      `/pods/${namespace}/${name}/events`,
      { params: { applicationId } }
    );
    return response.data.data;
  },
};
