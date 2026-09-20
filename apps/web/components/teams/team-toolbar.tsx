"use client";

import { Grid2X2, List, Plus, RefreshCw } from "lucide-react";

import { Toolbar, ToolbarSearch, ViewToggle, controlClass, controlStyle } from "@/components/common/toolbar";
import { Button } from "@/components/ui/button";
import { useIsPlatformAdmin } from "@/store/auth-store";

import type { TeamFilters, TeamOrder, TeamSort, TeamView } from "./types";

type TeamToolbarProps = {
  filters: TeamFilters;
  view: TeamView;
  onFiltersChange: (filters: Partial<TeamFilters>) => void;
  onViewChange: (view: TeamView) => void;
  onRefresh: () => void;
  onCreate: () => void;
  refreshing?: boolean;
};

export function TeamToolbar({ filters, view, onFiltersChange, onViewChange, onRefresh, onCreate, refreshing = false }: TeamToolbarProps) {
  const isAdmin = useIsPlatformAdmin();
  return (
    <Toolbar
      search={
        <ToolbarSearch
          id="team-search"
          label="Search teams"
          placeholder="Search teams"
          clearLabel="Clear team search"
          value={filters.query}
          onChange={(query) => onFiltersChange({ query })}
        />
      }
      actions={
        <>
          <ViewToggle
            label="Team view"
            value={view}
            onChange={onViewChange}
            options={[
              { value: "grid", label: "Grid", icon: Grid2X2 },
              { value: "table", label: "Table", icon: List },
            ]}
          />
          <Button type="button" variant="ghost" onClick={onRefresh} loading={refreshing} className="px-3" aria-label="Refresh teams">
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>
          {isAdmin ? (
            <Button type="button" onClick={onCreate}>
              <Plus aria-hidden="true" className="h-4 w-4" />
              Create team
            </Button>
          ) : null}
        </>
      }
      filters={
        <>
          <select aria-label="Sort teams" className={controlClass} style={controlStyle} value={filters.sort} onChange={(event) => onFiltersChange({ sort: event.target.value as TeamSort })}>
            <option value="updated_at">Recently updated</option>
            <option value="created_at">Recently created</option>
            <option value="name">Name</option>
          </select>
          <select aria-label="Sort order" className={controlClass} style={controlStyle} value={filters.order} onChange={(event) => onFiltersChange({ order: event.target.value as TeamOrder })}>
            <option value="desc">Descending</option>
            <option value="asc">Ascending</option>
          </select>
        </>
      }
    />
  );
}
