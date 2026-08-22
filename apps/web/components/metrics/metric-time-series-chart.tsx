"use client";

import { useMemo } from "react";
import { AlertTriangle, Gauge } from "lucide-react";
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { formatChartTimestamp } from "@/lib/metric-view";

export type MetricChartPoint = { timestamp: string; value: number; count?: number };

type TooltipPayloadEntry = { value?: number; payload?: MetricChartPoint & { label: string } };

/**
 * A single historical metric chart card, covering the states Phase 13
 * requires: loading, error (with retry), no-data, and an explicit
 * "unavailable" state distinct from no-data (used when no metrics provider
 * exists for this series at all, e.g. application-level metrics today).
 * Built on recharts/the existing chart visual language (see
 * components/metrics/metric-charts.tsx), extended with real tooltips,
 * readable axis timestamps (never hidden), and honest empty states.
 */
export function MetricTimeSeriesChart({
  title,
  points,
  unit,
  isLoading = false,
  isError = false,
  errorMessage,
  onRetry,
  unavailableReason,
  valueFormatter,
}: {
  title: string;
  points: MetricChartPoint[];
  unit?: string;
  isLoading?: boolean;
  isError?: boolean;
  errorMessage?: string;
  onRetry?: () => void;
  /** When set, renders a distinct "not available" state instead of the chart, regardless of `points`. */
  unavailableReason?: string;
  valueFormatter?: (value: number) => string;
}) {
  const format = useMemo(
    () => valueFormatter ?? ((value: number) => `${roundForDisplay(value)}${unit ? ` ${unit}` : ""}`),
    [valueFormatter, unit]
  );

  const chartData = useMemo(() => {
    if (points.length === 0) return [];
    const span = new Date(points[points.length - 1]!.timestamp).getTime() - new Date(points[0]!.timestamp).getTime();
    return points.map((point) => ({ ...point, label: formatChartTimestamp(point.timestamp, span) }));
  }, [points]);

  return (
    <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
      <h3 className="text-sm font-semibold">{title}</h3>
      <div className="mt-3">
        {unavailableReason ? (
          <EmptyState icon={AlertTriangle} title="Not available" description={unavailableReason} />
        ) : isError ? (
          <ErrorState description={errorMessage ?? "Unable to load this chart. Please try again."} onRetry={onRetry} />
        ) : isLoading ? (
          <ListSkeleton rows={3} />
        ) : chartData.length === 0 ? (
          <EmptyState icon={Gauge} title="No data" description="No metrics were collected for this time range." />
        ) : (
          <ResponsiveContainer width="100%" height={220}>
            <AreaChart data={chartData} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
              <CartesianGrid stroke="var(--border)" strokeDasharray="3 3" />
              <XAxis dataKey="label" tick={{ fontSize: 11, fill: "var(--muted-foreground)" }} minTickGap={32} />
              <YAxis tick={{ fontSize: 11, fill: "var(--muted-foreground)" }} width={56} />
              <Tooltip content={<ChartTooltip format={format} />} />
              <Area
                type="monotone"
                dataKey="value"
                stroke="var(--primary)"
                fill="color-mix(in srgb, var(--primary) 20%, transparent)"
                isAnimationActive={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        )}
      </div>
    </div>
  );
}

function ChartTooltip({
  active,
  payload,
  format,
}: {
  active?: boolean;
  payload?: TooltipPayloadEntry[];
  format: (value: number) => string;
}) {
  if (!active || !payload || payload.length === 0) return null;
  const entry = payload[0];
  const timestamp = entry?.payload?.timestamp;

  return (
    <div
      className="rounded-xl border px-3 py-2 text-xs shadow-[var(--shadow-sm)]"
      style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
    >
      <p style={{ color: "var(--muted-foreground)" }}>{timestamp ? new Date(timestamp).toLocaleString() : ""}</p>
      <p className="mt-0.5 font-semibold">{format(Number(entry?.value ?? 0))}</p>
    </div>
  );
}

function roundForDisplay(value: number): string {
  return String(Math.round(value * 100) / 100);
}
