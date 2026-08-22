"use client";

import { METRIC_TIME_RANGE_OPTIONS, type MetricTimeRangeKey } from "@/lib/metric-time-range";

export function MetricTimeRangeSelect({
  value,
  onChange,
}: {
  value: MetricTimeRangeKey;
  onChange: (value: MetricTimeRangeKey) => void;
}) {
  return (
    <select
      aria-label="Time range"
      value={value}
      onChange={(event) => onChange(event.target.value as MetricTimeRangeKey)}
      className="h-10 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]"
      style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
    >
      {METRIC_TIME_RANGE_OPTIONS.map((option) => (
        <option key={option.key} value={option.key}>{option.label}</option>
      ))}
    </select>
  );
}
