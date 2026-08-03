"use client";

import { Search, X } from "lucide-react";

type ClusterSearchProps = {
  value: string;
  onChange: (value: string) => void;
};

export function ClusterSearch({ value, onChange }: ClusterSearchProps) {
  return (
    <div className="relative min-w-0 flex-1">
      <Search
        aria-hidden="true"
        className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2"
        style={{ color: "var(--muted-foreground)" }}
      />
      <label htmlFor="cluster-search" className="sr-only">
        Search clusters
      </label>
      <input
        id="cluster-search"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder="Search clusters by name, endpoint, provider, or region"
        className="h-11 w-full rounded-2xl border bg-transparent py-2 pl-10 pr-9 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]"
        style={{ borderColor: "var(--border)" }}
      />
      {value ? (
        <button
          type="button"
          onClick={() => onChange("")}
          className="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-1 hover:bg-[var(--muted)]"
          aria-label="Clear cluster search"
        >
          <X aria-hidden="true" className="h-4 w-4" />
        </button>
      ) : null}
    </div>
  );
}
