import { api } from "@/lib/api";
import type { ApplicationSLOResponse, ApplicationSREMetricsResponse, ConfigureApplicationSLORequest } from "@/types/sre-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const sreService = {
  async getSLO(applicationId: string): Promise<ApplicationSLOResponse> {
    const response = await api.get<APIResponse<ApplicationSLOResponse>>(`/applications/${applicationId}/slo`);
    return response.data.data;
  },

  async configureSLO(applicationId: string, payload: ConfigureApplicationSLORequest): Promise<ApplicationSLOResponse> {
    const response = await api.put<APIResponse<ApplicationSLOResponse>>(`/applications/${applicationId}/slo`, payload);
    return response.data.data;
  },

  async getMetrics(applicationId: string): Promise<ApplicationSREMetricsResponse> {
    const response = await api.get<APIResponse<ApplicationSREMetricsResponse>>(`/applications/${applicationId}/slo/metrics`);
    return response.data.data;
  },
};
