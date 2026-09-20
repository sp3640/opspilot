"use client";

import { Grid2X2, List } from "lucide-react";

import { ViewToggle } from "@/components/common/toolbar";

import type { ClusterView } from "./types";

export function ClusterViewToggle({ view, onChange }: { view: ClusterView; onChange: (view: ClusterView) => void }) {
  return (
    <ViewToggle
      label="Cluster view"
      value={view}
      onChange={onChange}
      options={[
        { value: "grid", label: "Grid", icon: Grid2X2 },
        { value: "table", label: "Table", icon: List },
      ]}
    />
  );
}
