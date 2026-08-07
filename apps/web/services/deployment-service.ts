import { api } from "@/lib/api";
import type {
  DeploymentHistoryListResponse,
  DeploymentListResponse,
  DeploymentQueryParams,
  DeploymentResponse,
  RollbackDeploymentResponse,
} from "@/types/deployment-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const deploymentService = {
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
