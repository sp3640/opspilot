export type OrganizationResponse = {
  id: string;
  name: string;
  slug: string;
  description: string;
  ownerId: number;
  createdAt: string;
  updatedAt: string;
  projectCount?: number;
  projectsCount?: number;
};

export type OrganizationListResponse = {
  items: OrganizationResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type CreateOrganizationRequest = {
  name: string;
  slug?: string;
  description: string;
};

export type UpdateOrganizationRequest = {
  name: string;
  slug?: string;
  description: string;
};

export type OrganizationQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "name" | "created_at" | "updated_at";
  order?: "asc" | "desc";
};
