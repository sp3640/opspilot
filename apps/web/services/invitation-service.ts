import { api } from "@/lib/api";
import type {
  AcceptInvitationRequest,
  InvitationListResponse,
  InvitationQueryParams,
  InvitationResponse,
  InviteRequest,
  ValidateInvitationResponse,
} from "@/types/invitation-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const invitationService = {
  async listInvitations(params: InvitationQueryParams): Promise<InvitationListResponse> {
    const response = await api.get<APIResponse<InvitationListResponse>>("/invitations", { params });
    return response.data.data;
  },

  async inviteUser(payload: InviteRequest): Promise<InvitationResponse> {
    const response = await api.post<APIResponse<InvitationResponse>>("/invitations", payload);
    return response.data.data;
  },

  async acceptInvitation(payload: AcceptInvitationRequest): Promise<InvitationResponse> {
    const response = await api.post<APIResponse<InvitationResponse>>("/invitations/accept", payload);
    return response.data.data;
  },

  async validateInvitation(token: string): Promise<ValidateInvitationResponse> {
    const response = await api.get<APIResponse<ValidateInvitationResponse>>("/invitations/validate", {
      params: { token },
    });
    return response.data.data;
  },

  async revokeInvitation(id: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/invitations/${id}`);
  },
};
