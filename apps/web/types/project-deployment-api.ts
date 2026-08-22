// GET /projects/:id/deployments returns the exact same dto.DeploymentListResponse
// shape as GET /applications/:id/deployments (both are built via toDeploymentListResponse
// on the backend), so the project-scoped deployment types are re-exported here rather
// than duplicated.
export type {
  DeploymentResponse,
  DeploymentListResponse,
  DeploymentQueryParams,
} from "@/types/deployment-api";
