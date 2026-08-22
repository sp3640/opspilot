import type { InvitationResponse } from "@/types/invitation-api";

export type InvitationView = "grid" | "table";

export type InvitationSort = "created_at" | "updated_at" | "email" | "expires_at" | "status";

export type InvitationOrder = "asc" | "desc";

export type InvitationStatusFilter = "all" | "Pending" | "Accepted" | "Expired" | "Revoked";

export type InvitationFilters = {
  query: string;
  status: InvitationStatusFilter;
  sort: InvitationSort;
  order: InvitationOrder;
};

export type Invitation = InvitationResponse;
