"use client";
import { useQuery } from "@tanstack/react-query";
import { metricService } from "@/services/metric-service";
import { useAuthStore } from "@/store/auth-store";
import type { MetricAggregationParams, MetricHistoryParams, MetricQueryParams, MetricScopeParams } from "@/types/metric-api";
const keys = { list: (p: MetricQueryParams) => ["metrics", "list", p] as const, latest: (p: MetricScopeParams) => ["metrics", "latest", p] as const, history: (p: MetricHistoryParams) => ["metrics", "history", p] as const, aggregate: (p: MetricAggregationParams) => ["metrics", "aggregate", p] as const };
export function useMetrics(params: MetricQueryParams) { const token = useAuthStore((s) => s.accessToken); return useQuery({ queryKey: keys.list(params), queryFn: () => metricService.listMetrics(params), enabled: Boolean(token) }); }
export function useLatestMetrics(params: MetricScopeParams | null) { return useQuery({ queryKey: keys.latest(params ?? { projectId: "", metricType: "" }), queryFn: () => metricService.getLatestMetrics(params!), enabled: Boolean(params?.projectId && params.metricType) }); }
export function useMetricHistory(params: MetricHistoryParams | null) { return useQuery({ queryKey: keys.history(params ?? { projectId: "", metricType: "" }), queryFn: () => metricService.getMetricHistory(params!), enabled: Boolean(params?.projectId && params.metricType) }); }
export function useMetricAggregation(params: MetricAggregationParams | null) { return useQuery({ queryKey: keys.aggregate(params ?? { projectId: "", metricType: "", start: "", end: "" }), queryFn: () => metricService.getMetricAggregation(params!), enabled: Boolean(params?.projectId && params.metricType && params.start && params.end) }); }
