export type TeamResponse = {
  id: string;
  organizationId: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
};

export type TeamListResponse = {
  items: TeamResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type CreateTeamRequest = {
  name: string;
  description: string;
};

export type UpdateTeamRequest = {
  name: string;
  description: string;
};

export type TeamMemberResponse = {
  id: string;
  teamId: string;
  userId: number;
  name: string;
  email: string;
  createdAt: string;
};

export type TeamMemberListResponse = {
  items: TeamMemberResponse[];
  total: number;
};

export type AddTeamMemberRequest = {
  userId: number;
};

export type TeamQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "name" | "created_at" | "updated_at";
  order?: "asc" | "desc";
};
