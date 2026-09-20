"use client";

import { ToolbarSearch } from "@/components/common/toolbar";

export function ResourceSearch({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  return (
    <ToolbarSearch
      id="resource-search"
      label="Search resources"
      placeholder="Search resources..."
      clearLabel="Clear resource search"
      value={value}
      onChange={onChange}
    />
  );
}
