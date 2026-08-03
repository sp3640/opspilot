import { StatusBadge } from "@/components/common";
export function ResourceStatus({ status }: { status: string }) { const variant = status === "ACTIVE" ? "success" : status === "DELETED" ? "archived" : status === "PENDING" || status === "UPDATING" ? "warning" : status === "TERMINATING" ? "critical" : "info"; return <StatusBadge variant={variant}>{status}</StatusBadge>; }
