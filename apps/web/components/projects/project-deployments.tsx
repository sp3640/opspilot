"use client";

import { useMemo, useState } from "react";
import { Rocket } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { DeploymentDetailsDrawer } from "@/components/deployments/deployment-details-drawer";
import { useDeploymentsByProject } from "@/hooks/use-project-deployments";
import { PAGINATION_DEFAULT_PAGE_SIZE } from "@/lib/constants";
import type { DeploymentResponse } from "@/types/project-deployment-api";

export function ProjectDeployments({ projectId }: { projectId: string }) {
  const [selectedDeploymentID, setSelectedDeploymentID] = useState<string | null>(null);

  const params = useMemo(() => ({ page: 1, limit: PAGINATION_DEFAULT_PAGE_SIZE }), []);
  const { data, error, isError, isLoading, refetch } = useDeploymentsByProject(projectId, params);

  const deployments = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load deployments. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (deployments.length === 0) {
    return (
      <EmptyState
        icon={Rocket}
        title="No deployments yet"
        description="Deployments created for applications in this project will appear here."
      />
    );
  }

  const selectedDeployment = deployments.find((deployment) => deployment.id === selectedDeploymentID) ?? null;

  return (
    <div className="space-y-3">
      {deployments.map((deployment) => {
        const isSelected = selectedDeploymentID === deployment.id;

        return (
          <button
            key={deployment.id}
            type="button"
            onClick={() => setSelectedDeploymentID(deployment.id)}
            aria-pressed={isSelected}
            className="w-full rounded-2xl border p-4 text-left transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
            style={{ borderColor: isSelected ? "var(--primary)" : "var(--border)", backgroundColor: "var(--card)" }}
          >
            <div className="flex items-start justify-between gap-3">
              <p className="min-w-0 truncate font-semibold">{deploymentName(deployment)}</p>
              <StatusBadge variant={getStatusVariant(deployment.status)}>{deployment.status}</StatusBadge>
            </div>
            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Target cluster</dt>
                <dd className="mt-1 break-all">{deployment.targetClusterId}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Environment</dt>
                <dd className="mt-1">{deployment.environment || "-"}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Created</dt>
                <dd className="mt-1">{formatDate(deployment.createdAt)}</dd>
              </div>
            </dl>
          </button>
        );
      })}

      <DeploymentDetailsDrawer
        deployment={selectedDeployment}
        open={selectedDeploymentID !== null}
        onClose={() => setSelectedDeploymentID(null)}
      />
    </div>
  );
}

function deploymentName(deployment: DeploymentResponse): string {
  return deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image;
}

function getStatusVariant(status: string): "success" | "warning" | "critical" | "archived" | "info" {
  switch (status) {
    case "Succeeded":
      return "success";
    case "Failed":
      return "critical";
    case "Pending":
    case "Queued":
    case "RolledBack":
      return "warning";
    case "Cancelled":
      return "archived";
    default:
      return "info";
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
