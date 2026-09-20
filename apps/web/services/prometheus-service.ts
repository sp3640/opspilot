import { api } from "@/lib/api";
import type { PrometheusQueryResponse } from "@/types/prometheus-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const prometheusService = {
  async queryInstant(integrationId: string, query: string): Promise<PrometheusQueryResponse> {
    const response = await api.get<APIResponse<PrometheusQueryResponse>>(
      `/integrations/${integrationId}/prometheus/query`,
      { params: { query } }
    );
    return response.data.data;
  },
};
