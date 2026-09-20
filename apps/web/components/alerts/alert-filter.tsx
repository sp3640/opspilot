"use client";

import { ALERT_SEVERITIES, ALERT_STATUSES } from "@/lib/constants/alert";
import { controlClass, controlStyle } from "@/components/common/toolbar";
import type { ClusterResponse } from "@/types/cluster-api";
import type { ProjectResponse } from "@/types/project-api";

import type { AlertFilters } from "./types";

const ALL_LABELS: Record<string, string> = { Kind: "kinds", Status: "statuses", Health: "health states", Severity: "severities" };

type AlertFilterProps = {
  filters: AlertFilters;
  projects: ProjectResponse[];
  clusters: ClusterResponse[];
  onChange: (value: Partial<AlertFilters>) => void;
};

/** Renders the filter controls only - Toolbar owns their grid layout. */
export function AlertFilter({ filters, projects, clusters, onChange }: AlertFilterProps) {
  const select = (key: "severity" | "status", title: string, options: readonly string[]) => (
    <select aria-label={title} value={filters[key]} onChange={(event) => onChange({ [key]: event.target.value })} className={controlClass} style={controlStyle}>
      <option value="all">All {ALL_LABELS[title] ?? title.toLowerCase()}</option>
      {options.map((option) => (
        <option key={option}>{option}</option>
      ))}
    </select>
  );

  const textInput = (key: "resource" | "rule", placeholder: string) => (
    <input
      aria-label={placeholder}
      value={filters[key] === "all" ? "" : filters[key]}
      onChange={(event) => onChange({ [key]: event.target.value || "all" })}
      placeholder={placeholder}
      className={controlClass}
      style={controlStyle}
    />
  );

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
      <select aria-label="Cluster" value={filters.cluster} onChange={(event) => onChange({ cluster: event.target.value })} className={controlClass} style={controlStyle}>
        <option value="all">All clusters</option>
        {clusters.map((cluster) => (
          <option key={cluster.id} value={cluster.name}>
            {cluster.name}
          </option>
        ))}
      </select>
      {textInput("resource", "Resource")}
      {select("severity", "Severity", ALERT_SEVERITIES)}
      {select("status", "Status", ALERT_STATUSES)}
      {textInput("rule", "Rule")}
      <input aria-label="Created date" type="date" value={filters.createdDate} onChange={(event) => onChange({ createdDate: event.target.value })} className={controlClass} style={controlStyle} />
      <select aria-label="Sort alerts" value={filters.sort} onChange={(event) => onChange({ sort: event.target.value as AlertFilters["sort"] })} className={controlClass} style={controlStyle}>
        <option value="updated_at">Recently updated</option>
        <option value="created_at">Created</option>
        <option value="severity">Severity</option>
        <option value="status">Status</option>
        <option value="last_seen_at">Last seen</option>
      </select>
    </>
  );
}
