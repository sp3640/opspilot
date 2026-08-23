import { api } from "@/lib/api";
import type {
  ApplicationListResponse,
  ApplicationQueryParams,
  ApplicationResponse,
  CreateApplicationRequest,
  UpdateApplicationRequest,
} from "@/types/application-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const applicationService = {
  async listApplications(params: ApplicationQueryParams): Promise<ApplicationListResponse> {
    const response = await api.get<APIResponse<ApplicationListResponse>>("/applications", { params });
    return response.data.data;
  },

  async listApplicationsByProject(
    projectId: string,
    params: ApplicationQueryParams
  ): Promise<ApplicationListResponse> {
    const response = await api.get<APIResponse<ApplicationListResponse>>(
      `/projects/${projectId}/applications`,
      { params }
    );
    return response.data.data;
  },

  async getApplication(applicationId: string): Promise<ApplicationResponse> {
    const response = await api.get<APIResponse<ApplicationResponse>>(`/applications/${applicationId}`);
    return response.data.data;
  },

  async createApplication(
    projectId: string,
    payload: CreateApplicationRequest
  ): Promise<ApplicationResponse> {
    const response = await api.post<APIResponse<ApplicationResponse>>(
      `/projects/${projectId}/applications`,
      payload
    );
    return response.data.data;
  },

  async updateApplication(
    applicationId: string,
    payload: UpdateApplicationRequest
  ): Promise<ApplicationResponse> {
    const response = await api.put<APIResponse<ApplicationResponse>>(
      `/applications/${applicationId}`,
      payload
    );
    return response.data.data;
  },

  async deleteApplication(applicationId: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/applications/${applicationId}`);
  },
};
