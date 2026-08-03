import { StatusBadge } from "@/components/common";
export function AlertStatusBadge({ status }: { status: string }) { return <StatusBadge variant={status === "RESOLVED" ? "success" : status === "ACKNOWLEDGED" || status === "INVESTIGATING" ? "warning" : "info"}>{status}</StatusBadge>; }
