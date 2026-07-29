import { StatusBadge } from "@/components/common";

import type { ProjectHealth } from "./types";

const healthLabels: Record<string, string> = { healthy: "Operational", warning: "At risk", critical: "Critical" };
const environmentLabels: Record<string, string> = { production: "Production", staging: "Staging", development: "Development" };

/** Standard project health and environment labels built on shared status primitives. */
export function ProjectHealthBadge({ health }: { health: string }) {
  return <StatusBadge variant={health as ProjectHealth}>{healthLabels[health] ?? health}</StatusBadge>;
}

export function ProjectEnvironmentBadge({ environment }: { environment: string }) {
  return <StatusBadge variant={environment === "production" ? "info" : "archived"}>{environmentLabels[environment] ?? environment}</StatusBadge>;
}
