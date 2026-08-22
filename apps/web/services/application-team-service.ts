import { api } from "@/lib/api";
import type {
  ApplicationTeamListResponse,
  ApplicationTeamResponse,
  AssignApplicationTeamRequest,
} from "@/types/application-team-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const applicationTeamService = {
  async listApplicationTeams(applicationId: string): Promise<ApplicationTeamListResponse> {
    const response = await api.get<APIResponse<ApplicationTeamListResponse>>(`/applications/${applicationId}/teams`);
    return response.data.data;
  },

  async assignTeam(applicationId: string, payload: AssignApplicationTeamRequest): Promise<ApplicationTeamResponse> {
    const response = await api.post<APIResponse<ApplicationTeamResponse>>(`/applications/${applicationId}/teams`, payload);
    return response.data.data;
  },

  async removeTeam(applicationId: string, teamId: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/applications/${applicationId}/teams/${teamId}`);
  },

  async listTeamApplications(teamId: string): Promise<ApplicationTeamListResponse> {
    const response = await api.get<APIResponse<ApplicationTeamListResponse>>(`/teams/${teamId}/applications`);
    return response.data.data;
  },
};
