"use client";

import type { ReactNode } from "react";
import { Search, X } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

/** Shared look for every filter input/select so heights, radii and focus
 * states match across all list pages (previously each toolbar had its own
 * copy: h-11 / rounded-2xl / fixed widths that overflowed and overlapped). */
export const controlClass =
  "h-10 w-full min-w-0 rounded-xl border bg-transparent px-3 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";
export const controlStyle = { borderColor: "var(--input-border)", color: "var(--foreground)" } as const;

type ToolbarProps = {
  /** The search field. Takes all remaining width on the first row. */
  search?: ReactNode;
  /** View toggle + buttons. Sit to the right of search, wrapping below it on narrow screens. */
  actions?: ReactNode;
  /** Filter controls. Laid out in a wrapping grid on their own row. */
  filters?: ReactNode;
  className?: string;
};

/**
 * One layout for every list-page toolbar. The old per-page layout put the
 * search box and the whole filter group in a single non-wrapping row, so a
 * wide filter group squeezed the search field to nothing and controls drew
 * on top of each other. Here search + actions are row one, filters wrap in
 * a grid on row two, and nothing has a fixed width that can overflow.
 */
export function Toolbar({ search, actions, filters, className }: ToolbarProps) {
  return (
    <div className={cn("space-y-3", className)}>
      {(search || actions) && (
        <div className="flex flex-col gap-3 md:flex-row md:items-center">
          {search ? <div className="min-w-0 md:flex-1">{search}</div> : null}
          {actions ? <div className="flex flex-wrap items-center gap-2 md:shrink-0 md:justify-end">{actions}</div> : null}
        </div>
      )}
      {filters ? <div className="grid grid-cols-1 gap-2 sm:grid-cols-[repeat(auto-fill,minmax(11rem,1fr))]">{filters}</div> : null}
    </div>
  );
}

type ToolbarSearchProps = {
  id: string;
  label: string;
  placeholder: string;
  clearLabel: string;
  value: string;
  onChange: (value: string) => void;
};

export function ToolbarSearch({ id, label, placeholder, clearLabel, value, onChange }: ToolbarSearchProps) {
  return (
    <div className="relative min-w-0">
      <Search
        aria-hidden="true"
        className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2"
        style={{ color: "var(--muted-foreground)" }}
      />
      <label htmlFor={id} className="sr-only">
        {label}
      </label>
      <input
        id={id}
        type="text"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        className={cn(controlClass, "pl-10 pr-9")}
        style={controlStyle}
      />
      {value ? (
        <button
          type="button"
          onClick={() => onChange("")}
          aria-label={clearLabel}
          className="absolute right-2 top-1/2 flex h-6 w-6 -translate-y-1/2 items-center justify-center rounded-md transition-colors hover:bg-[var(--muted)]"
          style={{ color: "var(--muted-foreground)" }}
        >
          <X aria-hidden="true" className="h-4 w-4" />
        </button>
      ) : null}
    </div>
  );
}

type ViewOption<T extends string> = { value: T; label: string; icon: LucideIcon };

/** Segmented Grid/Table (etc.) switch, shared by every page that has one. */
export function ViewToggle<T extends string>({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: T;
  onChange: (value: T) => void;
  options: readonly ViewOption<T>[];
}) {
  return (
    <div
      role="group"
      aria-label={label}
      className="inline-flex h-10 shrink-0 items-center gap-0.5 rounded-xl border p-0.5"
      style={{ borderColor: "var(--border)", backgroundColor: "var(--muted)" }}
    >
      {options.map(({ value: optionValue, label: optionLabel, icon: Icon }) => {
        const active = value === optionValue;
        return (
          <button
            key={optionValue}
            type="button"
            onClick={() => onChange(optionValue)}
            aria-pressed={active}
            aria-label={optionLabel}
            className={cn(
              "inline-flex h-full items-center gap-2 rounded-[10px] px-3 text-sm font-medium transition-colors",
              active ? "shadow-[var(--shadow-sm)]" : "hover:bg-[var(--surface-hover)]"
            )}
            style={{
              backgroundColor: active ? "var(--card)" : "transparent",
              color: active ? "var(--foreground)" : "var(--muted-foreground)",
            }}
          >
            <Icon aria-hidden="true" className="h-4 w-4 shrink-0" />
            <span className="hidden sm:inline">{optionLabel}</span>
          </button>
        );
      })}
    </div>
  );
}
