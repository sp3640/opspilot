import type { MetricResponse } from "@/types/metric-api";
export type MetricView = "cards" | "table" | "charts";
export type MetricFilters = { query: string; projectId: string; clusterId: string; resourceId: string; metricType: string; timeRange: "24h" | "7d" | "30d"; sort: "timestamp" | "metric_type" | "metric_name" | "resource_kind" | "value" | "created_at"; order: "asc" | "desc" };
export type Metric = MetricResponse;
