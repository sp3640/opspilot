"use client";

import { useMemo } from "react";
import { AlertTriangle, ClipboardList } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { AlertSeverityBadge } from "@/components/alerts/alert-severity-badge";
import { AlertStatusBadge } from "@/components/alerts/alert-status-badge";
import { useAlerts } from "@/hooks/use-alerts";
import { useIncident, useIncidents } from "@/hooks/use-incidents";
import { findRelatedAlerts } from "@/lib/alert-correlation";
import { PAGINATION_MAX_PAGE_SIZE } from "@/lib/constants/pagination";
import type { AlertResponse } from "@/types/alert-api";

const RELATED_TIME_WINDOW_MS = 24 * 60 * 60 * 1000;

/**
 * Related alerts: alerts sharing this alert's exact resourceType+resourceId
 * (a real identifier match, not a guess).
 *
 * Related incidents: the explicitly attached incident (the strongest
 * correlation, already tracked by Alert.incidentId) plus other incidents in
 * the same project whose createdAt falls within 24h of this alert's
 * firstSeenAt - an honest, timestamp-based heuristic, labeled as such rather
 * than presented as a confirmed link.
 */
export function AlertRelated({ alert }: { alert: AlertResponse }) {
  const alertsParams = useMemo(() => ({ page: 1, limit: PAGINATION_MAX_PAGE_SIZE, projectId: alert.projectId }), [alert.projectId]);
  const { data: alertsData, error: alertsError, isError: isAlertsError, isLoading: isAlertsLoading, refetch: refetchAlerts } = useAlerts(alertsParams);
  const relatedAlerts = useMemo(() => findRelatedAlerts(alert, alertsData?.items ?? []), [alert, alertsData]);

  const { data: attachedIncident } = useIncident(alert.incidentId ?? null);

  const incidentsParams = useMemo(() => ({ page: 1, limit: PAGINATION_MAX_PAGE_SIZE, projectId: alert.projectId, sort: "created_at" as const, order: "desc" as const }), [alert.projectId]);
  const { data: incidentsData } = useIncidents(incidentsParams);
  const overlappingIncidents = useMemo(() => {
    const firstSeen = new Date(alert.firstSeenAt).getTime();
    return (incidentsData?.items ?? []).filter((incident) => {
      if (alert.incidentId && incident.id === alert.incidentId) return false;
      const createdAt = new Date(incident.createdAt).getTime();
      return Math.abs(createdAt - firstSeen) <= RELATED_TIME_WINDOW_MS;
    });
  }, [incidentsData, alert.firstSeenAt, alert.incidentId]);

  return (
    <div className="space-y-6">
      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Related alerts
        </h3>
        <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Other alerts for the exact same resource ({alert.resourceType}: {alert.resourceId}).
        </p>
        <div className="mt-3">
          {isAlertsError ? (
            <ErrorState
              description={alertsError instanceof Error ? alertsError.message : "Unable to load related alerts."}
              onRetry={() => {
                void refetchAlerts();
              }}
            />
          ) : isAlertsLoading ? (
            <ListSkeleton rows={2} />
          ) : relatedAlerts.length === 0 ? (
            <EmptyState icon={AlertTriangle} title="No related alerts" description="No other alerts share this alert's resource." />
          ) : (
            <ul className="space-y-2">
              {relatedAlerts.map((related) => (
                <li key={related.id} className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
                  <div className="flex items-start justify-between gap-3">
                    <p className="min-w-0 truncate text-sm font-medium">{related.title}</p>
                    <div className="flex shrink-0 gap-1.5">
                      <AlertSeverityBadge severity={related.severity} />
                      <AlertStatusBadge status={related.status} />
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Related incidents
        </h3>
        <div className="mt-3 space-y-3">
          {attachedIncident ? (
            <div className="rounded-2xl border p-3" style={{ borderColor: "var(--primary)" }}>
              <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>Linked incident</p>
              <p className="mt-1 text-sm font-medium">{attachedIncident.title}</p>
            </div>
          ) : null}

          {overlappingIncidents.length === 0 ? (
            attachedIncident ? null : (
              <EmptyState
                icon={ClipboardList}
                title="No related incidents"
                description="No incidents in this project were created within 24 hours of this alert first firing."
              />
            )
          ) : (
            <>
              <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                Possibly related — created within 24 hours of this alert, same project (not a confirmed link):
              </p>
              <ul className="space-y-2">
                {overlappingIncidents.map((incident) => (
                  <li key={incident.id} className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
                    <div className="flex items-start justify-between gap-3">
                      <p className="min-w-0 truncate text-sm font-medium">{incident.title}</p>
                      <StatusBadge variant="info">{incident.status}</StatusBadge>
                    </div>
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>
      </section>
    </div>
  );
}
