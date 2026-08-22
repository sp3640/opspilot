export type ClusterResponse = {
  id: string;
  organizationId: string;
  projectId: string;
  name: string;
  provider: string;
  status: string;
  isDefault: boolean;
  connectionType: string;
  credentialType: string;
  kubernetesVersion?: string;
  apiEndpoint: string;
  region: string;
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

// Status/version/validation fields are intentionally absent: they reflect
// real connectivity state and are only ever set by the server.
export type CreateClusterRequest = {
  project_id: string;
  name: string;
  provider: string;
  connection_type?: string;
  credential_type?: string;
  kubeconfig_encrypted: string;
  api_endpoint?: string;
  region?: string;
  metadata?: Record<string, unknown>;
};

// kubeconfig_encrypted is optional here on purpose: the backend never
// returns the stored credential, so the frontend has nothing to round-trip.
// Omit the key entirely to leave the stored credential untouched; only
// include it (non-empty) when the user is deliberately replacing it.
export type UpdateClusterRequest = {
  project_id: string;
  name: string;
  provider: string;
  connection_type?: string;
  credential_type?: string;
  kubeconfig_encrypted?: string;
  api_endpoint?: string;
  region?: string;
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
  status: string;
  kubernetesVersion?: string;
  apiServerUrl?: string;
  latencyMs: number;
  validatedAt: string;
  error?: string;
  cluster?: ClusterResponse;
};
