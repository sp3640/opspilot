import type { ProjectResponse } from "@/types/project-api";

export type ProjectHealth = "healthy" | "warning" | "critical";
export type ProjectEnvironment = "production" | "staging" | "development";
export type ProjectView = "grid" | "table";
export type ProjectSort = "updated" | "name" | "health";
export type ProjectIconName =
  | "boxes"
  | "cloud-cog"
  | "cpu"
  | "database"
  | "folder-kanban"
  | "globe"
  | "layers"
  | "network"
  | "shield-check";

export type Project = ProjectResponse;

export type ProjectFilters = {
  query: string;
  environment: ProjectEnvironment | "all";
  health: ProjectHealth | "all";
  sort: ProjectSort;
};
