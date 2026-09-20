"use client";

import { Grid2X2, List, Plus, RefreshCw } from "lucide-react";

import { Toolbar, ViewToggle } from "@/components/common/toolbar";
import { Button } from "@/components/ui/button";
import { useHasPermission } from "@/store/auth-store";
import type { ClusterResponse } from "@/types/cluster-api";
import type { ProjectResponse } from "@/types/project-api";

import { AlertFilter } from "./alert-filter";
import { AlertSearch } from "./alert-search";
import type { AlertFilters, AlertView } from "./types";

type AlertToolbarProps = {
  filters: AlertFilters;
  view: AlertView;
  projects: ProjectResponse[];
  clusters: ClusterResponse[];
  onFiltersChange: (filters: Partial<AlertFilters>) => void;
  onViewChange: (view: AlertView) => void;
  onRefresh: () => void;
  onCreate: () => void;
  refreshing: boolean;
};

export function AlertToolbar({ filters, view, projects, clusters, onFiltersChange, onViewChange, onRefresh, onCreate, refreshing }: AlertToolbarProps) {
  const canManageAlerts = useHasPermission("alert:manage");
  return (
    <Toolbar
      search={<AlertSearch value={filters.query} onChange={(query) => onFiltersChange({ query })} />}
      actions={
        <>
          <ViewToggle
            label="Alert view"
            value={view}
            onChange={onViewChange}
            options={[
              { value: "grid", label: "Grid", icon: Grid2X2 },
              { value: "table", label: "Table", icon: List },
            ]}
          />
          <Button type="button" variant="ghost" onClick={onRefresh} loading={refreshing} className="px-3" aria-label="Refresh alerts">
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>
          {canManageAlerts ? (
            <Button type="button" onClick={onCreate}>
              <Plus aria-hidden="true" className="h-4 w-4" />
              Create alert
            </Button>
          ) : null}
        </>
      }
      filters={<AlertFilter filters={filters} projects={projects} clusters={clusters} onChange={onFiltersChange} />}
    />
  );
}
