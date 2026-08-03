import { StatusBadge } from "@/components/common";
export function ResourceHealthBadge({ health }: { health: string }) { return <StatusBadge variant={health === "HEALTHY" ? "healthy" : health === "DEGRADED" ? "warning" : health === "UNHEALTHY" ? "critical" : "info"}>{health[0] + health.slice(1).toLowerCase()}</StatusBadge>; }
