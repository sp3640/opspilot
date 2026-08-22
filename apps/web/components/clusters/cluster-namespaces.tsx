"use client";

import { Boxes } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useNamespacesByCluster } from "@/hooks/use-namespaces";

export function ClusterNamespaces({ clusterId }: { clusterId: string }) {
  const { data, error, isError, isLoading, refetch } = useNamespacesByCluster(clusterId);

  const namespaces = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load namespaces. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (namespaces.length === 0) {
    return (
      <EmptyState
        icon={Boxes}
        title="No namespaces found"
        description="Namespaces discovered from this cluster will appear here."
      />
    );
  }

  return (
    <div className="space-y-3">
      {namespaces.map((namespace) => (
        <div
          key={namespace.name}
          className="w-full rounded-2xl border p-4 text-left"
          style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
        >
          <div className="flex items-start justify-between gap-3">
            <p className="min-w-0 truncate font-semibold">{namespace.name}</p>
            <StatusBadge variant={getStatusVariant(namespace.phase)}>{namespace.phase || "Unknown"}</StatusBadge>
          </div>
          <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
              <dd className="mt-1">{namespace.age || "-"}</dd>
            </div>
          </dl>
        </div>
      ))}
    </div>
  );
}

function getStatusVariant(phase: string): "success" | "warning" | "critical" | "info" {
  switch (phase) {
    case "Active":
      return "success";
    case "Terminating":
      return "warning";
    default:
      return "info";
  }
}
