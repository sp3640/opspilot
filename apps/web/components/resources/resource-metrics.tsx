"use client";

import { useMemo } from "react";
import { Gauge } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useResourceMetrics } from "@/hooks/use-metrics";
import { formatBytes, formatPercent, metricSource, pickLatestMetric } from "@/lib/metric-view";

const METRIC_LIMIT = 20;

/**
 * GET /resources/:id/metrics - functional today, but per-resource metric
 * collection (as opposed to cluster-wide aggregates) hasn't been built yet,
 * so this will honestly show an empty state for most resources rather than
 * inventing a value. It reuses the same CPU/Memory history contract the
 * cluster metrics panel uses, scoped to this one resource.
 */
export function ResourceMetrics({ resourceId, projectId }: { resourceId: string; projectId: string }) {
  const cpuParams = useMemo(() => ({ projectId, metricType: "CPU", limit: METRIC_LIMIT }), [projectId]);
  const memoryParams = useMemo(() => ({ projectId, metricType: "MEMORY", limit: METRIC_LIMIT }), [projectId]);

  const cpu = useResourceMetrics(resourceId, cpuParams);
  const memory = useResourceMetrics(resourceId, memoryParams);

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

  if (!cpuUsage && !memoryUsage) {
    return (
      <EmptyState
        icon={Gauge}
        title="No metrics recorded for this resource yet"
        description="OpsPilot currently collects metrics at the cluster level. Per-resource metric collection has not been implemented yet."
      />
    );
  }

  return (
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
  );
}
