"use client";

import { useMemo } from "react";
import { AlertTriangle, CheckCircle2, ChevronDown, Clock, HeartPulse, Rocket } from "lucide-react";

import { ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useAlerts } from "@/hooks/use-alerts";
import { useIncidents } from "@/hooks/use-incidents";
import { useMetricAggregation } from "@/hooks/use-metrics";
import { usePodsByApplication } from "@/hooks/use-pods";
import { parseAlertMetadata } from "@/lib/alert-correlation";
import { PAGINATION_MAX_PAGE_SIZE } from "@/lib/constants";
import {
  computeDeploymentWindows,
  evaluateDeploymentHealthCorrelation,
  type ClusterMetricSignal,
  type CountWindowSignal,
  type DeploymentHealthCorrelationResult,
} from "@/lib/deployment-health-correlation";
import { formatBytes, formatMillicores } from "@/lib/metric-view";
import type { DeploymentResponse } from "@/types/deployment-api";

const APPLICATION_QUERY_LIMIT = PAGINATION_MAX_PAGE_SIZE;
const CPU_METRIC_NAME = "cluster.cpu.usage.millicores";
const MEMORY_METRIC_NAME = "cluster.memory.usage.bytes";

/**
 * "Did this deployment cause the problem?" - compares real application
 * health signals in the window before this deployment against the window
 * after it, and surfaces a hedged correlation verdict. Never claims
 * definite causation: wording is always "potentially related" /
 * "correlation detected" / "investigation recommended", and any signal
 * without real backing data (error rate, latency, request rate; a "before"
 * snapshot of pod state) is shown as explicitly unavailable rather than
 * guessed. No AI involved - this is a deterministic comparison of
 * already-real data (see lib/deployment-health-correlation.ts).
 */
export function DeploymentHealthCorrelation({ deployment }: { deployment: DeploymentResponse }) {
  const { beforeWindow, afterWindow } = useMemo(() => computeDeploymentWindows(deployment), [deployment]);

  const alertsQuery = useAlerts({ page: 1, limit: APPLICATION_QUERY_LIMIT, projectId: deployment.projectId });
  const incidentsQuery = useIncidents({ page: 1, limit: APPLICATION_QUERY_LIMIT, projectId: deployment.projectId });
  const podsQuery = usePodsByApplication(deployment.applicationId, { namespace: deployment.namespace });

  const cpuBeforeQuery = useMetricAggregation({
    projectId: deployment.projectId,
    clusterId: deployment.targetClusterId,
    metricType: "CPU",
    metricName: CPU_METRIC_NAME,
    start: beforeWindow.start,
    end: beforeWindow.end,
  });
  const cpuAfterQuery = useMetricAggregation({
    projectId: deployment.projectId,
    clusterId: deployment.targetClusterId,
    metricType: "CPU",
    metricName: CPU_METRIC_NAME,
    start: afterWindow.start,
    end: afterWindow.end,
  });
  const memoryBeforeQuery = useMetricAggregation({
    projectId: deployment.projectId,
    clusterId: deployment.targetClusterId,
    metricType: "MEMORY",
    metricName: MEMORY_METRIC_NAME,
    start: beforeWindow.start,
    end: beforeWindow.end,
  });
  const memoryAfterQuery = useMetricAggregation({
    projectId: deployment.projectId,
    clusterId: deployment.targetClusterId,
    metricType: "MEMORY",
    metricName: MEMORY_METRIC_NAME,
    start: afterWindow.start,
    end: afterWindow.end,
  });

  const applicationAlerts = useMemo(
    () => (alertsQuery.data?.items ?? []).filter((alert) => parseAlertMetadata(alert.metadata).applicationId === deployment.applicationId),
    [alertsQuery.data, deployment.applicationId]
  );
  const applicationIncidents = useMemo(
    () => (incidentsQuery.data?.items ?? []).filter((incident) => incident.applicationId === deployment.applicationId),
    [incidentsQuery.data, deployment.applicationId]
  );

  const isError = alertsQuery.isError || incidentsQuery.isError;
  const isLoading =
    alertsQuery.isLoading ||
    incidentsQuery.isLoading ||
    podsQuery.isLoading ||
    cpuBeforeQuery.isLoading ||
    cpuAfterQuery.isLoading ||
    memoryBeforeQuery.isLoading ||
    memoryAfterQuery.isLoading;

  const result = useMemo(() => {
    if (isLoading || isError) return null;
    return evaluateDeploymentHealthCorrelation({
      deployment,
      alerts: applicationAlerts,
      incidents: applicationIncidents,
      pods: podsQuery.data?.items,
      cpuBefore: cpuBeforeQuery.data,
      cpuAfter: cpuAfterQuery.data,
      memoryBefore: memoryBeforeQuery.data,
      memoryAfter: memoryAfterQuery.data,
    });
  }, [
    isLoading,
    isError,
    deployment,
    applicationAlerts,
    applicationIncidents,
    podsQuery.data,
    cpuBeforeQuery.data,
    cpuAfterQuery.data,
    memoryBeforeQuery.data,
    memoryAfterQuery.data,
  ]);

  if (isError) {
    return (
      <ErrorState
        description="Unable to load the data needed to correlate this deployment with application health."
        onRetry={() => {
          void alertsQuery.refetch();
          void incidentsQuery.refetch();
        }}
      />
    );
  }

  if (isLoading || !result) {
    return <ListSkeleton rows={4} />;
  }

  return (
    <div className="space-y-6">
      <FlowDiagram deployment={deployment} result={result} />
      <VerdictBanner result={result} />
      <SignalGrid result={result} />
    </div>
  );
}

function FlowDiagram({ deployment, result }: { deployment: DeploymentResponse; result: DeploymentHealthCorrelationResult }) {
  const nodes = [
    { label: `Deployment ${deploymentVersion(deployment)}`, icon: Rocket },
    { label: `Health before deployment (${result.windowMinutes}m window)`, icon: HeartPulse },
    { label: "Deployment executed", icon: Rocket },
    { label: `Health after deployment (${result.windowMinutes}m window)`, icon: HeartPulse },
  ];

  return (
    <div className="flex flex-col items-center gap-1">
      {nodes.map((node, index) => (
        <div key={node.label} className="flex w-full flex-col items-center">
          <div
            className="flex w-full max-w-sm items-center justify-center gap-2 rounded-2xl border px-4 py-3 text-center text-sm font-medium"
            style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
          >
            <node.icon aria-hidden="true" className="h-4 w-4 shrink-0" style={{ color: "var(--primary)" }} />
            {node.label}
          </div>
          {index < nodes.length - 1 ? (
            <ChevronDown aria-hidden="true" className="my-1 h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
          ) : null}
        </div>
      ))}
    </div>
  );
}

function VerdictBanner({ result }: { result: DeploymentHealthCorrelationResult }) {
  if (result.verdict === "potential_degradation") {
    return (
      <div
        className="rounded-2xl border p-4"
        style={{ borderColor: "var(--danger)", backgroundColor: "color-mix(in srgb, var(--danger) 10%, transparent)" }}
      >
        <div className="flex items-start gap-3">
          <AlertTriangle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" style={{ color: "var(--danger)" }} />
          <div>
            <p className="font-semibold" style={{ color: "var(--danger)" }}>
              Potential degradation detected after deployment.
            </p>
            <p className="mt-1 text-sm leading-6" style={{ color: "var(--foreground)" }}>
              The signals below are potentially related to this deployment. This is a correlation, not confirmed
              causation — investigation recommended.
            </p>
            <ul className="mt-2 list-disc space-y-1 pl-5 text-sm" style={{ color: "var(--foreground)" }}>
              {result.reasons.map((reason) => (
                <li key={reason}>{reason}</li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    );
  }

  if (result.verdict === "insufficient_data") {
    return (
      <div
        className="flex items-start gap-3 rounded-2xl border p-4"
        style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
      >
        <Clock aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" style={{ color: "var(--muted-foreground)" }} />
        <p className="text-sm leading-6">
          Not enough time has passed since this deployment to draw a conclusion — the {result.windowMinutes}-minute
          observation window hasn&apos;t fully elapsed yet. Check back shortly.
        </p>
      </div>
    );
  }

  return (
    <div
      className="flex items-start gap-3 rounded-2xl border p-4"
      style={{ borderColor: "var(--success)", backgroundColor: "color-mix(in srgb, var(--success) 10%, transparent)" }}
    >
      <CheckCircle2 aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" style={{ color: "var(--success)" }} />
      <p className="text-sm leading-6">
        No degradation detected in the {result.windowMinutes}-minute window after this deployment, based on alerts,
        incidents, and current pod health.
      </p>
    </div>
  );
}

function SignalGrid({ result }: { result: DeploymentHealthCorrelationResult }) {
  return (
    <dl className="grid grid-cols-2 gap-3">
      <CountTile label="Alerts" signal={result.alerts} unit="alert" />
      <CountTile label="Incidents" signal={result.incidents} unit="incident" />

      <SignalTile label="Pod restarts (since deployment)">
        {result.pods.status === "unavailable" ? (
          <NotAvailable reason="No pod data for this application." />
        ) : (
          <div className="flex items-center gap-2">
            <span className="font-semibold">{result.pods.restartedSinceDeployment}</span>
            <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>
              of {result.pods.totalPods} pod(s) — current state only, no pre-deployment snapshot exists
            </span>
          </div>
        )}
      </SignalTile>

      <SignalTile label="Pod readiness (current)">
        {result.pods.status === "unavailable" ? (
          <NotAvailable reason="No pod data for this application." />
        ) : (
          <StatusBadge variant={result.pods.unreadyPods > 0 ? "warning" : "success"}>
            {`${result.pods.totalPods - result.pods.unreadyPods}/${result.pods.totalPods} ready`}
          </StatusBadge>
        )}
      </SignalTile>

      <ClusterMetricTile label="CPU (cluster-wide)" signal={result.cpu} formatValue={formatMillicores} />
      <ClusterMetricTile label="Memory (cluster-wide)" signal={result.memory} formatValue={formatBytes} />

      <SignalTile label="Error rate">
        <NotAvailable reason="No APM/Prometheus integration connected yet." />
      </SignalTile>
      <SignalTile label="Latency">
        <NotAvailable reason="No APM/Prometheus integration connected yet." />
      </SignalTile>
      <SignalTile label="Request rate">
        <NotAvailable reason="No APM/Prometheus integration connected yet." />
      </SignalTile>
    </dl>
  );
}

function CountTile({ label, signal, unit }: { label: string; signal: CountWindowSignal; unit: string }) {
  return (
    <SignalTile label={label}>
      <div className="flex flex-wrap items-center gap-2">
        <span className="font-semibold">
          {signal.before} before → {signal.after} after
        </span>
        {signal.degraded ? <StatusBadge variant="warning">Increased</StatusBadge> : null}
      </div>
      <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
        {unit}(s) with firstSeenAt/createdAt in each window
      </p>
    </SignalTile>
  );
}

function ClusterMetricTile({
  label,
  signal,
  formatValue,
}: {
  label: string;
  signal: ClusterMetricSignal;
  formatValue: (value: number | null | undefined) => string;
}) {
  return (
    <SignalTile label={label}>
      {signal.status === "unavailable" ? (
        <NotAvailable reason="No metrics collected for this cluster in these windows." />
      ) : (
        <div>
          <span className="font-semibold">
            {formatValue(signal.before)} → {formatValue(signal.after)}
          </span>
          {signal.changePercent !== null ? (
            <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
              {signal.changePercent >= 0 ? "+" : ""}
              {signal.changePercent.toFixed(1)}% · cluster-wide, not application-scoped
            </p>
          ) : (
            <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
              cluster-wide, not application-scoped
            </p>
          )}
        </div>
      )}
    </SignalTile>
  );
}

function SignalTile({ label, children }: { label: string; children: React.ReactNode }) {
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

function NotAvailable({ reason }: { reason: string }) {
  return <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>Not available — {reason}</span>;
}

function deploymentVersion(deployment: DeploymentResponse): string {
  return deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image;
}
