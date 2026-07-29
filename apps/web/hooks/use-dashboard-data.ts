import { useQuery } from "@tanstack/react-query";
import { dashboardService } from "@/services/dashboard-service";
import type {
  DashboardSummary,
  DashboardStats,
  DashboardActivity,
  ServiceHealth,
} from "@/types/dashboard";

const DASHBOARD_QUERY_KEYS = {
  summary: ["dashboard", "summary"] as const,
  stats: ["dashboard", "stats"] as const,
  activity: ["dashboard", "activity"] as const,
  services: ["dashboard", "services"] as const,
};

export const useDashboardSummary = () => {
  return useQuery<DashboardSummary>({
    queryKey: DASHBOARD_QUERY_KEYS.summary,
    queryFn: () => dashboardService.getSummary(),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 10,
  });
};

export const useDashboardStats = () => {
  return useQuery<DashboardStats>({
    queryKey: DASHBOARD_QUERY_KEYS.stats,
    queryFn: () => dashboardService.getStats(),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 10,
  });
};

export const useDashboardActivity = () => {
  return useQuery<DashboardActivity[]>({
    queryKey: DASHBOARD_QUERY_KEYS.activity,
    queryFn: () => dashboardService.getActivity(),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 10,
  });
};

export const useDashboardServices = () => {
  return useQuery<ServiceHealth[]>({
    queryKey: DASHBOARD_QUERY_KEYS.services,
    queryFn: () => dashboardService.getServices(),
    staleTime: 1000 * 60 * 5,
    gcTime: 1000 * 60 * 10,
  });
};
