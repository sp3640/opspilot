"use client";

import { useMemo, useState } from "react";
import { AlertTriangle } from "lucide-react";

import { AlertDetailsDrawer } from "@/components/alerts/alert-details-drawer";
import { AlertSeverityBadge } from "@/components/alerts/alert-severity-badge";
import { AlertStatusBadge } from "@/components/alerts/alert-status-badge";
import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useAlerts } from "@/hooks/use-alerts";
import { useProjects } from "@/hooks/use-projects";
import { PROJECT_LOOKUP_QUERY } from "@/lib/constants";
import type { IncidentResponse } from "@/types/incident-api";

const LOOKUP_LIMIT = 100;

/** Incident -> Alerts: every alert explicitly attached to this incident (GET /alerts?incidentId=). */
export function IncidentAlerts({ incident }: { incident: IncidentResponse }) {
  const [selectedAlertId, setSelectedAlertId] = useState<number | null>(null);
  const params = useMemo(() => ({ page: 1, limit: LOOKUP_LIMIT, incidentId: incident.id }), [incident.id]);
  const { data, error, isError, isLoading, refetch } = useAlerts(params);
  const { data: projectsData } = useProjects(PROJECT_LOOKUP_QUERY);
  const alerts = data?.items ?? [];
  const projects = useMemo(() => projectsData?.items ?? [], [projectsData?.items]);

  const projectNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const project of projects) map.set(project.id, project.name);
    return map;
  }, [projects]);

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load attached alerts."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={3} />;
  }

  if (alerts.length === 0) {
    return (
      <EmptyState
        icon={AlertTriangle}
        title="No attached alerts"
        description="No alerts have been linked to this incident yet."
      />
    );
  }

  return (
    <>
      <ul className="space-y-2">
        {alerts.map((alert) => (
          <li key={alert.id} className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
            <button
              type="button"
              onClick={() => setSelectedAlertId(alert.id)}
              className="flex w-full items-start justify-between gap-3 text-left"
            >
              <div className="min-w-0">
                <p className="truncate text-sm font-medium underline-offset-2 hover:underline" style={{ color: "var(--primary)" }}>
                  {alert.title}
                </p>
                <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  {alert.resourceType}: {alert.resourceId}
                </p>
              </div>
              <div className="flex shrink-0 gap-1.5">
                <AlertSeverityBadge severity={alert.severity} />
                <AlertStatusBadge status={alert.status} />
              </div>
            </button>
          </li>
        ))}
      </ul>

      <AlertDetailsDrawer
        alertID={selectedAlertId}
        projects={projects}
        projectNameById={projectNameById}
        onClose={() => setSelectedAlertId(null)}
      />
    </>
  );
}
