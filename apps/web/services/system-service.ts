import { api } from "@/lib/api";
import { env } from "@/lib/env";
import type { BackendHealthResponse } from "@/types/system";

const apiOrigin = env.API_URL.replace(/\/api\/v1\/?$/, "");

export const systemService = {
  async getHealth(): Promise<BackendHealthResponse> {
    const response = await api.get<BackendHealthResponse>(`${apiOrigin}/health`);
    return response.data;
  },
};
