"use client";

import { Grid2X2, List } from "lucide-react";

import { ViewToggle } from "@/components/common/toolbar";

import type { ProjectView } from "./types";

/** Accessible switch between the project card and compact table representations. */
export function ProjectViewToggle({ view, onChange }: { view: ProjectView; onChange: (view: ProjectView) => void }) {
  return (
    <ViewToggle
      label="Project view"
      value={view}
      onChange={onChange}
      options={[
        { value: "grid", label: "Grid", icon: Grid2X2 },
        { value: "table", label: "Table", icon: List },
      ]}
    />
  );
}
