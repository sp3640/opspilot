export type DeploymentResponse = {
  id: string;
  applicationId: string;
  projectId: string;
  organizationId: string;
  image: string;
  imageTag: string;
  environment: string;
  namespace: string;
  replicaCount: number;
  status: string;
  deploymentStrategy: string;
  targetClusterId: string;
  commitSha?: string;
  author?: string;
  createdBy: number;
  updatedBy: number;
  startedAt?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateDeploymentRequest = {
  applicationId: string;
  projectId: string;
  targetClusterId: string;
  image: string;
  imageTag?: string;
  environment: string;
  namespace: string;
  replicaCount: number;
  deploymentStrategy: string;
  commitSha?: string;
  author?: string;
};

export type UpdateDeploymentRequest = {
  targetClusterId?: string;
  image?: string;
  imageTag?: string;
  environment?: string;
  namespace?: string;
  replicaCount?: number;
  deploymentStrategy?: string;
  commitSha?: string;
  author?: string;
};

export type DeploymentListResponse = {
  items: DeploymentResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type DeploymentQueryParams = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: "created_at" | "updated_at" | "started_at" | "completed_at" | "status";
  order?: "asc" | "desc";
  projectId?: string;
  applicationId?: string;
};

export type DeploymentHistoryResponse = {
  id: string;
  deploymentId: string;
  applicationId: string;
  projectId: string;
  organizationId: string;
  revision: number;
  image: string;
  imageTag: string;
  environment: string;
  namespace: string;
  replicaCount: number;
  deploymentStrategy: string;
  status: string;
  commitSha?: string;
  author?: string;
  changeSummary: string;
  triggeredBy: number;
  createdAt: string;
};

export type DeploymentHistoryListResponse = {
  items: DeploymentHistoryResponse[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type RollbackDeploymentResponse = {
  deployment: DeploymentResponse;
  currentRevision: number;
  rollbackSourceRevision: number;
};
