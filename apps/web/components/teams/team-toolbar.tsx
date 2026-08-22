"use client";

import { Grid2X2, List, Plus, RefreshCw, Search, SlidersHorizontal, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useIsPlatformAdmin } from "@/store/auth-store";

import type { TeamFilters, TeamOrder, TeamSort, TeamView } from "./types";

type TeamToolbarProps = {
  filters: TeamFilters;
  view: TeamView;
  onFiltersChange: (filters: Partial<TeamFilters>) => void;
  onViewChange: (view: TeamView) => void;
  onRefresh: () => void;
  onCreate: () => void;
  refreshing?: boolean;
};

const selectClass =
  "h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]";
const selectStyle = { borderColor: "var(--border)", color: "var(--foreground)" };

export function TeamToolbar({
  filters,
  view,
  onFiltersChange,
  onViewChange,
  onRefresh,
  onCreate,
  refreshing = false,
}: TeamToolbarProps) {
  const isAdmin = useIsPlatformAdmin();
  return (
    <div className="space-y-3">
      <div className="flex flex-col gap-3 lg:flex-row">
        <div className="relative min-w-0 flex-1">
          <Search
            aria-hidden="true"
            className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2"
            style={{ color: "var(--muted-foreground)" }}
          />
          <label htmlFor="team-search" className="sr-only">
            Search teams
          </label>
          <input
            id="team-search"
            value={filters.query}
            onChange={(event) => onFiltersChange({ query: event.target.value })}
            placeholder="Search teams"
            className="h-11 w-full rounded-2xl border bg-transparent py-2 pl-10 pr-9 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]"
            style={{ borderColor: "var(--border)" }}
          />
          {filters.query && (
            <button
              type="button"
              onClick={() => onFiltersChange({ query: "" })}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-1 hover:bg-[var(--muted)]"
              aria-label="Clear team search"
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

          <label className="sr-only" htmlFor="team-sort">
            Sort teams
          </label>
          <select
            id="team-sort"
            className={selectClass}
            style={selectStyle}
            value={filters.sort}
            onChange={(event) => onFiltersChange({ sort: event.target.value as TeamSort })}
          >
            <option value="updated_at">Recently updated</option>
            <option value="created_at">Recently created</option>
            <option value="name">Name</option>
          </select>

          <label className="sr-only" htmlFor="team-order">
            Sort order
          </label>
          <select
            id="team-order"
            className={selectClass}
            style={selectStyle}
            value={filters.order}
            onChange={(event) => onFiltersChange({ order: event.target.value as TeamOrder })}
          >
            <option value="desc">Descending</option>
            <option value="asc">Ascending</option>
          </select>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <div
          role="group"
          aria-label="Team view"
          className="flex rounded-2xl border p-1"
          style={{ borderColor: "var(--border)", backgroundColor: "var(--muted)" }}
        >
          {(
            [
              { value: "grid", label: "Grid", icon: Grid2X2 },
              { value: "table", label: "Table", icon: List },
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

        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="ghost"
            onClick={onRefresh}
            loading={refreshing}
            className="px-3"
            aria-label="Refresh teams"
          >
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>
          {isAdmin ? (
            <Button type="button" onClick={onCreate}>
              <Plus aria-hidden="true" className="h-4 w-4" />
              Create team
            </Button>
          ) : null}
        </div>
      </div>
    </div>
  );
}
