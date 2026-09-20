"use client";

import { Search } from "lucide-react";

import { useCommandPaletteStore } from "@/store/command-palette-store";

/**
 * Opens the shared CommandPalette (same one Cmd+K opens) rather than being
 * its own separate, disconnected search input - previously this rendered a
 * decorative <input> with no state, no handlers, and no way to submit a
 * query at all.
 */
export function SearchBar() {
  const openPalette = useCommandPaletteStore((state) => state.open);

  return (
    <button
      type="button"
      onClick={openPalette}
      aria-label="Search projects, incidents, and alerts"
      className="
        flex
        h-11
        w-full
        max-w-md
        items-center
        gap-3
        rounded-2xl
        border
        px-4
        text-left
        transition-all
        duration-300
        hover:border-[var(--primary-hover)]/40
      "
      style={{
        background: "var(--surface-hover)",
        borderColor: "var(--border)",
        color: "var(--surface-foreground)",
      }}
    >
      <Search aria-hidden="true" className="h-4 w-4 shrink-0" style={{ color: "var(--subtle-foreground)" }} />

      <span className="min-w-0 flex-1 truncate text-sm" style={{ color: "var(--subtle-foreground)" }}>
        Search projects, incidents, alerts...
      </span>

      <kbd
        className="
          shrink-0
          rounded-lg
          border
          px-2
          py-1
          text-[11px]
        "
        style={{ borderColor: "var(--border)", color: "var(--subtle-foreground)" }}
      >
        ⌘K
      </kbd>
    </button>
  );
}
