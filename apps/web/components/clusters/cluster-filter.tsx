"use client";

import { SlidersHorizontal } from "lucide-react";
import {
  CLUSTER_PROVIDER_LABELS,
  CLUSTER_PROVIDER_VALUES,
  CLUSTER_SORT_FIELDS,
  CLUSTER_STATUS_LABELS,
  CLUSTER_STATUS_VALUES,
} from "@/lib/constants/cluster";
import type { ProjectResponse } from "@/types/project-api";

import type { ClusterFilters } from "./types";

type ClusterFilterProps = {
  filters: ClusterFilters;
  projects: ProjectResponse[];
  onFiltersChange: (filters: Partial<ClusterFilters>) => void;
};

export function ClusterFilter({ filters, projects, onFiltersChange }: ClusterFilterProps) {
  const selectClass =
    "h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]";
  const selectStyle = { borderColor: "var(--border)", color: "var(--foreground)" };

  return (
    <div className="flex flex-wrap items-center gap-2">
      <SlidersHorizontal
        aria-hidden="true"
        className="hidden h-4 w-4 lg:block"
        style={{ color: "var(--muted-foreground)" }}
      />

      <label className="sr-only" htmlFor="cluster-provider-filter">
        Provider
      </label>
      <select
        id="cluster-provider-filter"
        className={selectClass}
        style={selectStyle}
        value={filters.provider}
        onChange={(event) => onFiltersChange({ provider: event.target.value })}
      >
        <option value="all">All providers</option>
        {CLUSTER_PROVIDER_VALUES.map((provider) => (
          <option key={provider} value={provider}>
            {CLUSTER_PROVIDER_LABELS[provider]}
          </option>
        ))}
      </select>

      <label className="sr-only" htmlFor="cluster-status-filter">
        Status
      </label>
      <select
        id="cluster-status-filter"
        className={selectClass}
        style={selectStyle}
        value={filters.status}
        onChange={(event) => onFiltersChange({ status: event.target.value })}
      >
        <option value="all">All statuses</option>
        {CLUSTER_STATUS_VALUES.map((status) => (
          <option key={status} value={status}>
            {CLUSTER_STATUS_LABELS[status]}
          </option>
        ))}
      </select>

      <label className="sr-only" htmlFor="cluster-project-filter">
        Project
      </label>
      <select
        id="cluster-project-filter"
        className={selectClass}
        style={selectStyle}
        value={filters.projectId}
        onChange={(event) => onFiltersChange({ projectId: event.target.value })}
      >
        <option value="all">All projects</option>
        {projects.map((project) => (
          <option key={project.id} value={project.id}>
            {project.name}
          </option>
        ))}
      </select>

      <label className="sr-only" htmlFor="cluster-sort-filter">
        Sort
      </label>
      <select
        id="cluster-sort-filter"
        className={selectClass}
        style={selectStyle}
        value={filters.sort}
        onChange={(event) => onFiltersChange({ sort: event.target.value as ClusterFilters["sort"] })}
      >
        <option value={CLUSTER_SORT_FIELDS.UPDATED_AT}>Recently updated</option>
        <option value={CLUSTER_SORT_FIELDS.NAME}>Name</option>
        <option value={CLUSTER_SORT_FIELDS.PROVIDER}>Provider</option>
        <option value={CLUSTER_SORT_FIELDS.STATUS}>Status</option>
        <option value={CLUSTER_SORT_FIELDS.LAST_VALIDATED_AT}>Last validation</option>
        <option value={CLUSTER_SORT_FIELDS.LAST_DISCOVERY_AT}>Last discovery</option>
      </select>

      <label className="sr-only" htmlFor="cluster-order-filter">
        Sort order
      </label>
      <select
        id="cluster-order-filter"
        className={selectClass}
        style={selectStyle}
        value={filters.order}
        onChange={(event) => onFiltersChange({ order: event.target.value as ClusterFilters["order"] })}
      >
        <option value="desc">Descending</option>
        <option value="asc">Ascending</option>
      </select>
    </div>
  );
}
