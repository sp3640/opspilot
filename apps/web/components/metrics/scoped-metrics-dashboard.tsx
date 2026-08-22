"use client";

import { useMemo, useState } from "react";

import { useMetricAggregation } from "@/hooks/use-metrics";
import { DEFAULT_METRIC_TIME_RANGE, resolveMetricTimeRange, type MetricTimeRangeKey } from "@/lib/metric-time-range";

import { MetricTimeRangeSelect } from "./metric-time-range-select";
import { MetricTimeSeriesChart } from "./metric-time-series-chart";

export type MetricChartSpec = {
  key: string;
  title: string;
  metricType: string;
  metricName?: string;
  unit?: string;
  valueFormatter?: (value: number) => string;
};

/**
 * A grid of historical metric charts sharing one time-range control, scoped
 * to a cluster or a single resource via GET /metrics/aggregate's optional
 * clusterId/resourceId (Phase 13's extension of Phase 12's architecture).
 */
export function ScopedMetricsDashboard({
  projectId,
  clusterId,
  resourceId,
  specs,
}: {
  projectId: string;
  clusterId?: string;
  resourceId?: string;
  specs: MetricChartSpec[];
}) {
  const [range, setRange] = useState<MetricTimeRangeKey>(DEFAULT_METRIC_TIME_RANGE);

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <MetricTimeRangeSelect value={range} onChange={setRange} />
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        {specs.map((spec) => (
          <ScopedMetricChart
            key={spec.key}
            projectId={projectId}
            clusterId={clusterId}
            resourceId={resourceId}
            range={range}
            spec={spec}
          />
        ))}
      </div>
    </div>
  );
}

function ScopedMetricChart({
  projectId,
  clusterId,
  resourceId,
  range,
  spec,
}: {
  projectId: string;
  clusterId?: string;
  resourceId?: string;
  range: MetricTimeRangeKey;
  spec: MetricChartSpec;
}) {
  const params = useMemo(() => {
    const { start, end, interval } = resolveMetricTimeRange(range);
    return {
      projectId,
      clusterId,
      resourceId,
      metricType: spec.metricType,
      metricName: spec.metricName,
      interval,
      start,
      end,
    };
  }, [projectId, clusterId, resourceId, range, spec.metricType, spec.metricName]);

  const { data, isLoading, isError, error, refetch } = useMetricAggregation(params);

  const points = (data?.points ?? []).map((point) => ({ timestamp: point.bucket, value: point.value, count: point.count }));

  return (
    <MetricTimeSeriesChart
      title={spec.title}
      points={points}
      unit={spec.unit ?? data?.unit}
      isLoading={isLoading}
      isError={isError}
      errorMessage={error instanceof Error ? error.message : undefined}
      onRetry={() => {
        void refetch();
      }}
      valueFormatter={spec.valueFormatter}
    />
  );
}
