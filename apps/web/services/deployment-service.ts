import { api } from "@/lib/api";
import type {
  CreateDeploymentRequest,
  DeploymentHistoryListResponse,
  DeploymentHistoryResponse,
  DeploymentListResponse,
  DeploymentQueryParams,
  DeploymentResponse,
  RollbackDeploymentResponse,
  UpdateDeploymentRequest,
} from "@/types/deployment-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const deploymentService = {
  async getDeployment(id: string): Promise<DeploymentResponse> {
    const response = await api.get<APIResponse<DeploymentResponse>>(`/deployments/${id}`);
    return response.data.data;
  },

  async createDeployment(payload: CreateDeploymentRequest): Promise<DeploymentResponse> {
    const response = await api.post<APIResponse<DeploymentResponse>>("/deployments", payload);
    return response.data.data;
  },

  async updateDeployment(id: string, payload: UpdateDeploymentRequest): Promise<DeploymentResponse> {
    const response = await api.patch<APIResponse<DeploymentResponse>>(`/deployments/${id}`, payload);
    return response.data.data;
  },

  async deleteDeployment(id: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/deployments/${id}`);
  },

  async getDeploymentHistoryRevision(deploymentId: string, revision: number): Promise<DeploymentHistoryResponse> {
    const response = await api.get<APIResponse<DeploymentHistoryResponse>>(
      `/deployments/${deploymentId}/history/${revision}`
    );
    return response.data.data;
  },

  async listDeploymentsByApplication(
    applicationId: string,
    params: DeploymentQueryParams
  ): Promise<DeploymentListResponse> {
    const response = await api.get<APIResponse<DeploymentListResponse>>(
      `/applications/${applicationId}/deployments`,
      { params }
    );
    return response.data.data;
  },

  async getLatestDeploymentByApplication(applicationId: string): Promise<DeploymentResponse> {
    const response = await api.get<APIResponse<DeploymentResponse>>(
      `/applications/${applicationId}/deployments/latest`
    );
    return response.data.data;
  },

  async getDeploymentHistory(deploymentId: string): Promise<DeploymentHistoryListResponse> {
    const response = await api.get<APIResponse<DeploymentHistoryListResponse>>(
      `/deployments/${deploymentId}/history`
    );
    return response.data.data;
  },

  async rollbackDeployment(id: string, revision: number): Promise<RollbackDeploymentResponse> {
    const response = await api.post<APIResponse<RollbackDeploymentResponse>>(
      `/deployments/${id}/rollback`,
      { revision }
    );
    return response.data.data;
  },

  async cancelDeployment(id: string): Promise<DeploymentResponse> {
    const response = await api.patch<APIResponse<DeploymentResponse>>(
      `/deployments/${id}/cancel`
    );
    return response.data.data;
  },
};
