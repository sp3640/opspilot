"use client";

import { useMemo, useState } from "react";
import { Cpu } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useEventsByApplication } from "@/hooks/use-events";

export function ApplicationEvents({ applicationId }: { applicationId: string }) {
  const [selectedEventKey, setSelectedEventKey] = useState<string | null>(null);

  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = useEventsByApplication(applicationId, params);

  const events = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load events. Please try again."}
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
        description="Events for this application will appear here as they are recorded."
      />
    );
  }

  return (
    <div className="space-y-3">
      {events.map((event) => {
        const eventKey = `${event.namespace}/${event.name}`;
        const isSelected = selectedEventKey === eventKey;

        return (
          <button
            key={eventKey}
            type="button"
            onClick={() => setSelectedEventKey(eventKey)}
            aria-pressed={isSelected}
            className="w-full rounded-2xl border p-4 text-left transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
            style={{ borderColor: isSelected ? "var(--primary)" : "var(--border)", backgroundColor: "var(--card)" }}
          >
            <div className="flex items-start justify-between gap-3">
              <p className="min-w-0 truncate font-semibold">{event.reason}</p>
              <StatusBadge variant={getEventTypeVariant(event.type)}>{event.type}</StatusBadge>
            </div>
            <p className="mt-2 break-words text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>
              {event.message || "No message"}
            </p>
            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Object</dt>
                <dd className="mt-1 break-all">{event.involvedObject || "-"}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Count</dt>
                <dd className="mt-1">{event.count}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Last seen</dt>
                <dd className="mt-1">{formatDateTime(event.lastTimestamp)}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
                <dd className="mt-1">{event.age || "-"}</dd>
              </div>
            </dl>
          </button>
        );
      })}
    </div>
  );
}

function getEventTypeVariant(type: string): "warning" | "info" {
  return type === "Warning" ? "warning" : "info";
}

function formatDateTime(value?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}
