"use client";

import { ToolbarSearch } from "@/components/common/toolbar";

export function IncidentSearch({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  return (
    <ToolbarSearch
      id="incident-search"
      label="Search incidents"
      placeholder="Search incidents..."
      clearLabel="Clear search"
      value={value}
      onChange={onChange}
    />
  );
}
