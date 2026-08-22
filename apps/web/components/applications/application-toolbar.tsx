"use client";

import { Grid2X2, List, Plus, Search, SlidersHorizontal, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useHasPermission } from "@/store/auth-store";

type ApplicationView = "grid" | "table";

type ApplicationToolbarProps = {
  search: string;
  onSearchChange: (value: string) => void;
  runtime?: string;
  onRuntimeChange?: (value: string) => void;
  status?: string;
  onStatusChange?: (value: string) => void;
  view: ApplicationView;
  onViewChange: (view: ApplicationView) => void;
  onCreate: () => void;
};

const RUNTIME_OPTIONS = ["NodeJS", "Go", "Java", "Python", "DotNet", "Static", "Docker"] as const;
const STATUS_OPTIONS = ["Draft", "Ready", "Archived"] as const;

const selectClass =
  "h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]";
const selectStyle = { borderColor: "var(--border)", color: "var(--foreground)" };

export function ApplicationToolbar({
  search,
  onSearchChange,
  runtime,
  onRuntimeChange,
  status,
  onStatusChange,
  view,
  onViewChange,
  onCreate,
}: ApplicationToolbarProps) {
  const canManageApplications = useHasPermission("application:manage");
  const showFilters = Boolean(onRuntimeChange || onStatusChange);

  return (
    <div className="space-y-3">
      <div className="flex flex-col gap-3 lg:flex-row">
        <div className="relative min-w-0 flex-1">
          <Search
            aria-hidden="true"
            className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2"
            style={{ color: "var(--muted-foreground)" }}
          />
          <label htmlFor="application-search" className="sr-only">
            Search applications
          </label>
          <input
            id="application-search"
            value={search}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Search applications"
            className="h-11 w-full rounded-2xl border bg-transparent py-2 pl-10 pr-9 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]"
            style={{ borderColor: "var(--border)" }}
          />
          {search && (
            <button
              type="button"
              onClick={() => onSearchChange("")}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-1 hover:bg-[var(--muted)]"
              aria-label="Clear application search"
            >
              <X aria-hidden="true" className="h-4 w-4" />
            </button>
          )}
        </div>

        {showFilters && (
          <div className="flex flex-wrap items-center gap-2">
            <SlidersHorizontal
              aria-hidden="true"
              className="hidden h-4 w-4 lg:block"
              style={{ color: "var(--muted-foreground)" }}
            />

            {onRuntimeChange && (
              <>
                <label className="sr-only" htmlFor="application-runtime-filter">
                  Runtime
                </label>
                <select
                  id="application-runtime-filter"
                  className={selectClass}
                  style={selectStyle}
                  value={runtime ?? "all"}
                  onChange={(event) => onRuntimeChange(event.target.value)}
                >
                  <option value="all">All runtimes</option>
                  {RUNTIME_OPTIONS.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
              </>
            )}

            {onStatusChange && (
              <>
                <label className="sr-only" htmlFor="application-status-filter">
                  Status
                </label>
                <select
                  id="application-status-filter"
                  className={selectClass}
                  style={selectStyle}
                  value={status ?? "all"}
                  onChange={(event) => onStatusChange(event.target.value)}
                >
                  <option value="all">All statuses</option>
                  {STATUS_OPTIONS.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
              </>
            )}
          </div>
        )}
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <div role="group" aria-label="Application view" className="flex rounded-2xl border p-1" style={{ borderColor: "var(--border)", backgroundColor: "var(--muted)" }}>
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
          {canManageApplications ? (
            <Button type="button" onClick={onCreate}>
              <Plus aria-hidden="true" className="h-4 w-4" />
              Create application
            </Button>
          ) : null}
        </div>
      </div>
    </div>
  );
}
