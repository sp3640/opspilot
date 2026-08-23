import { api } from "@/lib/api";
import type { IncidentRCAResponse } from "@/types/incident-rca-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const incidentRCAService = {
  async getIncidentRCA(id: number): Promise<IncidentRCAResponse> {
    const response = await api.get<APIResponse<IncidentRCAResponse>>(`/incidents/${id}/rca`);
    return response.data.data;
  },
};
