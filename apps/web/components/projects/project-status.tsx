import { StatusBadge } from "@/components/common";

import type { ProjectEnvironment, ProjectHealth } from "./types";

const healthLabels: Record<ProjectHealth, string> = { healthy: "Operational", warning: "At risk", critical: "Critical" };
const environmentLabels: Record<ProjectEnvironment, string> = { production: "Production", staging: "Staging", development: "Development" };

/** Standard project health and environment labels built on shared status primitives. */
export function ProjectHealthBadge({ health }: { health: ProjectHealth }) {
  return <StatusBadge variant={health}>{healthLabels[health]}</StatusBadge>;
}

export function ProjectEnvironmentBadge({ environment }: { environment: ProjectEnvironment }) {
  return <StatusBadge variant={environment === "production" ? "info" : "archived"}>{environmentLabels[environment]}</StatusBadge>;
}
