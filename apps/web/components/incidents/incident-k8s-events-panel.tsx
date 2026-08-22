"use client";

import axios from "axios";
import { Activity } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useLatestDeployment } from "@/hooks/use-deployments";
import { useEventsByApplication } from "@/hooks/use-events";
import type { IncidentResponse } from "@/types/incident-api";

/**
 * Incident -> Kubernetes Events: live cluster events for the affected
 * application's namespace. These are read straight from the cluster (no
 * database record), so only whatever the cluster's event TTL currently
 * retains is available - there is no historical log beyond that.
 */
export function IncidentK8sEventsPanel({ incident }: { incident: IncidentResponse }) {
  const hasApplication = Boolean(incident.applicationId);
  const { data: deployment, error: deploymentError, isLoading: isDeploymentLoading } = useLatestDeployment(
    incident.applicationId ?? null,
    hasApplication
  );
  const notDeployed = axios.isAxiosError(deploymentError) && deploymentError.response?.status === 404;
  const namespace = deployment?.namespace;
  const enabled = hasApplication && Boolean(namespace);

  const { data, error, isError, isLoading, refetch } = useEventsByApplication(
    incident.applicationId ?? null,
    { namespace: namespace ?? "" },
    enabled
  );

  if (!hasApplication) {
    return (
      <EmptyState
        icon={Activity}
        title="No events available"
        description="This incident has no affected application to fetch cluster events for."
      />
    );
  }

  if (isDeploymentLoading) {
    return <ListSkeleton rows={3} />;
  }

  if (notDeployed || !namespace) {
    return (
      <EmptyState
        icon={Activity}
        title="No events available"
        description="The affected application has no deployment on record, so there's no namespace to fetch events from."
      />
    );
  }

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load Kubernetes events."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={3} />;
  }

  const events = data?.items ?? [];

  if (events.length === 0) {
    return (
      <EmptyState
        icon={Activity}
        title="No recent events"
        description="No Kubernetes events are currently retained for this namespace."
      />
    );
  }

  return (
    <ul className="space-y-2">
      {events.map((event, index) => (
        <li key={`${event.involvedObject}-${event.reason}-${index}`} className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div className="flex items-center gap-2">
              <StatusBadge variant={event.type === "Warning" ? "warning" : "info"}>{event.type}</StatusBadge>
              <span className="text-sm font-medium">{event.reason}</span>
            </div>
            <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>
              {event.lastTimestamp ? new Date(event.lastTimestamp).toLocaleString() : event.age}
            </span>
          </div>
          <p className="mt-2 text-sm leading-6">{event.message}</p>
          <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
            {event.involvedObject} {event.count > 1 ? `· occurred ${event.count} times` : ""}
          </p>
        </li>
      ))}
    </ul>
  );
}
