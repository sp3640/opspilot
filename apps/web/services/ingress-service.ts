import { api } from "@/lib/api";
import type { IngressListResponse, IngressQueryParams } from "@/types/ingress-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const ingressService = {
  async listIngressesByApplication(
    applicationId: string,
    params: IngressQueryParams
  ): Promise<IngressListResponse> {
    const response = await api.get<APIResponse<IngressListResponse>>(
      `/applications/${applicationId}/ingresses`,
      { params }
    );
    return response.data.data;
  },

  async listIngressesByCluster(clusterId: string, params: IngressQueryParams): Promise<IngressListResponse> {
    const response = await api.get<APIResponse<IngressListResponse>>(`/clusters/${clusterId}/ingresses`, { params });
    return response.data.data;
  },
};
