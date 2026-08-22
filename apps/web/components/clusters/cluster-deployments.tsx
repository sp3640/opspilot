"use client";

import { useMemo } from "react";
import { Server } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useRuntimeDeploymentsByCluster } from "@/hooks/use-runtime-deployments";

export function ClusterDeployments({ clusterId }: { clusterId: string }) {
  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = useRuntimeDeploymentsByCluster(clusterId, params);

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
        icon={Server}
        title="No deployments found"
        description="Deployments discovered from this cluster will appear here."
      />
    );
  }

  return (
    <div className="space-y-3">
      {deployments.map((deployment) => (
        <div
          key={`${deployment.namespace}/${deployment.name}`}
          className="w-full rounded-2xl border p-4 text-left"
          style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
        >
          <div className="flex items-start justify-between gap-3">
            <p className="min-w-0 truncate font-semibold">{deployment.name}</p>
            <StatusBadge variant={getStatusVariant(deployment.status)}>{deployment.status}</StatusBadge>
          </div>
          <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Namespace</dt>
              <dd className="mt-1 break-all">{deployment.namespace}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Ready</dt>
              <dd className="mt-1">{deployment.readyReplicas}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Available</dt>
              <dd className="mt-1">{deployment.availableReplicas}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
              <dd className="mt-1">{deployment.age || "-"}</dd>
            </div>
          </dl>
        </div>
      ))}
    </div>
  );
}

function getStatusVariant(status: string): "success" | "warning" | "critical" | "info" {
  switch (status) {
    case "Healthy":
      return "success";
    case "Progressing":
    case "Scaling":
      return "warning";
    case "Unavailable":
    case "Failed":
      return "critical";
    default:
      return "info";
  }
}
