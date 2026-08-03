import { StatusBadge } from "@/components/common";
export function AlertSeverityBadge({ severity }: { severity: string }) { return <StatusBadge variant={severity === "CRITICAL" ? "critical" : severity === "HIGH" ? "warning" : severity === "MEDIUM" ? "info" : "healthy"}>{severity}</StatusBadge>; }
