"use client";

import { ToolbarSearch } from "@/components/common/toolbar";

export function ProjectSearch({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  return (
    <ToolbarSearch
      id="project-search"
      label="Search projects"
      placeholder="Search projects..."
      clearLabel="Clear project search"
      value={value}
      onChange={onChange}
    />
  );
}
