import { api } from "@/lib/api";
import type {
  DashboardSummary,
  DashboardStats,
  DashboardActivity,
  ServiceHealth,
} from "@/types/dashboard";

export const dashboardService = {
  getSummary: async (): Promise<DashboardSummary> => {
    const response = await api.get<{ success: boolean; data: DashboardSummary }>(
      "/dashboard/summary"
    );
    return response.data.data;
  },

  getStats: async (): Promise<DashboardStats> => {
    const response = await api.get<{ success: boolean; data: DashboardStats }>(
      "/dashboard/stats"
    );
    return response.data.data;
  },

  getActivity: async (): Promise<DashboardActivity[]> => {
    const response = await api.get<{
      success: boolean;
      data: DashboardActivity[];
    }>("/dashboard/activity");
    return response.data.data;
  },

  getServices: async (): Promise<ServiceHealth[]> => {
    const response = await api.get<{ success: boolean; data: ServiceHealth[] }>(
      "/dashboard/recent-incidents"
    );
    return response.data.data;
  },
};
