import { api } from "@/lib/api";
import type { PodListResponse, PodQueryParams } from "@/types/pod-api";

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
};
