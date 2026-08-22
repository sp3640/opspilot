"use client";

import { useMemo } from "react";
import { ClipboardList } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useConfigMapsByCluster } from "@/hooks/use-configmaps";

export function ClusterConfigMaps({ clusterId }: { clusterId: string }) {
  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = useConfigMapsByCluster(clusterId, params);

  const configMaps = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load config maps. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (configMaps.length === 0) {
    return (
      <EmptyState
        icon={ClipboardList}
        title="No config maps found"
        description="Config maps discovered from this cluster will appear here."
      />
    );
  }

  return (
    <div className="space-y-3">
      {configMaps.map((configMap) => (
        <div
          key={`${configMap.namespace}/${configMap.name}`}
          className="w-full rounded-2xl border p-4 text-left"
          style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
        >
          <p className="truncate font-semibold">{configMap.name}</p>
          <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Namespace</dt>
              <dd className="mt-1 break-all">{configMap.namespace}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Keys</dt>
              <dd className="mt-1">{configMap.dataKeyCount}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
              <dd className="mt-1">{configMap.age || "-"}</dd>
            </div>
          </dl>
        </div>
      ))}
    </div>
  );
}
