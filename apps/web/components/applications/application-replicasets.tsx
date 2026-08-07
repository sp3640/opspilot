"use client";

import { useMemo, useState } from "react";
import { Layers3 } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useReplicaSetsByApplication } from "@/hooks/use-replicasets";

export function ApplicationReplicaSets({ applicationId }: { applicationId: string }) {
  const [selectedReplicaSetKey, setSelectedReplicaSetKey] = useState<string | null>(null);

  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = useReplicaSetsByApplication(applicationId, params);

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
        description="Replica sets for this application will appear here once it is running."
      />
    );
  }

  return (
    <div className="space-y-3">
      {replicaSets.map((replicaSet) => {
        const replicaSetKey = `${replicaSet.namespace}/${replicaSet.name}`;
        const isSelected = selectedReplicaSetKey === replicaSetKey;

        return (
          <button
            key={replicaSetKey}
            type="button"
            onClick={() => setSelectedReplicaSetKey(replicaSetKey)}
            aria-pressed={isSelected}
            className="w-full rounded-2xl border p-4 text-left transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
            style={{ borderColor: isSelected ? "var(--primary)" : "var(--border)", backgroundColor: "var(--card)" }}
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
          </button>
        );
      })}
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
