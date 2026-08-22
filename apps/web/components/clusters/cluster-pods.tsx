"use client";

import { useMemo } from "react";
import { Boxes } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { usePodsByCluster } from "@/hooks/use-pods";
import { classifyPodHealth } from "@/lib/pod-health";

export function ClusterPods({ clusterId }: { clusterId: string }) {
  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = usePodsByCluster(clusterId, params);

  const pods = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load pods. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (pods.length === 0) {
    return (
      <EmptyState
        icon={Boxes}
        title="No pods found"
        description="Pods discovered from this cluster will appear here."
      />
    );
  }

  return (
    <div className="space-y-3">
      {pods.map((pod) => {
        const health = classifyPodHealth(pod);

        return (
          <div
            key={`${pod.namespace}/${pod.name}`}
            className="w-full rounded-2xl border p-4 text-left"
            style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
          >
            <div className="flex items-start justify-between gap-3">
              <p className="min-w-0 truncate font-semibold">{pod.name}</p>
              <StatusBadge variant={health.variant}>{health.state}</StatusBadge>
            </div>
            {health.reason ? (
              <p className="mt-1 text-xs" style={{ color: "var(--danger)" }}>{health.reason}</p>
            ) : null}
            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Namespace</dt>
                <dd className="mt-1 break-all">{pod.namespace}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Node</dt>
                <dd className="mt-1 break-all">{pod.nodeName || "-"}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Ready</dt>
                <dd className="mt-1">{pod.readyContainerCount}/{pod.containerCount}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Restarts</dt>
                <dd className="mt-1">{pod.restartCount}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
                <dd className="mt-1">{pod.age || "-"}</dd>
              </div>
            </dl>
          </div>
        );
      })}
    </div>
  );
}
