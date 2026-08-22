import { api } from "@/lib/api";
import type { ClusterOrResourceMetricsParams, MetricAggregateResponse, MetricAggregationParams, MetricHistoryParams, MetricListResponse, MetricQueryParams, MetricResponse, MetricScopeParams, MetricsStatusResponse } from "@/types/metric-api";
type APIResponse<T> = { success: boolean; message: string; data: T };
export const metricService = {
  listMetrics: async (params: MetricQueryParams) => (await api.get<APIResponse<MetricListResponse>>("/metrics", { params })).data.data,
  getLatestMetrics: async (params: MetricScopeParams) => (await api.get<APIResponse<MetricResponse>>("/metrics/latest", { params })).data.data,
  getMetricHistory: async (params: MetricHistoryParams) => (await api.get<APIResponse<MetricResponse[]>>("/metrics/history", { params })).data.data,
  getMetricAggregation: async (params: MetricAggregationParams) => (await api.get<APIResponse<MetricAggregateResponse>>("/metrics/aggregate", { params })).data.data,
  // GET /clusters/:id/metrics and GET /resources/:id/metrics: the same
  // history data as /metrics/history, scoped by path rather than query param.
  getClusterMetrics: async (clusterId: string, params: ClusterOrResourceMetricsParams) =>
    (await api.get<APIResponse<MetricResponse[]>>(`/clusters/${clusterId}/metrics`, { params })).data.data,
  getResourceMetrics: async (resourceId: string, params: ClusterOrResourceMetricsParams) =>
    (await api.get<APIResponse<MetricResponse[]>>(`/resources/${resourceId}/metrics`, { params })).data.data,
  getClusterMetricsStatus: async (clusterId: string) =>
    (await api.get<APIResponse<MetricsStatusResponse>>(`/clusters/${clusterId}/metrics/status`)).data.data,
};
