"use client";

import { useMemo } from "react";
import { Layers3 } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useReplicaSetsByCluster } from "@/hooks/use-replicasets";

export function ClusterReplicaSets({ clusterId }: { clusterId: string }) {
  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = useReplicaSetsByCluster(clusterId, params);

  const replicaSets = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load replica sets. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (replicaSets.length === 0) {
    return (
      <EmptyState
        icon={Layers3}
        title="No replica sets found"
        description="Replica sets discovered from this cluster will appear here."
      />
    );
  }

  return (
    <div className="space-y-3">
      {replicaSets.map((replicaSet) => (
        <div
          key={`${replicaSet.namespace}/${replicaSet.name}`}
          className="w-full rounded-2xl border p-4 text-left"
          style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
        >
          <div className="flex items-start justify-between gap-3">
            <p className="min-w-0 truncate font-semibold">{replicaSet.name}</p>
            <StatusBadge variant={getStatusVariant(replicaSet.status)}>{replicaSet.status}</StatusBadge>
          </div>
          <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Namespace</dt>
              <dd className="mt-1 break-all">{replicaSet.namespace}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Desired</dt>
              <dd className="mt-1">{replicaSet.desiredReplicas}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Ready</dt>
              <dd className="mt-1">{replicaSet.readyReplicas}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Available</dt>
              <dd className="mt-1">{replicaSet.availableReplicas}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
              <dd className="mt-1">{replicaSet.age || "-"}</dd>
            </div>
          </dl>
        </div>
      ))}
    </div>
  );
}

function getStatusVariant(status: string): "success" | "warning" | "critical" | "archived" | "info" {
  switch (status) {
    case "Ready":
      return "success";
    case "Progressing":
      return "warning";
    case "Unavailable":
      return "critical";
    case "ScaledDown":
      return "archived";
    default:
      return "info";
  }
}
