import { api } from "@/lib/api";
import type {
  DeploymentListResponse,
  DeploymentQueryParams,
} from "@/types/project-deployment-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const projectDeploymentService = {
  async listDeploymentsByProject(
    projectId: string,
    params: DeploymentQueryParams
  ): Promise<DeploymentListResponse> {
    const response = await api.get<APIResponse<DeploymentListResponse>>(
      `/projects/${projectId}/deployments`,
      { params }
    );
    return response.data.data;
  },
};
