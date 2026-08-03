import { api } from "@/lib/api";
import type {
  DashboardActivity,
  DashboardAlert,
  DashboardMetrics,
  DashboardOverview,
  DashboardServiceHealth,
} from "@/types/dashboard";

export const dashboardService = {
  getOverview: async (): Promise<DashboardOverview> => {
    const response = await api.get<{ success: boolean; data: DashboardOverview }>(
      "/dashboard/overview"
    );
    return response.data.data;
  },

  getResources: async (): Promise<DashboardActivity[]> => {
    const response = await api.get<{
      success: boolean;
      data: DashboardActivity[];
    }>("/dashboard/resources");
    return response.data.data;
  },

  getAlerts: async (): Promise<DashboardAlert[]> => {
    const response = await api.get<{
      success: boolean;
      data: DashboardAlert[];
    }>("/dashboard/alerts");
    return response.data.data;
  },

  getClusters: async (): Promise<DashboardServiceHealth[]> => {
    const response = await api.get<{ success: boolean; data: DashboardServiceHealth[] }>(
      "/dashboard/clusters"
    );
    return response.data.data;
  },

  getMetrics: async (): Promise<DashboardMetrics> => {
    const response = await api.get<{ success: boolean; data: DashboardMetrics }>(
      "/dashboard/metrics"
    );
    return response.data.data;
  },

  // Backward-compatible service aliases used by existing dashboard hooks/components.
  getSummary: async (): Promise<DashboardOverview> => dashboardService.getOverview(),
  getStats: async (): Promise<DashboardMetrics> => dashboardService.getMetrics(),
  getActivity: async (): Promise<DashboardActivity[]> => dashboardService.getResources(),
  getServices: async (): Promise<DashboardServiceHealth[]> => dashboardService.getClusters(),
};
