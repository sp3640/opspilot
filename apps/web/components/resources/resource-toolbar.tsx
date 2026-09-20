"use client";

import { Plus, RefreshCw, RotateCw } from "lucide-react";

import { Toolbar } from "@/components/common/toolbar";
import { Button } from "@/components/ui/button";
import type { ClusterResponse } from "@/types/cluster-api";
import type { ProjectResponse } from "@/types/project-api";

import { ResourceFilter } from "./resource-filter";
import { ResourceSearch } from "./resource-search";
import { ResourceViewToggle } from "./resource-view-toggle";
import type { ResourceFilters, ResourceView } from "./types";

type ResourceToolbarProps = {
  filters: ResourceFilters;
  view: ResourceView;
  projects: ProjectResponse[];
  clusters: ClusterResponse[];
  onFiltersChange: (filters: Partial<ResourceFilters>) => void;
  onViewChange: (view: ResourceView) => void;
  onRefresh: () => void;
  onCreate: () => void;
  onSync: () => void;
  refreshing?: boolean;
  syncing?: boolean;
};

export function ResourceToolbar({
  filters,
  view,
  projects,
  clusters,
  onFiltersChange,
  onViewChange,
  onRefresh,
  onCreate,
  onSync,
  refreshing,
  syncing,
}: ResourceToolbarProps) {
  return (
    <Toolbar
      search={<ResourceSearch value={filters.query} onChange={(query) => onFiltersChange({ query })} />}
      actions={
        <>
          <ResourceViewToggle view={view} onChange={onViewChange} />
          <Button type="button" variant="ghost" onClick={onRefresh} loading={refreshing} className="px-3" aria-label="Refresh resources">
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden xl:inline">Refresh</span>
          </Button>
          <Button type="button" variant="secondary" onClick={onSync} loading={syncing}>
            <RotateCw aria-hidden="true" className="h-4 w-4" />
            Sync
          </Button>
          <Button type="button" onClick={onCreate}>
            <Plus aria-hidden="true" className="h-4 w-4" />
            Create resource
          </Button>
        </>
      }
      filters={<ResourceFilter filters={filters} projects={projects} clusters={clusters} onFiltersChange={onFiltersChange} />}
    />
  );
}
