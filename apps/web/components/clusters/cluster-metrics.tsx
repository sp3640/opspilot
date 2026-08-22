"use client";

import { useMemo } from "react";
import { AlertTriangle, Gauge } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import type { MetricChartSpec } from "@/components/metrics/scoped-metrics-dashboard";
import { ScopedMetricsDashboard } from "@/components/metrics/scoped-metrics-dashboard";
import { useClusterMetrics, useClusterMetricsStatus } from "@/hooks/use-metrics";
import { formatBytes, formatCount, formatMillicores, formatPercent, metricSource, pickLatestMetric } from "@/lib/metric-view";
import type { MetricsStatusResponse } from "@/types/metric-api";

export const CLUSTER_HISTORICAL_CHART_SPECS: MetricChartSpec[] = [
  { key: "cpu", title: "CPU usage", metricType: "CPU", metricName: "cluster.cpu.usage.millicores", valueFormatter: formatMillicores },
  { key: "memory", title: "Memory usage", metricType: "MEMORY", metricName: "cluster.memory.usage.bytes", valueFormatter: formatBytes },
  { key: "node-health", title: "Node health (not ready)", metricType: "AVAILABILITY", metricName: "cluster.node.not_ready.count", valueFormatter: formatCount },
  { key: "pod-count", title: "Pod count", metricType: "AVAILABILITY", metricName: "cluster.pod.count", valueFormatter: formatCount },
  { key: "restarts", title: "Pod restarts", metricType: "RESTARTCOUNT", metricName: "cluster.pod.restarts.total", valueFormatter: formatCount },
];

const METRIC_LIMIT = 20;

/**
 * Kubernetes -> Metrics Provider -> OpsPilot Metrics Service -> Frontend.
 * Uses the previously-unused GET /clusters/:id/metrics contract directly
 * (path-scoped, not the generic /metrics/history query-param form).
 *
 * Infrastructure CPU/Memory/Storage come from the Kubernetes collector
 * (real metrics-server usage when installed, otherwise an honestly-labeled
 * requested-capacity estimate). Network and Application metrics have no
 * data source today - they render "Not available" rather than a fabricated
 * number, per the provider-status contract from GET /clusters/:id/metrics/status.
 */
export function ClusterMetrics({ clusterId, projectId }: { clusterId: string; projectId: string }) {
  const { data: status } = useClusterMetricsStatus(clusterId);

  const cpuParams = useMemo(() => ({ projectId, metricType: "CPU", limit: METRIC_LIMIT }), [projectId]);
  const memoryParams = useMemo(() => ({ projectId, metricType: "MEMORY", limit: METRIC_LIMIT }), [projectId]);
  const storageParams = useMemo(() => ({ projectId, metricType: "STORAGE", limit: METRIC_LIMIT }), [projectId]);
  const restartParams = useMemo(() => ({ projectId, metricType: "RESTARTCOUNT", limit: METRIC_LIMIT }), [projectId]);
  const replicaParams = useMemo(() => ({ projectId, metricType: "REPLICACOUNT", limit: METRIC_LIMIT }), [projectId]);
  const availabilityParams = useMemo(() => ({ projectId, metricType: "AVAILABILITY", limit: METRIC_LIMIT }), [projectId]);

  const cpu = useClusterMetrics(clusterId, cpuParams);
  const memory = useClusterMetrics(clusterId, memoryParams);
  const storage = useClusterMetrics(clusterId, storageParams);
  const restarts = useClusterMetrics(clusterId, restartParams);
  const replicas = useClusterMetrics(clusterId, replicaParams);
  const availability = useClusterMetrics(clusterId, availabilityParams);

  const isLoading = [cpu, memory, storage, restarts, replicas, availability].some((q) => q.isLoading);
  const isError = [cpu, memory, storage, restarts, replicas, availability].some((q) => q.isError);

  if (isLoading) return <ListSkeleton rows={6} />;

  if (isError) {
    return (
      <ErrorState
        description="Unable to load cluster metrics. Please try again."
        onRetry={() => {
          [cpu, memory, storage, restarts, replicas, availability].forEach((q) => void q.refetch());
        }}
      />
    );
  }

  const cpuUsage = pickLatestMetric(cpu.data ?? [], "cluster.cpu.usage.millicores");
  const cpuCapacity = pickLatestMetric(cpu.data ?? [], "cluster.cpu.capacity.millicores");
  const memoryUsage = pickLatestMetric(memory.data ?? [], "cluster.memory.usage.bytes");
  const memoryCapacity = pickLatestMetric(memory.data ?? [], "cluster.memory.capacity.bytes");
  const storageUsage = pickLatestMetric(storage.data ?? [], "cluster.storage.usage.bytes");
  const storageCapacity = pickLatestMetric(storage.data ?? [], "cluster.storage.capacity.bytes");

  const podRestarts = pickLatestMetric(restarts.data ?? [], "cluster.pod.restarts.total");
  const availableReplicas = pickLatestMetric(replicas.data ?? [], "cluster.deployment.available_replicas.total");
  const unavailableReplicas = pickLatestMetric(replicas.data ?? [], "cluster.deployment.unavailable_replicas.total");
  const nodeCount = pickLatestMetric(availability.data ?? [], "cluster.node.count");
  const nodeNotReady = pickLatestMetric(availability.data ?? [], "cluster.node.not_ready.count");
  const podCount = pickLatestMetric(availability.data ?? [], "cluster.pod.count");
  const podNotReady = pickLatestMetric(availability.data ?? [], "cluster.pod.not_ready.count");

  const hasAnyData = Boolean(cpuUsage || memoryUsage || storageUsage || podRestarts || nodeCount);

  return (
    <div className="space-y-6">
      {status ? <ProviderStatusBanner status={status} /> : null}

      {!hasAnyData ? (
        <EmptyState
          icon={Gauge}
          title="No metrics collected yet"
          description="Metrics are collected roughly once a minute while the cluster is reachable. Check back shortly."
        />
      ) : (
        <>
          <section>
            <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
              Infrastructure
            </h3>
            <dl className="mt-3 grid grid-cols-2 gap-3">
              <Tile label="CPU" value={formatPercent(cpuUsage?.value, cpuCapacity?.value)} caption={sourceCaption(cpuUsage ? metricSource(cpuUsage) : null)} />
              <Tile label="Memory" value={formatPercent(memoryUsage?.value, memoryCapacity?.value)} caption={sourceCaption(memoryUsage ? metricSource(memoryUsage) : null)} />
              <Tile label="Storage" value={formatPercent(storageUsage?.value, storageCapacity?.value)} caption={sourceCaption(storageUsage ? metricSource(storageUsage) : null)} />
              <Tile label="Network" value="Not available" caption="No network metrics source is configured yet." />
            </dl>
            <p className="mt-2 text-xs" style={{ color: "var(--muted-foreground)" }}>
              CPU usage: {formatMillicores(cpuUsage?.value)} of {formatMillicores(cpuCapacity?.value)} capacity · Memory usage: {formatBytes(memoryUsage?.value)} of {formatBytes(memoryCapacity?.value)} capacity
            </p>
          </section>

          <section>
            <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
              Kubernetes
            </h3>
            <dl className="mt-3 grid grid-cols-2 gap-3">
              <Tile label="Node health" value={readyRatio(nodeCount?.value, nodeNotReady?.value)} />
              <Tile label="Pod availability" value={readyRatio(podCount?.value, podNotReady?.value)} />
              <Tile label="Pod restarts" value={formatCount(podRestarts?.value)} />
              <Tile
                label="Deployment replica availability"
                value={
                  availableReplicas
                    ? `${formatCount(availableReplicas.value)} available${unavailableReplicas?.value ? ` / ${formatCount(unavailableReplicas.value)} unavailable` : ""}`
                    : "Not available"
                }
              />
            </dl>
          </section>
        </>
      )}

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Historical trends
        </h3>
        <div className="mt-3">
          <ScopedMetricsDashboard projectId={projectId} clusterId={clusterId} specs={CLUSTER_HISTORICAL_CHART_SPECS} />
        </div>
      </section>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Application
        </h3>
        <div className="mt-3">
          <EmptyState
            icon={AlertTriangle}
            title="Not available"
            description={status?.applicationMetrics.reason ?? "No application observability provider is connected yet."}
          />
        </div>
      </section>
    </div>
  );
}

function Tile({ label, value, caption }: { label: string; value: string; caption?: string }) {
  return (
    <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}>
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</dt>
      <dd className="mt-1 font-semibold">{value}</dd>
      {caption ? <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>{caption}</p> : null}
    </div>
  );
}

function ProviderStatusBanner({ status }: { status: MetricsStatusResponse }) {
  if (!status.kubernetesReachable) {
    return (
      <Banner variant="critical">
        Metrics unavailable: this cluster is not reachable right now.
      </Banner>
    );
  }

  if (!status.metricsServerAvailable) {
    return (
      <Banner variant="warning">
        metrics-server is not installed on this cluster, so CPU/Memory below are estimated from declared resource
        requests rather than live usage.
      </Banner>
    );
  }

  return (
    <Banner variant="success">
      Live usage from metrics-server{status.lastCollectedAt ? ` · last collected ${new Date(status.lastCollectedAt).toLocaleString()}` : ""}.
    </Banner>
  );
}

function Banner({ variant, children }: { variant: "critical" | "warning" | "success"; children: React.ReactNode }) {
  const color = variant === "critical" ? "var(--danger)" : variant === "warning" ? "var(--warning)" : "var(--success)";
  return (
    <p
      className="rounded-xl border p-3 text-xs"
      style={{ borderColor: color, color, backgroundColor: `color-mix(in srgb, ${color} 8%, transparent)` }}
    >
      {children}
    </p>
  );
}

function sourceCaption(source: string | null): string | undefined {
  if (source === "metrics-server") return "Live usage";
  if (source === "requested-capacity") return "Estimated from requests";
  return undefined;
}

function readyRatio(total: number | null | undefined, notReady: number | null | undefined): string {
  if (total === null || total === undefined) return "Not available";
  const ready = total - (notReady ?? 0);
  return `${formatCount(ready)}/${formatCount(total)} ready`;
}
