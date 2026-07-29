import { api } from "@/lib/api";
import type {
  CreateProjectRequest,
  ProjectListResponse,
  ProjectQueryParams,
  ProjectResponse,
  UpdateProjectRequest,
} from "@/types/project-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const projectService = {
  async listProjects(params: ProjectQueryParams): Promise<ProjectListResponse> {
    const response = await api.get<APIResponse<ProjectListResponse>>("/projects", { params });
    return response.data.data;
  },

  async getProject(id: string): Promise<ProjectResponse> {
    const response = await api.get<APIResponse<ProjectResponse>>(`/projects/${id}`);
    return response.data.data;
  },

  async createProject(payload: CreateProjectRequest): Promise<ProjectResponse> {
    const response = await api.post<APIResponse<ProjectResponse>>("/projects", payload);
    return response.data.data;
  },

  async updateProject(id: string, payload: UpdateProjectRequest): Promise<ProjectResponse> {
    const response = await api.put<APIResponse<ProjectResponse>>(`/projects/${id}`, payload);
    return response.data.data;
  },

  async deleteProject(id: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/projects/${id}`);
  },
};
