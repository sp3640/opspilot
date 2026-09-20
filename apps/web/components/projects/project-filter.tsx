"use client";

import { controlClass, controlStyle } from "@/components/common/toolbar";

import type { ProjectEnvironment, ProjectHealth, ProjectSort } from "./types";

type ProjectFilterProps = {
  environment: ProjectEnvironment | "all";
  health: ProjectHealth | "all";
  sort: ProjectSort;
  onEnvironmentChange: (value: ProjectEnvironment | "all") => void;
  onHealthChange: (value: ProjectHealth | "all") => void;
  onSortChange: (value: ProjectSort) => void;
};

/** Native selects for project environment, health, and ordering - Toolbar owns their grid layout. */
export function ProjectFilter({ environment, health, sort, onEnvironmentChange, onHealthChange, onSortChange }: ProjectFilterProps) {
  return (
    <>
      <select aria-label="Environment" className={controlClass} style={controlStyle} value={environment} onChange={(event) => onEnvironmentChange(event.target.value as ProjectEnvironment | "all")}>
        <option value="all">All environments</option>
        <option value="production">Production</option>
        <option value="staging">Staging</option>
        <option value="development">Development</option>
      </select>
      <select aria-label="Health" className={controlClass} style={controlStyle} value={health} onChange={(event) => onHealthChange(event.target.value as ProjectHealth | "all")}>
        <option value="all">All health</option>
        <option value="healthy">Operational</option>
        <option value="warning">At risk</option>
        <option value="critical">Critical</option>
      </select>
      <select aria-label="Sort projects" className={controlClass} style={controlStyle} value={sort} onChange={(event) => onSortChange(event.target.value as ProjectSort)}>
        <option value="updated">Recently updated</option>
        <option value="name">Name</option>
        <option value="health">Health</option>
      </select>
    </>
  );
}
