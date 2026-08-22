"use client";

import { Cpu } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { usePodEvents } from "@/hooks/use-pods";
import type { PodResponse } from "@/types/pod-api";

/**
 * Real Kubernetes events for this pod only (filtered server-side by
 * involvedObject.name/namespace). There is no synthetic fallback: if the
 * cluster has recorded no events for this pod, that is what is shown.
 */
export function PodEvents({ applicationId, pod }: { applicationId: string; pod: PodResponse }) {
  const { data, error, isError, isLoading, refetch } = usePodEvents(applicationId, pod.namespace, pod.name);

  const events = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load pod events. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (events.length === 0) {
    return (
      <EmptyState
        icon={Cpu}
        title="No events found"
        description="Kubernetes events for this pod will appear here as the cluster records them."
      />
    );
  }

  return (
    <div className="space-y-3">
      {events.map((event) => (
        <div
          key={`${event.namespace}/${event.name}`}
          className="rounded-2xl border p-4"
          style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
        >
          <div className="flex items-start justify-between gap-3">
            <p className="min-w-0 truncate font-semibold">{event.reason}</p>
            <StatusBadge variant={event.type === "Warning" ? "warning" : "info"}>{event.type}</StatusBadge>
          </div>
          <p className="mt-2 break-words text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>
            {event.message || "No message"}
          </p>
          <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Count</dt>
              <dd className="mt-1">{event.count}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Last seen</dt>
              <dd className="mt-1">{formatDateTime(event.lastTimestamp)}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Source</dt>
              <dd className="mt-1 break-all">{event.source || event.component || "-"}</dd>
            </div>
            <div>
              <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
              <dd className="mt-1">{event.age || "-"}</dd>
            </div>
          </dl>
        </div>
      ))}
    </div>
  );
}

function formatDateTime(value?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}
