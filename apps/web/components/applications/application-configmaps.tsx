"use client";

import { useMemo, useState } from "react";
import { ClipboardList } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useConfigMapsByApplication } from "@/hooks/use-configmaps";

export function ApplicationConfigMaps({ applicationId }: { applicationId: string }) {
  const [selectedConfigMapKey, setSelectedConfigMapKey] = useState<string | null>(null);

  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = useConfigMapsByApplication(applicationId, params);

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
        description="Config maps for this application will appear here once they are created."
      />
    );
  }

  return (
    <div className="space-y-3">
      {configMaps.map((configMap) => {
        const configMapKey = `${configMap.namespace}/${configMap.name}`;
        const isSelected = selectedConfigMapKey === configMapKey;

        return (
          <button
            key={configMapKey}
            type="button"
            onClick={() => setSelectedConfigMapKey(configMapKey)}
            aria-pressed={isSelected}
            className="w-full rounded-2xl border p-4 text-left transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
            style={{ borderColor: isSelected ? "var(--primary)" : "var(--border)", backgroundColor: "var(--card)" }}
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
          </button>
        );
      })}
    </div>
  );
}
