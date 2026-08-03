"use client";
import { Plus, RefreshCw, RotateCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { ClusterResponse } from "@/types/cluster-api";
import type { ProjectResponse } from "@/types/project-api";
import { ResourceFilter } from "./resource-filter";
import { ResourceSearch } from "./resource-search";
import { ResourceViewToggle } from "./resource-view-toggle";
import type { ResourceFilters, ResourceView } from "./types";
export function ResourceToolbar({ filters, view, projects, clusters, onFiltersChange, onViewChange, onRefresh, onCreate, onSync, refreshing, syncing }: { filters: ResourceFilters; view: ResourceView; projects: ProjectResponse[]; clusters: ClusterResponse[]; onFiltersChange: (f: Partial<ResourceFilters>) => void; onViewChange: (view: ResourceView) => void; onRefresh: () => void; onCreate: () => void; onSync: () => void; refreshing?: boolean; syncing?: boolean }) { return <div className="space-y-3"><div className="flex flex-col gap-3 lg:flex-row"><ResourceSearch value={filters.query} onChange={(query) => onFiltersChange({ query })} /><ResourceFilter filters={filters} projects={projects} clusters={clusters} onFiltersChange={onFiltersChange} /></div><div className="flex flex-wrap items-center justify-between gap-3"><ResourceViewToggle view={view} onChange={onViewChange} /><div className="flex items-center gap-2"><Button type="button" variant="ghost" onClick={onRefresh} loading={refreshing} className="px-3"><RefreshCw className="h-4 w-4" /><span className="hidden sm:inline">Refresh</span></Button><Button type="button" variant="secondary" onClick={onSync} loading={syncing}><RotateCw className="h-4 w-4" />Sync resources</Button><Button type="button" onClick={onCreate}><Plus className="h-4 w-4" />Create resource</Button></div></div></div>; }
