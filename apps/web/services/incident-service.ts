import { api } from "@/lib/api";
import type {
  AssignIncidentRequest,
  CreateIncidentRequest,
  IncidentListResponse,
  IncidentQueryParams,
  IncidentResponse,
  UpdateIncidentRequest,
} from "@/types/incident-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const incidentService = {
  async listIncidents(params: IncidentQueryParams): Promise<IncidentListResponse> {
    const response = await api.get<APIResponse<IncidentListResponse>>("/incidents", { params });
    return response.data.data;
  },

  async getIncident(id: number): Promise<IncidentResponse> {
    const response = await api.get<APIResponse<IncidentResponse>>(`/incidents/${id}`);
    return response.data.data;
  },

  async createIncident(payload: CreateIncidentRequest): Promise<IncidentResponse> {
    const response = await api.post<APIResponse<IncidentResponse>>("/incidents", payload);
    return response.data.data;
  },

  async updateIncident(id: number, payload: UpdateIncidentRequest): Promise<IncidentResponse> {
    const response = await api.put<APIResponse<IncidentResponse>>(`/incidents/${id}`, payload);
    return response.data.data;
  },

  async deleteIncident(id: number): Promise<void> {
    await api.delete<APIResponse<null>>(`/incidents/${id}`);
  },

  async assignIncident(id: number, payload: AssignIncidentRequest): Promise<IncidentResponse> {
    const response = await api.patch<APIResponse<IncidentResponse>>(`/incidents/${id}/assign`, payload);
    return response.data.data;
  },

  async acknowledgeIncident(id: number): Promise<IncidentResponse> {
    const response = await api.patch<APIResponse<IncidentResponse>>(`/incidents/${id}/acknowledge`);
    return response.data.data;
  },
};
