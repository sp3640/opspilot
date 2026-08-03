import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "@/services/dashboard-service";
import type {
  DashboardActivity,
  DashboardAlert,
  DashboardMetrics,
  DashboardOverview,
  DashboardServiceHealth,
} from "@/types/dashboard";

const DASHBOARD_QUERY_KEYS = {
  overview: ["dashboard", "overview"] as const,
  resources: ["dashboard", "resources"] as const,
  alerts: ["dashboard", "alerts"] as const,
  clusters: ["dashboard", "clusters"] as const,
  metrics: ["dashboard", "metrics"] as const,
};

export const useDashboardOverview = () => {
  return useQuery<DashboardOverview>({
    queryKey: DASHBOARD_QUERY_KEYS.overview,
    queryFn: () => dashboardService.getOverview(),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 10,
  });
};

export const useDashboardResources = () => {
  return useQuery<DashboardActivity[]>({
    queryKey: DASHBOARD_QUERY_KEYS.resources,
    queryFn: () => dashboardService.getResources(),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 10,
  });
};

export const useDashboardAlerts = () => {
  return useQuery<DashboardAlert[]>({
    queryKey: DASHBOARD_QUERY_KEYS.alerts,
    queryFn: () => dashboardService.getAlerts(),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 10,
  });
};

export const useDashboardClusters = () => {
  return useQuery<DashboardServiceHealth[]>({
    queryKey: DASHBOARD_QUERY_KEYS.clusters,
    queryFn: () => dashboardService.getClusters(),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 10,
  });
};

export const useDashboardMetrics = () => {
  return useQuery<DashboardMetrics>({
    queryKey: DASHBOARD_QUERY_KEYS.metrics,
    queryFn: () => dashboardService.getMetrics(),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 10,
  });
};

// Backward-compatible hook aliases used by existing dashboard components.
export const useDashboardSummary = useDashboardOverview;
export const useDashboardStats = useDashboardMetrics;
export const useDashboardActivity = useDashboardResources;
export const useDashboardServices = useDashboardClusters;
