"use client";

import { useMemo } from "react";
import { Gauge } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import type { MetricChartSpec } from "@/components/metrics/scoped-metrics-dashboard";
import { ScopedMetricsDashboard } from "@/components/metrics/scoped-metrics-dashboard";
import { useResourceMetrics } from "@/hooks/use-metrics";
import { formatBytes, formatCount, formatMillicores, formatPercent, metricSource, pickLatestMetric } from "@/lib/metric-view";

const METRIC_LIMIT = 20;

const BASE_CHART_SPECS: MetricChartSpec[] = [
  { key: "cpu", title: "CPU usage", metricType: "CPU", metricName: "cluster.cpu.usage.millicores", valueFormatter: formatMillicores },
  { key: "memory", title: "Memory usage", metricType: "MEMORY", metricName: "cluster.memory.usage.bytes", valueFormatter: formatBytes },
];

/** The one additional chart that makes sense for this resource's kind, when a matching metric name exists. */
function kindSpecificSpec(kind: string): MetricChartSpec | null {
  switch (kind) {
    case "Node":
      return { key: "node-health", title: "Node health (not ready)", metricType: "AVAILABILITY", metricName: "cluster.node.not_ready.count", valueFormatter: formatCount };
    case "Deployment":
      return { key: "replicas", title: "Available replicas", metricType: "REPLICACOUNT", metricName: "cluster.deployment.available_replicas.total", valueFormatter: formatCount };
    case "Pod":
      return { key: "restarts", title: "Pod restarts", metricType: "RESTARTCOUNT", metricName: "cluster.pod.restarts.total", valueFormatter: formatCount };
    case "Cluster":
      return { key: "restarts", title: "Pod restarts", metricType: "RESTARTCOUNT", metricName: "cluster.pod.restarts.total", valueFormatter: formatCount };
    default:
      return null;
  }
}

/**
 * GET /resources/:id/metrics - functional today, but per-resource metric
 * collection (as opposed to cluster-wide aggregates) hasn't been built yet,
 * so this will honestly show an empty state for most resources rather than
 * inventing a value. The one exception is a resource representing the
 * cluster itself (Kind=Cluster, ID=the cluster's own ID) - that one does
 * have real historical data, since it's exactly what the cluster-level
 * collector persists.
 */
export function ResourceMetrics({ resourceId, projectId, kind }: { resourceId: string; projectId: string; kind: string }) {
  const cpuParams = useMemo(() => ({ projectId, metricType: "CPU", limit: METRIC_LIMIT }), [projectId]);
  const memoryParams = useMemo(() => ({ projectId, metricType: "MEMORY", limit: METRIC_LIMIT }), [projectId]);

  const cpu = useResourceMetrics(resourceId, cpuParams);
  const memory = useResourceMetrics(resourceId, memoryParams);

  const chartSpecs = useMemo(() => {
    const extra = kindSpecificSpec(kind);
    return extra ? [...BASE_CHART_SPECS, extra] : BASE_CHART_SPECS;
  }, [kind]);

  if (cpu.isLoading || memory.isLoading) return <ListSkeleton rows={2} />;

  if (cpu.isError || memory.isError) {
    return (
      <ErrorState
        description="Unable to load metrics for this resource. Please try again."
        onRetry={() => {
          void cpu.refetch();
          void memory.refetch();
        }}
      />
    );
  }

  const cpuUsage = pickLatestMetric(cpu.data ?? [], "cluster.cpu.usage.millicores");
  const cpuCapacity = pickLatestMetric(cpu.data ?? [], "cluster.cpu.capacity.millicores");
  const memoryUsage = pickLatestMetric(memory.data ?? [], "cluster.memory.usage.bytes");

  return (
    <div className="space-y-4">
      {!cpuUsage && !memoryUsage ? (
        <EmptyState
          icon={Gauge}
          title="No current metrics for this resource yet"
          description="OpsPilot currently collects metrics at the cluster level. Per-resource metric collection has not been implemented yet."
        />
      ) : (
        <dl className="grid grid-cols-2 gap-3">
          <div className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>CPU</dt>
            <dd className="mt-1 font-semibold">{formatPercent(cpuUsage?.value, cpuCapacity?.value)}</dd>
            <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
              {metricSource(cpuUsage) === "metrics-server" ? "Live usage" : "Estimated from requests"}
            </p>
          </div>
          <div className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Memory</dt>
            <dd className="mt-1 font-semibold">{formatBytes(memoryUsage?.value)}</dd>
            <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
              {metricSource(memoryUsage) === "metrics-server" ? "Live usage" : "Estimated from requests"}
            </p>
          </div>
        </dl>
      )}

      <div>
        <p className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Historical trends
        </p>
        <div className="mt-3">
          <ScopedMetricsDashboard projectId={projectId} resourceId={resourceId} specs={chartSpecs} />
        </div>
      </div>
    </div>
  );
}
