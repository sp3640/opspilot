"use client";

import { Grid2X2, List } from "lucide-react";

import { ViewToggle } from "@/components/common/toolbar";

import type { ResourceView } from "./types";

export function ResourceViewToggle({ view, onChange }: { view: ResourceView; onChange: (view: ResourceView) => void }) {
  return (
    <ViewToggle
      label="Resource view"
      value={view}
      onChange={onChange}
      options={[
        { value: "grid", label: "Grid", icon: Grid2X2 },
        { value: "table", label: "Table", icon: List },
      ]}
    />
  );
}
