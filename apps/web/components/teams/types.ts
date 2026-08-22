import type { TeamResponse } from "@/types/team-api";

export type TeamView = "grid" | "table";

export type TeamSort = "name" | "created_at" | "updated_at";

export type TeamOrder = "asc" | "desc";

export type TeamFilters = {
  query: string;
  sort: TeamSort;
  order: TeamOrder;
};

export type Team = TeamResponse;
