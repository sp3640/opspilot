"use client";

import { Plus, RefreshCw } from "lucide-react";

import { Toolbar } from "@/components/common/toolbar";
import { Button } from "@/components/ui/button";
import { useHasPermission } from "@/store/auth-store";
import type { ProjectResponse } from "@/types/project-api";

import { ClusterFilter } from "./cluster-filter";
import { ClusterSearch } from "./cluster-search";
import { ClusterViewToggle } from "./cluster-view-toggle";
import type { ClusterFilters, ClusterView } from "./types";

type ClusterToolbarProps = {
  filters: ClusterFilters;
  view: ClusterView;
  projects: ProjectResponse[];
  onFiltersChange: (filters: Partial<ClusterFilters>) => void;
  onViewChange: (view: ClusterView) => void;
  onRefresh: () => void;
  onCreate: () => void;
  refreshing?: boolean;
};

export function ClusterToolbar({ filters, view, projects, onFiltersChange, onViewChange, onRefresh, onCreate, refreshing = false }: ClusterToolbarProps) {
  const canManageClusters = useHasPermission("cluster:manage");
  return (
    <Toolbar
      search={<ClusterSearch value={filters.query} onChange={(query) => onFiltersChange({ query })} />}
      actions={
        <>
          <ClusterViewToggle view={view} onChange={onViewChange} />
          <Button type="button" variant="ghost" onClick={onRefresh} loading={refreshing} className="px-3" aria-label="Refresh clusters">
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>
          {canManageClusters ? (
            <Button type="button" onClick={onCreate}>
              <Plus aria-hidden="true" className="h-4 w-4" />
              Create cluster
            </Button>
          ) : null}
        </>
      }
      filters={<ClusterFilter filters={filters} projects={projects} onFiltersChange={onFiltersChange} />}
    />
  );
}
