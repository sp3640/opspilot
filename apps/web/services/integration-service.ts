import { api } from "@/lib/api";
import type {
  CreateIntegrationRequest,
  IntegrationCheckResponse,
  IntegrationListResponse,
  IntegrationResponse,
  UpdateIntegrationRequest,
} from "@/types/integration-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const integrationService = {
  async listIntegrations(): Promise<IntegrationListResponse> {
    const response = await api.get<APIResponse<IntegrationListResponse>>("/integrations", {
      params: { limit: 100 },
    });
    return response.data.data;
  },

  async getIntegration(id: string): Promise<IntegrationResponse> {
    const response = await api.get<APIResponse<IntegrationResponse>>(`/integrations/${id}`);
    return response.data.data;
  },

  async createIntegration(payload: CreateIntegrationRequest): Promise<IntegrationResponse> {
    const response = await api.post<APIResponse<IntegrationResponse>>("/integrations", payload);
    return response.data.data;
  },

  async updateIntegration(id: string, payload: UpdateIntegrationRequest): Promise<IntegrationResponse> {
    const response = await api.put<APIResponse<IntegrationResponse>>(`/integrations/${id}`, payload);
    return response.data.data;
  },

  async deleteIntegration(id: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/integrations/${id}`);
  },

  async testIntegration(id: string): Promise<IntegrationCheckResponse> {
    const response = await api.post<APIResponse<IntegrationCheckResponse>>(`/integrations/${id}/test`);
    return response.data.data;
  },

  async checkIntegration(id: string): Promise<IntegrationCheckResponse> {
    const response = await api.post<APIResponse<IntegrationCheckResponse>>(`/integrations/${id}/check`);
    return response.data.data;
  },
};
