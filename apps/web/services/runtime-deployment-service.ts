import { api } from "@/lib/api";
import type {
  RuntimeDeploymentListResponse,
  RuntimeDeploymentQueryParams,
} from "@/types/runtime-deployment-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const runtimeDeploymentService = {
  async listRuntimeDeploymentsByApplication(
    applicationId: string,
    params: RuntimeDeploymentQueryParams
  ): Promise<RuntimeDeploymentListResponse> {
    const response = await api.get<APIResponse<RuntimeDeploymentListResponse>>(
      `/applications/${applicationId}/runtime/deployments`,
      { params }
    );
    return response.data.data;
  },
};
