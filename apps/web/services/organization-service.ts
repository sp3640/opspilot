import { api } from "@/lib/api";
import type {
  CreateOrganizationRequest,
  OrganizationListResponse,
  OrganizationQueryParams,
  OrganizationResponse,
  UpdateOrganizationRequest,
} from "@/types/organization-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const organizationService = {
  async listOrganizations(params: OrganizationQueryParams): Promise<OrganizationListResponse> {
    const response = await api.get<APIResponse<OrganizationListResponse>>("/organizations", {
      params,
    });
    return response.data.data;
  },

  async getOrganization(id: string): Promise<OrganizationResponse> {
    const response = await api.get<APIResponse<OrganizationResponse>>(`/organizations/${id}`);
    return response.data.data;
  },

  async createOrganization(payload: CreateOrganizationRequest): Promise<OrganizationResponse> {
    const response = await api.post<APIResponse<OrganizationResponse>>("/organizations", payload);
    return response.data.data;
  },

  async updateOrganization(id: string, payload: UpdateOrganizationRequest): Promise<OrganizationResponse> {
    const response = await api.put<APIResponse<OrganizationResponse>>(`/organizations/${id}`, payload);
    return response.data.data;
  },

  async deleteOrganization(id: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/organizations/${id}`);
  },
};
