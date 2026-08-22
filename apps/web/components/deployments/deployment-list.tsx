"use client";

import { useState } from "react";
import { Rocket } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useCluster } from "@/hooks/use-clusters";
import { formatDuration } from "@/lib/alert-correlation";
import type { DeploymentResponse } from "@/types/deployment-api";

import { DeploymentDetailsDrawer } from "./deployment-details-drawer";

/**
 * Shared deployment list card, used by both Application -> Deployments and
 * Project -> Deployments (identical `DeploymentResponse` shape on both
 * endpoints) so cluster-name resolution, duration, and commit/author
 * previews stay consistent between the two entry points rather than
 * drifting across two copies.
 */
export function DeploymentList({
  deployments,
  isLoading,
  isError,
  error,
  onRetry,
  emptyDescription,
}: {
  deployments: DeploymentResponse[];
  isLoading: boolean;
  isError: boolean;
  error: unknown;
  onRetry: () => void;
  emptyDescription: string;
}) {
  const [selectedDeploymentID, setSelectedDeploymentID] = useState<string | null>(null);

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load deployments. Please try again."}
        onRetry={onRetry}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (deployments.length === 0) {
    return <EmptyState icon={Rocket} title="No deployments yet" description={emptyDescription} />;
  }

  const selectedDeployment = deployments.find((deployment) => deployment.id === selectedDeploymentID) ?? null;

  return (
    <div className="space-y-3">
      {deployments.map((deployment) => (
        <DeploymentListRow
          key={deployment.id}
          deployment={deployment}
          isSelected={selectedDeploymentID === deployment.id}
          onSelect={() => setSelectedDeploymentID(deployment.id)}
        />
      ))}

      <DeploymentDetailsDrawer
        deployment={selectedDeployment}
        open={selectedDeploymentID !== null}
        onClose={() => setSelectedDeploymentID(null)}
      />
    </div>
  );
}

function DeploymentListRow({
  deployment,
  isSelected,
  onSelect,
}: {
  deployment: DeploymentResponse;
  isSelected: boolean;
  onSelect: () => void;
}) {
  const { data: cluster } = useCluster(deployment.targetClusterId);

  return (
    <button
      type="button"
      onClick={onSelect}
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
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Cluster</dt>
          <dd className="mt-1 break-all">{cluster?.name ?? deployment.targetClusterId}</dd>
        </div>
        <div>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Environment</dt>
          <dd className="mt-1">{deployment.environment || "-"}</dd>
        </div>
        <div>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Created</dt>
          <dd className="mt-1">{formatDate(deployment.createdAt)}</dd>
        </div>
        <div>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Duration</dt>
          <dd className="mt-1">{deployment.startedAt ? formatDuration(deployment.startedAt, deployment.completedAt) : "-"}</dd>
        </div>
        {deployment.commitSha || deployment.author ? (
          <div className="col-span-2">
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Commit / Author</dt>
            <dd className="mt-1 break-all">
              {deployment.commitSha ?? "-"}
              {deployment.author ? ` · ${deployment.author}` : ""}
            </dd>
          </div>
        ) : null}
      </dl>
    </button>
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
