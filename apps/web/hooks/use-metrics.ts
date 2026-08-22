"use client";
import { useQuery } from "@tanstack/react-query";
import { metricService } from "@/services/metric-service";
import { useAuthStore } from "@/store/auth-store";
import type { ClusterOrResourceMetricsParams, MetricAggregationParams, MetricHistoryParams, MetricQueryParams, MetricScopeParams } from "@/types/metric-api";
const keys = {
  list: (p: MetricQueryParams) => ["metrics", "list", p] as const,
  latest: (p: MetricScopeParams) => ["metrics", "latest", p] as const,
  history: (p: MetricHistoryParams) => ["metrics", "history", p] as const,
  aggregate: (p: MetricAggregationParams) => ["metrics", "aggregate", p] as const,
  cluster: (clusterId: string, p: ClusterOrResourceMetricsParams) => ["metrics", "cluster", clusterId, p] as const,
  resource: (resourceId: string, p: ClusterOrResourceMetricsParams) => ["metrics", "resource", resourceId, p] as const,
  clusterStatus: (clusterId: string) => ["metrics", "cluster-status", clusterId] as const,
};
export function useMetrics(params: MetricQueryParams) { const token = useAuthStore((s) => s.accessToken); return useQuery({ queryKey: keys.list(params), queryFn: () => metricService.listMetrics(params), enabled: Boolean(token) }); }
export function useLatestMetrics(params: MetricScopeParams | null) { return useQuery({ queryKey: keys.latest(params ?? { projectId: "", metricType: "" }), queryFn: () => metricService.getLatestMetrics(params!), enabled: Boolean(params?.projectId && params.metricType) }); }
export function useMetricHistory(params: MetricHistoryParams | null) { return useQuery({ queryKey: keys.history(params ?? { projectId: "", metricType: "" }), queryFn: () => metricService.getMetricHistory(params!), enabled: Boolean(params?.projectId && params.metricType) }); }
export function useMetricAggregation(params: MetricAggregationParams | null) { return useQuery({ queryKey: keys.aggregate(params ?? { projectId: "", metricType: "", start: "", end: "" }), queryFn: () => metricService.getMetricAggregation(params!), enabled: Boolean(params?.projectId && params.metricType && params.start && params.end) }); }

export function useClusterMetrics(clusterId: string | null, params: ClusterOrResourceMetricsParams | null) {
  const token = useAuthStore((s) => s.accessToken);
  return useQuery({
    queryKey: keys.cluster(clusterId ?? "", params ?? { projectId: "", metricType: "" }),
    queryFn: () => metricService.getClusterMetrics(clusterId ?? "", params!),
    enabled: Boolean(token) && Boolean(clusterId) && Boolean(params?.projectId) && Boolean(params?.metricType),
  });
}

export function useResourceMetrics(resourceId: string | null, params: ClusterOrResourceMetricsParams | null) {
  const token = useAuthStore((s) => s.accessToken);
  return useQuery({
    queryKey: keys.resource(resourceId ?? "", params ?? { projectId: "", metricType: "" }),
    queryFn: () => metricService.getResourceMetrics(resourceId ?? "", params!),
    enabled: Boolean(token) && Boolean(resourceId) && Boolean(params?.projectId) && Boolean(params?.metricType),
  });
}

export function useClusterMetricsStatus(clusterId: string | null) {
  const token = useAuthStore((s) => s.accessToken);
  return useQuery({
    queryKey: keys.clusterStatus(clusterId ?? ""),
    queryFn: () => metricService.getClusterMetricsStatus(clusterId ?? ""),
    enabled: Boolean(token) && Boolean(clusterId),
  });
}
