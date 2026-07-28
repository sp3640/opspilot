"use client";

import { Plus, RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";

import { ProjectFilter } from "./project-filter";
import { ProjectSearch } from "./project-search";
import { ProjectViewToggle } from "./project-view-toggle";
import type { ProjectFilters, ProjectView } from "./types";

type ProjectToolbarProps = { filters: ProjectFilters; view: ProjectView; onFiltersChange: (filters: Partial<ProjectFilters>) => void; onViewChange: (view: ProjectView) => void; onRefresh: () => void; onCreate: () => void };

/** Workspace controls, intentionally separated from content rendering. */
export function ProjectToolbar({ filters, view, onFiltersChange, onViewChange, onRefresh, onCreate }: ProjectToolbarProps) {
  return <div className="space-y-3"><div className="flex flex-col gap-3 lg:flex-row"><ProjectSearch value={filters.query} onChange={(query) => onFiltersChange({ query })} /><ProjectFilter environment={filters.environment} health={filters.health} sort={filters.sort} onEnvironmentChange={(environment) => onFiltersChange({ environment })} onHealthChange={(health) => onFiltersChange({ health })} onSortChange={(sort) => onFiltersChange({ sort })} /></div><div className="flex flex-wrap items-center justify-between gap-3"><ProjectViewToggle view={view} onChange={onViewChange} /><div className="flex items-center gap-2"><Button type="button" variant="ghost" onClick={onRefresh} className="px-3" aria-label="Refresh projects"><RefreshCw aria-hidden="true" className="h-4 w-4" /><span className="hidden sm:inline">Refresh</span></Button><Button type="button" onClick={onCreate}><Plus aria-hidden="true" className="h-4 w-4" />Create project</Button></div></div></div>;
}
