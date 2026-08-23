export type ApplicationResponse = {
  id: string;
  organizationId: string;
  projectId: string;
  name: string;
  slug: string;
  description: string;
  repositoryUrl: string;
  defaultBranch: string;
  runtime: string;
  buildCommand: string;
  startCommand: string;
  port: number;
  environment: string;
  status: string;
  createdAt: string;
  updatedAt: string;
};

export type ApplicationListResponse = {
  items: ApplicationResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type CreateApplicationRequest = {
  name: string;
  slug?: string;
  description?: string;
  repository_url?: string;
  default_branch?: string;
  runtime: string;
  build_command?: string;
  start_command?: string;
  port: number;
  environment?: string;
  status?: string;
};

export type UpdateApplicationRequest = {
  name?: string;
  slug?: string;
  description?: string;
  repository_url?: string;
  default_branch?: string;
  runtime?: string;
  build_command?: string;
  start_command?: string;
  port?: number;
  environment?: string;
  status?: string;
};

export type ApplicationQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "name" | "slug" | "runtime" | "status" | "created_at" | "updated_at";
  order?: "asc" | "desc";
  projectId?: string;
};
