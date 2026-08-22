import { api } from "@/lib/api";
import type {
  AddTeamMemberRequest,
  CreateTeamRequest,
  TeamListResponse,
  TeamMemberListResponse,
  TeamMemberResponse,
  TeamQueryParams,
  TeamResponse,
  UpdateTeamRequest,
} from "@/types/team-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const teamService = {
  async listTeams(params: TeamQueryParams): Promise<TeamListResponse> {
    const response = await api.get<APIResponse<TeamListResponse>>("/teams", { params });
    return response.data.data;
  },

  async getTeam(id: string): Promise<TeamResponse> {
    const response = await api.get<APIResponse<TeamResponse>>(`/teams/${id}`);
    return response.data.data;
  },

  async createTeam(payload: CreateTeamRequest): Promise<TeamResponse> {
    const response = await api.post<APIResponse<TeamResponse>>("/teams", payload);
    return response.data.data;
  },

  async updateTeam(id: string, payload: UpdateTeamRequest): Promise<TeamResponse> {
    const response = await api.put<APIResponse<TeamResponse>>(`/teams/${id}`, payload);
    return response.data.data;
  },

  async deleteTeam(id: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/teams/${id}`);
  },

  async listTeamMembers(teamId: string): Promise<TeamMemberListResponse> {
    const response = await api.get<APIResponse<TeamMemberListResponse>>(`/teams/${teamId}/members`);
    return response.data.data;
  },

  async addTeamMember(teamId: string, payload: AddTeamMemberRequest): Promise<TeamMemberResponse> {
    const response = await api.post<APIResponse<TeamMemberResponse>>(`/teams/${teamId}/members`, payload);
    return response.data.data;
  },

  async removeTeamMember(teamId: string, userId: number): Promise<void> {
    await api.delete<APIResponse<null>>(`/teams/${teamId}/members/${userId}`);
  },
};
