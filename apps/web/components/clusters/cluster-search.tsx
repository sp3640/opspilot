"use client";

import { ToolbarSearch } from "@/components/common/toolbar";

export function ClusterSearch({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  return (
    <ToolbarSearch
      id="cluster-search"
      label="Search clusters"
      placeholder="Search clusters..."
      clearLabel="Clear cluster search"
      value={value}
      onChange={onChange}
    />
  );
}
