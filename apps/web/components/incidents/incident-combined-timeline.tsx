"use client";

import { useMemo, useState } from "react";
import { useQueries } from "@tanstack/react-query";
import {
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  History,
  MessageSquare,
  RefreshCw,
  Rocket,
  Undo2,
  type LucideIcon,
} from "lucide-react";

import { AlertDetailsDrawer } from "@/components/alerts/alert-details-drawer";
import { DeploymentDetailsDrawer } from "@/components/deployments/deployment-details-drawer";
import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useAlerts } from "@/hooks/use-alerts";
import { useComments } from "@/hooks/use-comments";
import { useDeploymentsByApplication, useLatestDeployment } from "@/hooks/use-deployments";
import { usePodsByApplication } from "@/hooks/use-pods";
import { useProjects } from "@/hooks/use-projects";
import { PROJECT_LOOKUP_QUERY } from "@/lib/constants";
import { buildIncidentTimeline, type TimelineEvent, type TimelineEventType } from "@/lib/incident-timeline";
import { deploymentService } from "@/services/deployment-service";
import type { DeploymentResponse } from "@/types/deployment-api";
import type { IncidentResponse } from "@/types/incident-api";

const DEPLOYMENT_LOOKUP = { page: 1, limit: 20, sort: "created_at" as const, order: "desc" as const };
const COMMENT_LOOKUP = { page: 1, limit: 50, sort: "created_at" as const, order: "asc" as const };
const ALERT_LOOKUP_LIMIT = 100;

const EVENT_ICONS: Record<TimelineEventType, LucideIcon> = {
  incident_created: AlertCircle,
  alert_fired: AlertTriangle,
  alert_acknowledged: CheckCircle2,
  deployment_occurred: Rocket,
  rollback_performed: Undo2,
  pod_restarted: RefreshCw,
  comment_added: MessageSquare,
  incident_resolved: CheckCircle2,
};

/**
 * The incident's real, multi-source story: incident created/resolved,
 * every attached alert firing/acknowledged, every deployment and rollback
 * on the affected application, comments, and any pod's own last known
 * restart - merged and sorted by lib/incident-timeline's buildIncidentTimeline.
 * Nothing here is fabricated: an event only ever appears because its
 * backing query returned a real record.
 */
export function IncidentCombinedTimeline({ incident }: { incident: IncidentResponse }) {
  const [selectedAlertId, setSelectedAlertId] = useState<number | null>(null);
  const [selectedDeployment, setSelectedDeployment] = useState<DeploymentResponse | null>(null);
  const hasApplication = Boolean(incident.applicationId);

  const alertsQuery = useAlerts(useMemo(() => ({ page: 1, limit: ALERT_LOOKUP_LIMIT, incidentId: incident.id }), [incident.id]));
  const commentsQuery = useComments(incident.id, COMMENT_LOOKUP);
  const deploymentsQuery = useDeploymentsByApplication(incident.applicationId ?? null, DEPLOYMENT_LOOKUP, hasApplication);
  const latestDeploymentQuery = useLatestDeployment(incident.applicationId ?? null, hasApplication);
  const namespace = latestDeploymentQuery.data?.namespace;
  const podsQuery = usePodsByApplication(
    incident.applicationId ?? null,
    { namespace: namespace ?? "" },
    hasApplication && Boolean(namespace)
  );

  const deployments = useMemo(() => deploymentsQuery.data?.items ?? [], [deploymentsQuery.data?.items]);
  const historyQueries = useQueries({
    queries: deployments.map((deployment) => ({
      queryKey: ["deployments", "history", deployment.id],
      queryFn: () => deploymentService.getDeploymentHistory(deployment.id),
      enabled: hasApplication,
    })),
  });
  const deploymentHistories = historyQueries.flatMap((query) => query.data?.items ?? []);

  const { data: projectsData } = useProjects(PROJECT_LOOKUP_QUERY);
  const projects = useMemo(() => projectsData?.items ?? [], [projectsData?.items]);
  const projectNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const project of projects) map.set(project.id, project.name);
    return map;
  }, [projects]);
  const deploymentById = useMemo(() => new Map(deployments.map((deployment) => [deployment.id, deployment])), [deployments]);

  const isCoreLoading = alertsQuery.isLoading || commentsQuery.isLoading || (hasApplication && deploymentsQuery.isLoading);
  const isCoreError = alertsQuery.isError || commentsQuery.isError || deploymentsQuery.isError;

  const events = useMemo(
    () =>
      buildIncidentTimeline({
        incident,
        alerts: alertsQuery.data?.items,
        comments: commentsQuery.data?.items,
        deployments: hasApplication ? deployments : undefined,
        deploymentHistories: hasApplication ? deploymentHistories : undefined,
        pods: podsQuery.data?.items,
      }),
    [incident, alertsQuery.data, commentsQuery.data, deployments, deploymentHistories, podsQuery.data, hasApplication]
  );

  if (isCoreError) {
    return (
      <ErrorState
        description="Unable to load the incident's combined timeline."
        onRetry={() => {
          void alertsQuery.refetch();
          void commentsQuery.refetch();
          void deploymentsQuery.refetch();
        }}
      />
    );
  }

  if (isCoreLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (events.length === 0) {
    return (
      <EmptyState
        icon={History}
        title="No timeline events yet"
        description="Once alerts, deployments, or comments are attached to this incident, they will appear here in order."
      />
    );
  }

  const handleEventClick = (event: TimelineEvent) => {
    if (!event.nav) return;
    if (event.nav.kind === "alert") {
      setSelectedAlertId(event.nav.alertId);
    } else if (event.nav.kind === "deployment") {
      const deployment = deploymentById.get(event.nav.deploymentId);
      if (deployment) setSelectedDeployment(deployment);
    }
  };

  return (
    <>
      <ol className="space-y-2">
        {events.map((event) => {
          const Icon = EVENT_ICONS[event.type];
          const clickable = Boolean(event.nav);

          return (
            <li key={event.id} className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
              <button
                type="button"
                disabled={!clickable}
                onClick={() => handleEventClick(event)}
                className="flex w-full items-start gap-3 text-left disabled:cursor-default"
              >
                <div
                  className="mt-0.5 rounded-xl p-1.5"
                  style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}
                >
                  <Icon aria-hidden="true" className="h-3.5 w-3.5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span
                      className="text-sm font-medium"
                      style={clickable ? { color: "var(--primary)", textDecorationLine: "underline", textUnderlineOffset: "2px" } : undefined}
                    >
                      {event.title}
                    </span>
                    <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                      {formatDateTime(event.timestamp)}
                    </span>
                  </div>
                  {event.description ? (
                    <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                      {event.description}
                    </p>
                  ) : null}
                </div>
              </button>
            </li>
          );
        })}
      </ol>

      <AlertDetailsDrawer
        alertID={selectedAlertId}
        projects={projects}
        projectNameById={projectNameById}
        onClose={() => setSelectedAlertId(null)}
      />
      <DeploymentDetailsDrawer
        deployment={selectedDeployment}
        open={Boolean(selectedDeployment)}
        onClose={() => setSelectedDeployment(null)}
      />
    </>
  );
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString();
}
