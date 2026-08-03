"use client";
import { SlidersHorizontal } from "lucide-react";
import { RESOURCE_HEALTH_VALUES, RESOURCE_KIND_VALUES, RESOURCE_SORT_FIELDS, RESOURCE_STATUS_VALUES } from "@/lib/constants/resource";
import type { ClusterResponse } from "@/types/cluster-api";
import type { ProjectResponse } from "@/types/project-api";
import type { ResourceFilters } from "./types";
const selectClass = "h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)]";
export function ResourceFilter({ filters, projects, clusters, onFiltersChange }: { filters: ResourceFilters; projects: ProjectResponse[]; clusters: ClusterResponse[]; onFiltersChange: (filters: Partial<ResourceFilters>) => void }) {
 const input = (key: "provider" | "namespace" | "region", label: string) => <input aria-label={label} value={filters[key] === "all" ? "" : filters[key]} onChange={(event) => onFiltersChange({ [key]: event.target.value || "all" })} placeholder={label} className={selectClass + " w-32"} style={{ borderColor: "var(--border)" }} />;
 const select = (key: keyof ResourceFilters, label: string, options: readonly string[]) => <select aria-label={label} value={String(filters[key])} onChange={(event) => onFiltersChange({ [key]: event.target.value } as Partial<ResourceFilters>)} className={selectClass} style={{ borderColor: "var(--border)" }}><option value="all">All {label.toLowerCase()}s</option>{options.map((option) => <option key={option} value={option}>{option}</option>)}</select>;
 return <div className="flex flex-wrap items-center gap-2"><SlidersHorizontal aria-hidden className="hidden h-4 w-4 lg:block" style={{ color: "var(--muted-foreground)" }} />
   <select aria-label="Project" value={filters.projectId} onChange={(event) => onFiltersChange({ projectId: event.target.value })} className={selectClass} style={{ borderColor: "var(--border)" }}><option value="all">All projects</option>{projects.map((project) => <option key={project.id} value={project.id}>{project.name}</option>)}</select>
   <select aria-label="Cluster" value={filters.cluster} onChange={(event) => onFiltersChange({ cluster: event.target.value })} className={selectClass} style={{ borderColor: "var(--border)" }}><option value="all">All clusters</option>{clusters.map((cluster) => <option key={cluster.id} value={cluster.name}>{cluster.name}</option>)}</select>
   {select("kind", "Kind", RESOURCE_KIND_VALUES)} {select("status", "Status", RESOURCE_STATUS_VALUES)} {select("health", "Health", RESOURCE_HEALTH_VALUES)} {input("namespace", "Namespace")} {input("provider", "Provider")} {input("region", "Region")}
   <select aria-label="Sort resources" value={filters.sort} onChange={(event) => onFiltersChange({ sort: event.target.value as ResourceFilters["sort"] })} className={selectClass} style={{ borderColor: "var(--border)" }}><option value={RESOURCE_SORT_FIELDS.UPDATED_AT}>Recently updated</option><option value={RESOURCE_SORT_FIELDS.NAME}>Name</option><option value={RESOURCE_SORT_FIELDS.KIND}>Kind</option><option value={RESOURCE_SORT_FIELDS.STATUS}>Status</option><option value={RESOURCE_SORT_FIELDS.HEALTH}>Health</option><option value={RESOURCE_SORT_FIELDS.CREATED_AT}>Created</option></select>
   <select aria-label="Sort order" value={filters.order} onChange={(event) => onFiltersChange({ order: event.target.value as ResourceFilters["order"] })} className={selectClass} style={{ borderColor: "var(--border)" }}><option value="desc">Descending</option><option value="asc">Ascending</option></select>
 </div>;
}
