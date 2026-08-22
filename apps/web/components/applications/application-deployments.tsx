"use client";

import { useMemo, useState } from "react";
import { Plus } from "lucide-react";

import { CreateDeploymentModal } from "@/components/deployments/create-deployment-modal";
import { DeploymentList } from "@/components/deployments/deployment-list";
import { Button } from "@/components/ui/button";
import { useDeploymentsByApplication } from "@/hooks/use-deployments";
import { PAGINATION_DEFAULT_PAGE_SIZE } from "@/lib/constants";
import { useHasPermission } from "@/store/auth-store";

export function ApplicationDeployments({ applicationId, projectId }: { applicationId: string; projectId: string }) {
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const canManageDeployments = useHasPermission("deployment:manage");

  const params = useMemo(() => ({ page: 1, limit: PAGINATION_DEFAULT_PAGE_SIZE }), []);
  const { data, error, isError, isLoading, refetch } = useDeploymentsByApplication(applicationId, params);

  const deployments = data?.items ?? [];

  return (
    <div className="space-y-4">
      {canManageDeployments ? (
        <div className="flex justify-end">
          <Button type="button" onClick={() => setCreateModalOpen(true)}>
            <Plus aria-hidden="true" className="h-4 w-4" />
            New deployment
          </Button>
        </div>
      ) : null}

      <DeploymentList
        deployments={deployments}
        isLoading={isLoading}
        isError={isError}
        error={error}
        onRetry={() => {
          void refetch();
        }}
        emptyDescription="Deployments created for this application will appear here."
      />

      <CreateDeploymentModal
        open={createModalOpen}
        applicationId={applicationId}
        projectId={projectId}
        onClose={() => setCreateModalOpen(false)}
      />
    </div>
  );
}
