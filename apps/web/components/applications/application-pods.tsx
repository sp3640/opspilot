"use client";

import { useMemo, useState } from "react";
import { Boxes } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { PodDetailsDrawer } from "@/components/pods/pod-details-drawer";
import { usePodsByApplication } from "@/hooks/use-pods";

export function ApplicationPods({ applicationId }: { applicationId: string }) {
  const [selectedPodKey, setSelectedPodKey] = useState<string | null>(null);

  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = usePodsByApplication(applicationId, params);

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
        description="Pods for this application will appear here once it is running."
      />
    );
  }

  const selectedPod = pods.find((pod) => `${pod.namespace}/${pod.name}` === selectedPodKey) ?? null;

  return (
    <div className="space-y-3">
      {pods.map((pod) => {
        const podKey = `${pod.namespace}/${pod.name}`;
        const isSelected = selectedPodKey === podKey;

        return (
          <button
            key={podKey}
            type="button"
            onClick={() => setSelectedPodKey(podKey)}
            aria-pressed={isSelected}
            className="w-full rounded-2xl border p-4 text-left transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
            style={{ borderColor: isSelected ? "var(--primary)" : "var(--border)", backgroundColor: "var(--card)" }}
          >
            <div className="flex items-start justify-between gap-3">
              <p className="min-w-0 truncate font-semibold">{pod.name}</p>
              <StatusBadge variant={getStatusVariant(pod.phase)}>{pod.phase}</StatusBadge>
            </div>
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
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Restarts</dt>
                <dd className="mt-1">{pod.restartCount}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
                <dd className="mt-1">{pod.age || "-"}</dd>
              </div>
            </dl>
          </button>
        );
      })}

      <PodDetailsDrawer
        pod={selectedPod}
        applicationId={applicationId}
        open={selectedPodKey !== null}
        onClose={() => setSelectedPodKey(null)}
      />
    </div>
  );
}

function getStatusVariant(phase: string): "success" | "warning" | "critical" | "info" {
  switch (phase) {
    case "Running":
    case "Succeeded":
      return "success";
    case "Pending":
      return "warning";
    case "Failed":
      return "critical";
    default:
      return "info";
  }
}
