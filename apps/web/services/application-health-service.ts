import { api } from "@/lib/api";
import type { ApplicationHealthScoreResponse } from "@/types/application-health-score-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const applicationHealthService = {
  async getApplicationHealth(applicationId: string): Promise<ApplicationHealthScoreResponse> {
    const response = await api.get<APIResponse<ApplicationHealthScoreResponse>>(`/applications/${applicationId}/health`);
    return response.data.data;
  },
};
