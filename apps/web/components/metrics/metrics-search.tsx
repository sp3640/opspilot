"use client";

import { ToolbarSearch } from "@/components/common/toolbar";

export function MetricsSearch({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  return (
    <ToolbarSearch
      id="metrics-search"
      label="Search metrics"
      placeholder="Search metrics..."
      clearLabel="Clear metrics search"
      value={value}
      onChange={onChange}
    />
  );
}
