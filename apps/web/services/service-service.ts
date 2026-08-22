import { api } from "@/lib/api";
import type { ServiceListResponse, ServiceQueryParams } from "@/types/service-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const serviceService = {
  async listServicesByApplication(
    applicationId: string,
    params: ServiceQueryParams
  ): Promise<ServiceListResponse> {
    const response = await api.get<APIResponse<ServiceListResponse>>(
      `/applications/${applicationId}/services`,
      { params }
    );
    return response.data.data;
  },

  async listServicesByCluster(clusterId: string, params: ServiceQueryParams): Promise<ServiceListResponse> {
    const response = await api.get<APIResponse<ServiceListResponse>>(`/clusters/${clusterId}/services`, { params });
    return response.data.data;
  },
};
