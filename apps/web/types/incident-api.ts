export type IncidentResponse = {
  id: number;
  title: string;
  description: string;
  severity: string;
  status: string;
  projectId: string;
  createdAt: string;
  updatedAt: string;
};

export type IncidentListResponse = {
  items: IncidentResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type CreateIncidentRequest = {
  title: string;
  description: string;
  severity: string;
  status: string;
  project_id: string;
};

export type UpdateIncidentRequest = {
  title: string;
  description: string;
  severity: string;
  status: string;
  project_id: string;
};

export type IncidentQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "title" | "severity" | "status" | "created_at" | "updated_at";
  order?: "asc" | "desc";
  projectId?: string;
  status?: string;
  severity?: string;
};
