"use client";

import { SlidersHorizontal } from "lucide-react";

import type { ProjectEnvironment, ProjectHealth, ProjectSort } from "./types";

type ProjectFilterProps = {
  environment: ProjectEnvironment | "all";
  health: ProjectHealth | "all";
  sort: ProjectSort;
  onEnvironmentChange: (value: ProjectEnvironment | "all") => void;
  onHealthChange: (value: ProjectHealth | "all") => void;
  onSortChange: (value: ProjectSort) => void;
};

/** Compact native selects for project environment, health, and ordering. */
export function ProjectFilter({ environment, health, sort, onEnvironmentChange, onHealthChange, onSortChange }: ProjectFilterProps) {
  const selectClass = "h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]";
  const selectStyle = { borderColor: "var(--border)", color: "var(--foreground)" };
  return <div className="flex flex-wrap items-center gap-2"><SlidersHorizontal aria-hidden="true" className="hidden h-4 w-4 lg:block" style={{ color: "var(--muted-foreground)" }} /><label className="sr-only" htmlFor="environment-filter">Environment</label><select id="environment-filter" className={selectClass} style={selectStyle} value={environment} onChange={(event) => onEnvironmentChange(event.target.value as ProjectEnvironment | "all")}><option value="all">All environments</option><option value="production">Production</option><option value="staging">Staging</option><option value="development">Development</option></select><label className="sr-only" htmlFor="health-filter">Health</label><select id="health-filter" className={selectClass} style={selectStyle} value={health} onChange={(event) => onHealthChange(event.target.value as ProjectHealth | "all")}><option value="all">All health</option><option value="healthy">Operational</option><option value="warning">At risk</option><option value="critical">Critical</option></select><label className="sr-only" htmlFor="sort-filter">Sort projects</label><select id="sort-filter" className={selectClass} style={selectStyle} value={sort} onChange={(event) => onSortChange(event.target.value as ProjectSort)}><option value="updated">Recently updated</option><option value="name">Name</option><option value="health">Health</option></select></div>;
}
