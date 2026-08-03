import type { AlertResponse } from "@/types/alert-api";
export type AlertView = "grid" | "table";
export type AlertFilters = { query: string; projectId: string; cluster: string; resource: string; severity: string; status: string; rule: string; createdDate: string; sort: "created_at" | "updated_at" | "severity" | "status" | "last_seen_at"; order: "asc" | "desc" };
export type Alert = AlertResponse;
