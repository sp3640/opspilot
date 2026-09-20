"use client";

import { ToolbarSearch } from "@/components/common/toolbar";

export function AlertSearch({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  return (
    <ToolbarSearch
      id="alert-search"
      label="Search alerts"
      placeholder="Search alerts..."
      clearLabel="Clear alert search"
      value={value}
      onChange={onChange}
    />
  );
}
