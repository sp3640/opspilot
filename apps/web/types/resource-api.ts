export type JsonObject = Record<string, unknown>;

export type ResourceResponse = {
  id: string;
  projectId: string;
  parentResourceId?: string;
  kind: string;
  name: string;
  displayName: string;
  externalId: string;
  provider: string;
  region: string;
  namespace: string;
  cluster: string;
  status: string;
  health: string;
  labels: JsonObject;
  annotations: JsonObject;
  metadata: JsonObject;
  createdBy: number;
  createdAt: string;
  updatedAt: string;
};

export type ResourceListResponse = {
  items: ResourceResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type ResourcePayload = {
  project_id: string;
  parent_resource_id?: string;
  kind: string;
  name: string;
  display_name?: string;
  external_id?: string;
  provider?: string;
  region?: string;
  namespace?: string;
  cluster?: string;
  status: string;
  health: string;
  labels?: JsonObject;
  annotations?: JsonObject;
  metadata?: JsonObject;
};

export type CreateResourceRequest = ResourcePayload;
export type UpdateResourceRequest = ResourcePayload;

export type ResourceQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "created_at" | "updated_at" | "name" | "kind" | "status" | "health";
  order?: "asc" | "desc";
  projectId?: string;
  kind?: string;
  status?: string;
  health?: string;
  provider?: string;
  region?: string;
  namespace?: string;
  cluster?: string;
};

export type SyncResourcesRequest = { projectId: string; clusterId: string };
export type SyncResourcesResponse = { created: number; updated: number; deleted: number; restored: number; duration?: number; errors?: string[] };
