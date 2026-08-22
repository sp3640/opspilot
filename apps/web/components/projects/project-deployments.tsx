"use client";

import { useMemo } from "react";

import { DeploymentList } from "@/components/deployments/deployment-list";
import { useDeploymentsByProject } from "@/hooks/use-project-deployments";
import { PAGINATION_DEFAULT_PAGE_SIZE } from "@/lib/constants";

export function ProjectDeployments({ projectId }: { projectId: string }) {
  const params = useMemo(() => ({ page: 1, limit: PAGINATION_DEFAULT_PAGE_SIZE }), []);
  const { data, error, isError, isLoading, refetch } = useDeploymentsByProject(projectId, params);

  const deployments = data?.items ?? [];

  return (
    <DeploymentList
      deployments={deployments}
      isLoading={isLoading}
      isError={isError}
      error={error}
      onRetry={() => {
        void refetch();
      }}
      emptyDescription="Deployments created for applications in this project will appear here."
    />
  );
}
