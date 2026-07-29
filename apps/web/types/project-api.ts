export type ProjectOwner = {
  id: number;
  name: string;
};

export type ProjectResponse = {
  id: string;
  name: string;
  slug: string;
  description: string;
  environment: string;
  health: string;
  members: number;
  services: number;
  owner: ProjectOwner;
  createdAt: string;
  updatedAt: string;
};

export type ProjectListResponse = {
  items: ProjectResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type CreateProjectRequest = {
  name: string;
  description: string;
};

export type UpdateProjectRequest = {
  name: string;
  description: string;
};

export type ProjectQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "name" | "created_at" | "updated_at";
  order?: "asc" | "desc";
};
