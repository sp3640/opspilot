"use client";

import { RESOURCE_HEALTH_VALUES, RESOURCE_KIND_VALUES, RESOURCE_SORT_FIELDS, RESOURCE_STATUS_VALUES } from "@/lib/constants/resource";
import { controlClass, controlStyle } from "@/components/common/toolbar";
import type { ClusterResponse } from "@/types/cluster-api";
import type { ProjectResponse } from "@/types/project-api";

import type { ResourceFilters } from "./types";

const ALL_LABELS: Record<string, string> = { Kind: "kinds", Status: "statuses", Health: "health states", Severity: "severities" };

type ResourceFilterProps = {
  filters: ResourceFilters;
  projects: ProjectResponse[];
  clusters: ClusterResponse[];
  onFiltersChange: (filters: Partial<ResourceFilters>) => void;
};

/** Renders the filter controls only - Toolbar owns their grid layout. */
export function ResourceFilter({ filters, projects, clusters, onFiltersChange }: ResourceFilterProps) {
  const textInput = (key: "provider" | "namespace" | "region", label: string) => (
    <input
      aria-label={label}
      value={filters[key] === "all" ? "" : filters[key]}
      onChange={(event) => onFiltersChange({ [key]: event.target.value || "all" })}
      placeholder={label}
      className={controlClass}
      style={controlStyle}
    />
  );

  const select = (key: "kind" | "status" | "health", label: string, options: readonly string[]) => (
    <select
      aria-label={label}
      value={String(filters[key])}
      onChange={(event) => onFiltersChange({ [key]: event.target.value } as Partial<ResourceFilters>)}
      className={controlClass}
      style={controlStyle}
    >
      <option value="all">All {ALL_LABELS[label] ?? label.toLowerCase()}</option>
      {options.map((option) => (
        <option key={option} value={option}>
          {option}
        </option>
      ))}
    </select>
  );

  return (
    <>
      <select aria-label="Project" value={filters.projectId} onChange={(event) => onFiltersChange({ projectId: event.target.value })} className={controlClass} style={controlStyle}>
        <option value="all">All projects</option>
        {projects.map((project) => (
          <option key={project.id} value={project.id}>
            {project.name}
          </option>
        ))}
      </select>
      <select aria-label="Cluster" value={filters.cluster} onChange={(event) => onFiltersChange({ cluster: event.target.value })} className={controlClass} style={controlStyle}>
        <option value="all">All clusters</option>
        {clusters.map((cluster) => (
          <option key={cluster.id} value={cluster.name}>
            {cluster.name}
          </option>
        ))}
      </select>
      {select("kind", "Kind", RESOURCE_KIND_VALUES)}
      {select("status", "Status", RESOURCE_STATUS_VALUES)}
      {select("health", "Health", RESOURCE_HEALTH_VALUES)}
      {textInput("namespace", "Namespace")}
      {textInput("provider", "Provider")}
      {textInput("region", "Region")}
      <select aria-label="Sort resources" value={filters.sort} onChange={(event) => onFiltersChange({ sort: event.target.value as ResourceFilters["sort"] })} className={controlClass} style={controlStyle}>
        <option value={RESOURCE_SORT_FIELDS.UPDATED_AT}>Recently updated</option>
        <option value={RESOURCE_SORT_FIELDS.NAME}>Name</option>
        <option value={RESOURCE_SORT_FIELDS.KIND}>Kind</option>
        <option value={RESOURCE_SORT_FIELDS.STATUS}>Status</option>
        <option value={RESOURCE_SORT_FIELDS.HEALTH}>Health</option>
        <option value={RESOURCE_SORT_FIELDS.CREATED_AT}>Created</option>
      </select>
      <select aria-label="Sort order" value={filters.order} onChange={(event) => onFiltersChange({ order: event.target.value as ResourceFilters["order"] })} className={controlClass} style={controlStyle}>
        <option value="desc">Descending</option>
        <option value="asc">Ascending</option>
      </select>
    </>
  );
}
