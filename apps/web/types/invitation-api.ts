export type InvitationResponse = {
  id: string;
  organizationId: string;
  email: string;
  role: string;
  token: string;
  status: string;
  invitedBy: number;
  expiresAt: string;
  acceptedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type InvitationListResponse = {
  items: InvitationResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type InviteRequest = {
  email: string;
  role: string;
};

export type AcceptInvitationRequest = {
  token: string;
};

export type ValidateInvitationResponse = {
  email: string;
  role: string;
  organizationId: string;
  organizationName: string;
  status: string;
  expiresAt: string;
};

export type InvitationQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "created_at" | "updated_at" | "email" | "expires_at" | "status";
  order?: "asc" | "desc";
};
