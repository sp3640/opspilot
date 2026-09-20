"use client";

import {
  CLUSTER_PROVIDER_LABELS,
  CLUSTER_PROVIDER_VALUES,
  CLUSTER_SORT_FIELDS,
  CLUSTER_STATUS_LABELS,
  CLUSTER_STATUS_VALUES,
} from "@/lib/constants/cluster";
import { controlClass, controlStyle } from "@/components/common/toolbar";
import type { ProjectResponse } from "@/types/project-api";

import type { ClusterFilters } from "./types";

type ClusterFilterProps = {
  filters: ClusterFilters;
  projects: ProjectResponse[];
  onFiltersChange: (filters: Partial<ClusterFilters>) => void;
};

/** Renders the filter controls only - Toolbar owns their grid layout. */
export function ClusterFilter({ filters, projects, onFiltersChange }: ClusterFilterProps) {
  return (
    <>
      <select aria-label="Provider" className={controlClass} style={controlStyle} value={filters.provider} onChange={(event) => onFiltersChange({ provider: event.target.value })}>
        <option value="all">All providers</option>
        {CLUSTER_PROVIDER_VALUES.map((provider) => (
          <option key={provider} value={provider}>
            {CLUSTER_PROVIDER_LABELS[provider]}
          </option>
        ))}
      </select>

      <select aria-label="Status" className={controlClass} style={controlStyle} value={filters.status} onChange={(event) => onFiltersChange({ status: event.target.value })}>
        <option value="all">All statuses</option>
        {CLUSTER_STATUS_VALUES.map((status) => (
          <option key={status} value={status}>
            {CLUSTER_STATUS_LABELS[status]}
          </option>
        ))}
      </select>

      <select aria-label="Project" className={controlClass} style={controlStyle} value={filters.projectId} onChange={(event) => onFiltersChange({ projectId: event.target.value })}>
        <option value="all">All projects</option>
        {projects.map((project) => (
          <option key={project.id} value={project.id}>
            {project.name}
          </option>
        ))}
      </select>

      <select aria-label="Sort" className={controlClass} style={controlStyle} value={filters.sort} onChange={(event) => onFiltersChange({ sort: event.target.value as ClusterFilters["sort"] })}>
        <option value={CLUSTER_SORT_FIELDS.UPDATED_AT}>Recently updated</option>
        <option value={CLUSTER_SORT_FIELDS.NAME}>Name</option>
        <option value={CLUSTER_SORT_FIELDS.PROVIDER}>Provider</option>
        <option value={CLUSTER_SORT_FIELDS.STATUS}>Status</option>
        <option value={CLUSTER_SORT_FIELDS.LAST_VALIDATED_AT}>Last validation</option>
        <option value={CLUSTER_SORT_FIELDS.LAST_DISCOVERY_AT}>Last discovery</option>
      </select>

      <select aria-label="Sort order" className={controlClass} style={controlStyle} value={filters.order} onChange={(event) => onFiltersChange({ order: event.target.value as ClusterFilters["order"] })}>
        <option value="desc">Descending</option>
        <option value="asc">Ascending</option>
      </select>
    </>
  );
}
