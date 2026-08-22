import { api } from "@/lib/api";
import type {
  AssignTeamRequest,
  ProjectTeamListResponse,
  ProjectTeamResponse,
} from "@/types/project-team-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const projectTeamService = {
  async listProjectTeams(projectId: string): Promise<ProjectTeamListResponse> {
    const response = await api.get<APIResponse<ProjectTeamListResponse>>(`/projects/${projectId}/teams`);
    return response.data.data;
  },

  async assignTeam(projectId: string, payload: AssignTeamRequest): Promise<ProjectTeamResponse> {
    const response = await api.post<APIResponse<ProjectTeamResponse>>(`/projects/${projectId}/teams`, payload);
    return response.data.data;
  },

  async removeTeam(projectId: string, teamId: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/projects/${projectId}/teams/${teamId}`);
  },

  async listTeamProjects(teamId: string): Promise<ProjectTeamListResponse> {
    const response = await api.get<APIResponse<ProjectTeamListResponse>>(`/teams/${teamId}/projects`);
    return response.data.data;
  },
};
