"use client";

import { METRIC_TYPES } from "@/lib/constants/metric";
import { controlClass, controlStyle } from "@/components/common/toolbar";
import type { ClusterResponse } from "@/types/cluster-api";
import type { ProjectResponse } from "@/types/project-api";

import type { MetricFilters } from "./types";

type MetricsFilterProps = {
  filters: MetricFilters;
  projects: ProjectResponse[];
  clusters: ClusterResponse[];
  onChange: (x: Partial<MetricFilters>) => void;
};

/** Renders the filter controls only - Toolbar owns their grid layout. */
export function MetricsFilter({ filters, projects, clusters, onChange }: MetricsFilterProps) {
  return (
    <>
      <select aria-label="Project" value={filters.projectId} onChange={(event) => onChange({ projectId: event.target.value })} className={controlClass} style={controlStyle}>
        <option value="all">All projects</option>
        {projects.map((project) => (
          <option key={project.id} value={project.id}>
            {project.name}
          </option>
        ))}
      </select>
      <select aria-label="Cluster" value={filters.clusterId} onChange={(event) => onChange({ clusterId: event.target.value })} className={controlClass} style={controlStyle}>
        <option value="all">All clusters</option>
        {clusters.map((cluster) => (
          <option key={cluster.id} value={cluster.id}>
            {cluster.name}
          </option>
        ))}
      </select>
      <input
        aria-label="Resource ID"
        value={filters.resourceId === "all" ? "" : filters.resourceId}
        onChange={(event) => onChange({ resourceId: event.target.value || "all" })}
        placeholder="Resource ID"
        className={controlClass}
        style={controlStyle}
      />
      <select aria-label="Metric type" value={filters.metricType} onChange={(event) => onChange({ metricType: event.target.value })} className={controlClass} style={controlStyle}>
        <option value="all">All types</option>
        {METRIC_TYPES.map((type) => (
          <option key={type}>{type}</option>
        ))}
      </select>
      <select aria-label="Time range" value={filters.timeRange} onChange={(event) => onChange({ timeRange: event.target.value as MetricFilters["timeRange"] })} className={controlClass} style={controlStyle}>
        <option value="24h">Last 24 hours</option>
        <option value="7d">Last 7 days</option>
        <option value="30d">Last 30 days</option>
      </select>
    </>
  );
}
