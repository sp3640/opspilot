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

export type ProjectMember = { name: string; initials: string };

export type Project = {
  id: string;
  name: string;
  description: string;
  health: ProjectHealth;
  environment: ProjectEnvironment;
  owner: ProjectMember;
  members: ProjectMember[];
  services: number;
  lastDeployment: string;
  updatedAt: string;
  deployments: number;
  icon: ProjectIconName;
};

export type ProjectFilters = {
  query: string;
  environment: ProjectEnvironment | "all";
  health: ProjectHealth | "all";
  sort: ProjectSort;
};

export type CreateProjectInput = {
  name: string;
  description: string;
  environment: ProjectEnvironment;
};
