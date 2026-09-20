"use client";

import { Grid2X2, List, UserPlus } from "lucide-react";

import { Toolbar, ToolbarSearch, ViewToggle, controlClass, controlStyle } from "@/components/common/toolbar";
import { Button } from "@/components/ui/button";

import type { InvitationFilters, InvitationOrder, InvitationSort, InvitationStatusFilter, InvitationView } from "./types";

type InvitationToolbarProps = {
  filters: InvitationFilters;
  view: InvitationView;
  onFiltersChange: (filters: Partial<InvitationFilters>) => void;
  onViewChange: (view: InvitationView) => void;
  onInvite: () => void;
};

export function InvitationToolbar({ filters, view, onFiltersChange, onViewChange, onInvite }: InvitationToolbarProps) {
  return (
    <Toolbar
      search={
        <ToolbarSearch
          id="invitation-search"
          label="Search invitations"
          placeholder="Search by email"
          clearLabel="Clear invitation search"
          value={filters.query}
          onChange={(query) => onFiltersChange({ query })}
        />
      }
      actions={
        <>
          <ViewToggle
            label="Invitation view"
            value={view}
            onChange={onViewChange}
            options={[
              { value: "table", label: "Table", icon: List },
              { value: "grid", label: "Grid", icon: Grid2X2 },
            ]}
          />
          <Button type="button" onClick={onInvite}>
            <UserPlus aria-hidden="true" className="h-4 w-4" />
            Invite member
          </Button>
        </>
      }
      filters={
        <>
          <select aria-label="Filter by status" className={controlClass} style={controlStyle} value={filters.status} onChange={(event) => onFiltersChange({ status: event.target.value as InvitationStatusFilter })}>
            <option value="all">All statuses</option>
            <option value="Pending">Pending</option>
            <option value="Accepted">Accepted</option>
            <option value="Expired">Expired</option>
            <option value="Revoked">Revoked</option>
          </select>
          <select aria-label="Sort invitations" className={controlClass} style={controlStyle} value={filters.sort} onChange={(event) => onFiltersChange({ sort: event.target.value as InvitationSort })}>
            <option value="created_at">Recently invited</option>
            <option value="updated_at">Recently updated</option>
            <option value="expires_at">Expiry date</option>
            <option value="email">Email</option>
            <option value="status">Status</option>
          </select>
          <select aria-label="Sort order" className={controlClass} style={controlStyle} value={filters.order} onChange={(event) => onFiltersChange({ order: event.target.value as InvitationOrder })}>
            <option value="desc">Descending</option>
            <option value="asc">Ascending</option>
          </select>
        </>
      }
    />
  );
}
