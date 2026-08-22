"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/store/auth-store";

import { projectDeploymentService } from "@/services/project-deployment-service";
import type { DeploymentQueryParams } from "@/types/project-deployment-api";

// Exported so hooks/use-deployments.ts can invalidate this cache too: the
// Project Details → Deployments tab reuses DeploymentDetailsDrawer, whose
// Rollback/Cancel actions live in the shared deployment mutation hooks.
export const projectDeploymentKeys = {
  all: ["project-deployments"] as const,
  list: (projectId: string, params: DeploymentQueryParams) =>
    ["project-deployments", "list", projectId, params] as const,
};

export function useDeploymentsByProject(
  projectId: string | null,
  params: DeploymentQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: projectDeploymentKeys.list(projectId ?? "", params),
    queryFn: () => projectDeploymentService.listDeploymentsByProject(projectId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(projectId),
  });
}
