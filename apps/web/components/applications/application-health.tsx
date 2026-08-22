"use client";

import { useMemo } from "react";
import axios from "axios";
import { Boxes, Rocket } from "lucide-react";

import { ClusterStatusBadge } from "@/components/clusters/cluster-status";
import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useAlerts } from "@/hooks/use-alerts";
import { useApplicationHealthMetrics } from "@/hooks/use-application-health-metrics";
import { useCluster } from "@/hooks/use-clusters";
import { useLatestDeployment } from "@/hooks/use-deployments";
import { useIncidents } from "@/hooks/use-incidents";
import { usePodsByApplication } from "@/hooks/use-pods";
import { useRuntimeDeploymentsByApplication } from "@/hooks/use-runtime-deployments";
import { PAGINATION_MAX_PAGE_SIZE } from "@/lib/constants/pagination";
import { matchRuntimeDeployment } from "@/lib/deployment-replica-state";
import type { PodResponse } from "@/types/pod-api";

const ACTIVE_INCIDENT_STATUSES = ["OPEN", "INVESTIGATING"];
const ACTIVE_ALERT_STATUSES = ["OPEN", "ACKNOWLEDGED", "INVESTIGATING"];

const HEALTH_METRIC_FIELDS: ReadonlyArray<{ key: "availability" | "errorRate" | "latency" | "cpu" | "memory"; label: string }> = [
  { key: "availability", label: "Availability" },
  { key: "errorRate", label: "Error rate" },
  { key: "latency", label: "Latency" },
  { key: "cpu", label: "CPU" },
  { key: "memory", label: "Memory" },
];

/**
 * The primary "is this application healthy" summary: Status, Pod health,
 * Deployment status, and active Alerts/Incidents come from real data the
 * backend already provides. Availability/Error rate/Latency/CPU/Memory have
 * no data source yet (no Prometheus integration) — they render "Not
 * available" via `useApplicationHealthMetrics`, a placeholder hook shaped
 * exactly like the real one will be, so wiring up real metrics later never
 * requires touching this component.
 *
 * Below the health summary, "Runtime" connects the chain this data comes
 * from: Application -> Deployment -> Cluster -> Namespace -> Pods, plus
 * real replica state (ready/available) matched from the live k8s Deployment
 * object via matchRuntimeDeployment - "Not available" when no confident
 * match exists, never a guess.
 */
export function ApplicationHealth({
  applicationId,
  projectId,
  status,
  environment,
}: {
  applicationId: string;
  projectId: string;
  status: string;
  environment: string;
}) {
  const {
    data: deployment,
    error: deploymentError,
    isError: isDeploymentError,
    isLoading: isDeploymentLoading,
    refetch: refetchDeployment,
  } = useLatestDeployment(applicationId);

  const notDeployedYet = axios.isAxiosError(deploymentError) && deploymentError.response?.status === 404;

  const { data: cluster, isLoading: isClusterLoading } = useCluster(deployment?.targetClusterId ?? null);

  const podsParams = useMemo(() => ({ namespace: deployment?.namespace ?? "" }), [deployment?.namespace]);
  const {
    data: podsData,
    isError: isPodsError,
    isLoading: isPodsLoading,
  } = usePodsByApplication(applicationId, podsParams, Boolean(deployment?.namespace));
  const pods = useMemo(() => podsData?.items ?? [], [podsData]);
  const podHealth = useMemo(() => summarizePodHealth(pods), [pods]);

  const {
    data: alertsData,
    isError: isAlertsError,
    isLoading: isAlertsLoading,
  } = useAlerts({ page: 1, limit: PAGINATION_MAX_PAGE_SIZE, projectId });
  const activeAlertCount = (alertsData?.items ?? []).filter((alert) => ACTIVE_ALERT_STATUSES.includes(alert.status)).length;

  const {
    data: incidentsData,
    isError: isIncidentsError,
    isLoading: isIncidentsLoading,
  } = useIncidents({ page: 1, limit: PAGINATION_MAX_PAGE_SIZE, projectId });
  const activeIncidentCount = (incidentsData?.items ?? []).filter((incident) =>
    ACTIVE_INCIDENT_STATUSES.includes(incident.status)
  ).length;

  const { data: healthMetrics } = useApplicationHealthMetrics(applicationId);

  const { data: runtimeData } = useRuntimeDeploymentsByApplication(
    applicationId,
    { namespace: deployment?.namespace ?? "" },
    Boolean(deployment?.namespace)
  );
  const runtimeMatch = deployment ? matchRuntimeDeployment(deployment, runtimeData?.items ?? []) : null;

  return (
    <div className="space-y-6">
      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Environment
        </h3>
        <div className="mt-3">
          {environment ? (
            <span
              className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
              style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}
            >
              {environment}
            </span>
          ) : (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
              No environment set for this application.
            </p>
          )}
        </div>
      </section>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Health
        </h3>
        <dl className="mt-3 grid grid-cols-2 gap-3">
          <HealthStat label="Status">
            <StatusBadge variant={getApplicationStatusVariant(status)}>{status}</StatusBadge>
          </HealthStat>

          <HealthStat label="Deployment status">
            {isDeploymentLoading ? (
              <Loading />
            ) : notDeployedYet ? (
              <NotAvailable label="Not deployed yet" />
            ) : isDeploymentError ? (
              <NotAvailable label="Unable to load" />
            ) : deployment ? (
              <StatusBadge variant={getDeploymentStatusVariant(deployment.status)}>{deployment.status}</StatusBadge>
            ) : null}
          </HealthStat>

          <HealthStat label="Pod health">
            {isDeploymentLoading || (Boolean(deployment?.namespace) && isPodsLoading) ? (
              <Loading />
            ) : notDeployedYet ? (
              <NotAvailable label="Not deployed yet" />
            ) : isPodsError ? (
              <NotAvailable label="Unable to load" />
            ) : pods.length === 0 ? (
              <NotAvailable label="No pods found" />
            ) : (
              <div className="flex items-center gap-2">
                <StatusBadge variant={podHealth.variant}>{podHealth.label}</StatusBadge>
              </div>
            )}
          </HealthStat>

          <HealthStat label="Active alerts">
            {isAlertsLoading ? (
              <Loading />
            ) : isAlertsError ? (
              <NotAvailable label="Unable to load" />
            ) : (
              <span className="font-semibold">{activeAlertCount}</span>
            )}
          </HealthStat>

          <HealthStat label="Active incidents">
            {isIncidentsLoading ? (
              <Loading />
            ) : isIncidentsError ? (
              <NotAvailable label="Unable to load" />
            ) : (
              <span className="font-semibold">{activeIncidentCount}</span>
            )}
          </HealthStat>
        </dl>

        <div className="mt-4">
          <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
            Metrics below require a Prometheus integration that is not connected yet.
          </p>
          <dl className="mt-2 grid grid-cols-2 gap-3 sm:grid-cols-3">
            {HEALTH_METRIC_FIELDS.map(({ key, label }) => {
              const metric = healthMetrics[key];
              return (
                <HealthStat key={key} label={label}>
                  {metric.value === null ? (
                    <NotAvailable label="Not available" />
                  ) : (
                    <span className="font-semibold">
                      {metric.value}
                      {metric.unit ? ` ${metric.unit}` : ""}
                    </span>
                  )}
                </HealthStat>
              );
            })}
          </dl>
        </div>
      </section>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Runtime
        </h3>
        <div className="mt-3">
          {isDeploymentLoading ? (
            <ListSkeleton rows={2} />
          ) : notDeployedYet ? (
            <EmptyState
              icon={Rocket}
              title="Not deployed yet"
              description="Deploy this application to see its runtime cluster, namespace, and pods here."
            />
          ) : isDeploymentError ? (
            <ErrorState
              description={
                deploymentError instanceof Error ? deploymentError.message : "Unable to load runtime information."
              }
              onRetry={() => {
                void refetchDeployment();
              }}
            />
          ) : deployment ? (
            <dl className="grid grid-cols-2 gap-3">
              <HealthStat label="Cluster">
                {isClusterLoading ? (
                  <Loading />
                ) : cluster ? (
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-semibold">{cluster.name}</span>
                    <ClusterStatusBadge status={cluster.status} />
                  </div>
                ) : (
                  <span className="break-all font-semibold">{deployment.targetClusterId}</span>
                )}
              </HealthStat>
              <HealthStat label="Namespace">
                <span className="break-all font-semibold">{deployment.namespace}</span>
              </HealthStat>
              <HealthStat label="Deployment">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="break-all font-semibold">
                    {deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image}
                  </span>
                  <StatusBadge variant={getDeploymentStatusVariant(deployment.status)}>{deployment.status}</StatusBadge>
                </div>
              </HealthStat>
              <HealthStat label="Pods">
                {isPodsLoading ? (
                  <Loading />
                ) : pods.length === 0 ? (
                  <NotAvailable label="No pods found" />
                ) : (
                  <div className="flex flex-wrap items-center gap-2">
                    <Boxes aria-hidden="true" className="h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
                    <span className="font-semibold">{pods.length} pod{pods.length === 1 ? "" : "s"}</span>
                    <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>{podHealth.breakdown}</span>
                  </div>
                )}
              </HealthStat>
              <HealthStat label="Replica state">
                {runtimeMatch ? (
                  <span className="font-semibold">
                    {runtimeMatch.readyReplicas}/{runtimeMatch.replicas} ready, {runtimeMatch.availableReplicas} available
                  </span>
                ) : (
                  <NotAvailable label="Not available" />
                )}
              </HealthStat>
            </dl>
          ) : null}
        </div>
      </section>
    </div>
  );
}

function HealthStat({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div
      className="rounded-2xl border p-4"
      style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
    >
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</dt>
      <dd className="mt-1">{children}</dd>
    </div>
  );
}

function Loading() {
  return <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>Loading...</span>;
}

function NotAvailable({ label }: { label: string }) {
  return <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>{label}</span>;
}

function summarizePodHealth(pods: PodResponse[]): {
  variant: "success" | "warning" | "critical" | "info";
  label: string;
  breakdown: string;
} {
  let healthy = 0;
  let degraded = 0;
  let unhealthy = 0;

  for (const pod of pods) {
    if (pod.phase === "Running" || pod.phase === "Succeeded") {
      healthy += pod.restartCount > 0 ? 0 : 1;
      if (pod.restartCount > 0) degraded += 1;
    } else if (pod.phase === "Pending") {
      degraded += 1;
    } else {
      unhealthy += 1;
    }
  }

  const total = pods.length;
  const breakdown = `${healthy} healthy, ${degraded} degraded, ${unhealthy} unhealthy`;

  if (unhealthy > 0) {
    return { variant: "critical", label: `${unhealthy}/${total} unhealthy`, breakdown };
  }
  if (degraded > 0) {
    return { variant: "warning", label: `${degraded}/${total} degraded`, breakdown };
  }
  return { variant: "success", label: `${healthy}/${total} healthy`, breakdown };
}

function getApplicationStatusVariant(status: string): "success" | "info" | "archived" {
  switch (status) {
    case "Ready":
      return "success";
    case "Archived":
      return "archived";
    default:
      return "info";
  }
}

function getDeploymentStatusVariant(status: string): "success" | "warning" | "critical" | "archived" | "info" {
  switch (status) {
    case "Succeeded":
      return "success";
    case "Failed":
      return "critical";
    case "Pending":
    case "Queued":
    case "RolledBack":
      return "warning";
    case "Cancelled":
      return "archived";
    default:
      return "info";
  }
}
