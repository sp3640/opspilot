"use client";

import { Grid2X2, List, Search, SlidersHorizontal, UserPlus, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

import type { InvitationFilters, InvitationOrder, InvitationSort, InvitationStatusFilter, InvitationView } from "./types";

type InvitationToolbarProps = {
  filters: InvitationFilters;
  view: InvitationView;
  onFiltersChange: (filters: Partial<InvitationFilters>) => void;
  onViewChange: (view: InvitationView) => void;
  onInvite: () => void;
};

const selectClass =
  "h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]";
const selectStyle = { borderColor: "var(--border)", color: "var(--foreground)" };

export function InvitationToolbar({ filters, view, onFiltersChange, onViewChange, onInvite }: InvitationToolbarProps) {
  return (
    <div className="space-y-3">
      <div className="flex flex-col gap-3 lg:flex-row">
        <div className="relative min-w-0 flex-1">
          <Search
            aria-hidden="true"
            className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2"
            style={{ color: "var(--muted-foreground)" }}
          />
          <label htmlFor="invitation-search" className="sr-only">
            Search invitations
          </label>
          <input
            id="invitation-search"
            value={filters.query}
            onChange={(event) => onFiltersChange({ query: event.target.value })}
            placeholder="Search by email"
            className="h-11 w-full rounded-2xl border bg-transparent py-2 pl-10 pr-9 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]"
            style={{ borderColor: "var(--border)" }}
          />
          {filters.query && (
            <button
              type="button"
              onClick={() => onFiltersChange({ query: "" })}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-1 hover:bg-[var(--muted)]"
              aria-label="Clear invitation search"
            >
              <X aria-hidden="true" className="h-4 w-4" />
            </button>
          )}
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <SlidersHorizontal
            aria-hidden="true"
            className="hidden h-4 w-4 lg:block"
            style={{ color: "var(--muted-foreground)" }}
          />

          <label className="sr-only" htmlFor="invitation-status">
            Filter by status
          </label>
          <select
            id="invitation-status"
            className={selectClass}
            style={selectStyle}
            value={filters.status}
            onChange={(event) => onFiltersChange({ status: event.target.value as InvitationStatusFilter })}
          >
            <option value="all">All statuses</option>
            <option value="Pending">Pending</option>
            <option value="Accepted">Accepted</option>
            <option value="Expired">Expired</option>
            <option value="Revoked">Revoked</option>
          </select>

          <label className="sr-only" htmlFor="invitation-sort">
            Sort invitations
          </label>
          <select
            id="invitation-sort"
            className={selectClass}
            style={selectStyle}
            value={filters.sort}
            onChange={(event) => onFiltersChange({ sort: event.target.value as InvitationSort })}
          >
            <option value="created_at">Recently invited</option>
            <option value="updated_at">Recently updated</option>
            <option value="expires_at">Expiry date</option>
            <option value="email">Email</option>
            <option value="status">Status</option>
          </select>

          <label className="sr-only" htmlFor="invitation-order">
            Sort order
          </label>
          <select
            id="invitation-order"
            className={selectClass}
            style={selectStyle}
            value={filters.order}
            onChange={(event) => onFiltersChange({ order: event.target.value as InvitationOrder })}
          >
            <option value="desc">Descending</option>
            <option value="asc">Ascending</option>
          </select>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <div
          role="group"
          aria-label="Invitation view"
          className="flex rounded-2xl border p-1"
          style={{ borderColor: "var(--border)", backgroundColor: "var(--muted)" }}
        >
          {(
            [
              { value: "table", label: "Table", icon: List },
              { value: "grid", label: "Grid", icon: Grid2X2 },
            ] as const
          ).map(({ value, label, icon: Icon }) => (
            <button
              key={value}
              type="button"
              onClick={() => onViewChange(value)}
              aria-pressed={view === value}
              className={cn(
                "inline-flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-medium transition-all",
                view === value ? "shadow-[var(--shadow-sm)]" : "hover:opacity-75"
              )}
              style={{
                backgroundColor: view === value ? "var(--card)" : "transparent",
                color: view === value ? "var(--foreground)" : "var(--muted-foreground)",
              }}
            >
              <Icon aria-hidden="true" className="h-4 w-4" />
              <span className="sr-only sm:not-sr-only">{label}</span>
            </button>
          ))}
        </div>

        <Button type="button" onClick={onInvite}>
          <UserPlus aria-hidden="true" className="h-4 w-4" />
          Invite member
        </Button>
      </div>
    </div>
  );
}
