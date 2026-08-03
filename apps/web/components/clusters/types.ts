import type { ClusterResponse } from "@/types/cluster-api";

export type ClusterView = "grid" | "table";

export type ClusterSort =
  | "created_at"
  | "updated_at"
  | "name"
  | "provider"
  | "status"
  | "last_validated_at"
  | "last_discovery_at";

export type ClusterOrder = "asc" | "desc";

export type ClusterFilters = {
  query: string;
  provider: string;
  status: string;
  projectId: string;
  sort: ClusterSort;
  order: ClusterOrder;
};

export type Cluster = ClusterResponse;
