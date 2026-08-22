"use client";

import { Cpu } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useNodesByCluster } from "@/hooks/use-nodes";

export function ClusterNodes({ clusterId }: { clusterId: string }) {
  const { data, error, isError, isLoading, refetch } = useNodesByCluster(clusterId);

  const nodes = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load nodes. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (nodes.length === 0) {
    return (
      <EmptyState
        icon={Cpu}
        title="No nodes found"
        description="Nodes discovered from this cluster will appear here."
      />
    );
  }

  return (
    <div className="space-y-3">
      {nodes.map((node) => (
        <div
          key={node.name}
          className="w-full rounded-2xl border p-4 text-left"
          style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
        >
          <div className="flex items-start justify-between gap-3">
            <p className="min-w-0 truncate font-semibold">{node.name}</p>
            <StatusBadge variant={node.ready ? "success" : "critical"}>{node.ready ? "Ready" : "Not Ready"}</StatusBadge>
          </div>
          <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Kubelet version</dt>
              <dd className="mt-1 break-all">{node.kubeletVersion || "-"}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>OS / Arch</dt>
              <dd className="mt-1 break-all">{[node.operatingSystem, node.architecture].filter(Boolean).join(" / ") || "-"}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Internal IP</dt>
              <dd className="mt-1 break-all">{node.internalIP || "-"}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Schedulable</dt>
              <dd className="mt-1">{node.unschedulable ? "No" : "Yes"}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
              <dd className="mt-1">{node.age || "-"}</dd>
            </div>
          </dl>
        </div>
      ))}
    </div>
  );
}
