import type { ResourceResponse } from "@/types/resource-api";
export type ResourceView = "grid" | "table";
export type ResourceSort = "created_at" | "updated_at" | "name" | "kind" | "status" | "health";
export type ResourceFilters = { query: string; projectId: string; cluster: string; namespace: string; provider: string; kind: string; status: string; health: string; region: string; sort: ResourceSort; order: "asc" | "desc" };
export type Resource = ResourceResponse;
