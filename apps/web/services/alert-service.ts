import { api } from "@/lib/api";
import type { AlertListResponse, AlertQueryParams, AlertResponse, CreateAlertRequest, UpdateAlertRequest } from "@/types/alert-api";
type APIResponse<T> = { success: boolean; message: string; data: T };
export const alertService = {
 listAlerts: async (params: AlertQueryParams) => (await api.get<APIResponse<AlertListResponse>>("/alerts", { params })).data.data,
 getAlert: async (id: number) => (await api.get<APIResponse<AlertResponse>>(`/alerts/${id}`)).data.data,
 createAlert: async (payload: CreateAlertRequest) => (await api.post<APIResponse<AlertResponse>>("/alerts", payload)).data.data,
 updateAlert: async (id: number, payload: UpdateAlertRequest) => (await api.put<APIResponse<AlertResponse>>(`/alerts/${id}`, payload)).data.data,
 deleteAlert: async (id: number) => { await api.delete(`/alerts/${id}`); },
 acknowledgeAlert: async (id: number) => (await api.post<APIResponse<AlertResponse>>(`/alerts/${id}/acknowledge`)).data.data,
 resolveAlert: async (id: number) => (await api.post<APIResponse<AlertResponse>>(`/alerts/${id}/resolve`)).data.data,
 reopenAlert: async (id: number) => (await api.post<APIResponse<AlertResponse>>(`/alerts/${id}/reopen`)).data.data,
 attachIncident: async (id: number, incidentId: number) => (await api.post<APIResponse<AlertResponse>>(`/alerts/${id}/incident`, { incidentId })).data.data,
};
