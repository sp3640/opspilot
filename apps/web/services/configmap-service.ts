import { api } from "@/lib/api";
import type { ConfigMapListResponse, ConfigMapQueryParams } from "@/types/configmap-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const configMapService = {
  async listConfigMapsByApplication(
    applicationId: string,
    params: ConfigMapQueryParams
  ): Promise<ConfigMapListResponse> {
    const response = await api.get<APIResponse<ConfigMapListResponse>>(
      `/applications/${applicationId}/configmaps`,
      { params }
    );
    return response.data.data;
  },
};
