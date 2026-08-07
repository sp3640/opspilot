import { api } from "@/lib/api";
import type { PodLogQueryParams, PodLogResponse } from "@/types/pod-log-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const podLogService = {
  async getPodLogs(
    namespace: string,
    name: string,
    params: PodLogQueryParams
  ): Promise<PodLogResponse> {
    const response = await api.get<APIResponse<PodLogResponse>>(
      `/pods/${namespace}/${name}/logs`,
      { params }
    );
    return response.data.data;
  },
};
