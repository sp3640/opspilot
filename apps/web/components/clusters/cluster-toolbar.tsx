"use client";

import { Plus, RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
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

export function ClusterToolbar({
  filters,
  view,
  projects,
  onFiltersChange,
  onViewChange,
  onRefresh,
  onCreate,
  refreshing = false,
}: ClusterToolbarProps) {
  return (
    <div className="space-y-3">
      <div className="flex flex-col gap-3 lg:flex-row">
        <ClusterSearch value={filters.query} onChange={(query) => onFiltersChange({ query })} />
        <ClusterFilter
          filters={filters}
          projects={projects}
          onFiltersChange={onFiltersChange}
        />
      </div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <ClusterViewToggle view={view} onChange={onViewChange} />
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="ghost"
            onClick={onRefresh}
            loading={refreshing}
            className="px-3"
            aria-label="Refresh clusters"
          >
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>
          <Button type="button" onClick={onCreate}>
            <Plus aria-hidden="true" className="h-4 w-4" />
            Create cluster
          </Button>
        </div>
      </div>
    </div>
  );
}
