export type ClusterResponse = {
  id: string;
  projectId: string;
  name: string;
  provider: string;
  status: string;
  isDefault: boolean;
  connectionType: string;
  kubeconfigEncrypted: string;
  apiEndpoint: string;
  region: string;
  version: string;
  lastValidatedAt?: string;
  lastDiscoveryAt?: string;
  createdBy: number;
  validationError: string;
  metadata: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
};

export type ClusterListResponse = {
  items: ClusterResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type CreateClusterRequest = {
  project_id: string;
  name: string;
  provider: string;
  status: string;
  connection_type?: string;
  kubeconfig_encrypted?: string;
  api_endpoint?: string;
  region?: string;
  version?: string;
  last_validated_at?: string;
  last_discovery_at?: string;
  validation_error?: string;
  metadata?: Record<string, unknown>;
};

export type UpdateClusterRequest = {
  project_id: string;
  name: string;
  provider: string;
  status: string;
  connection_type?: string;
  kubeconfig_encrypted?: string;
  api_endpoint?: string;
  region?: string;
  version?: string;
  last_validated_at?: string;
  last_discovery_at?: string;
  validation_error?: string;
  metadata?: Record<string, unknown>;
};

export type ClusterQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?:
    | "created_at"
    | "updated_at"
    | "name"
    | "provider"
    | "status"
    | "last_validated_at"
    | "last_discovery_at";
  order?: "asc" | "desc";
  projectId?: string;
  provider?: string;
  status?: string;
};

export type ClusterValidationResponse = {
  connected: boolean;
  clusterVersion?: string;
  apiServerUrl?: string;
  latencyMs: number;
  validatedAt: string;
  error?: string;
  cluster?: ClusterResponse;
};
