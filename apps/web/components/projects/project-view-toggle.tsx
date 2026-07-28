"use client";

import { Grid2X2, List } from "lucide-react";

import { cn } from "@/lib/utils";

import type { ProjectView } from "./types";

/** Accessible switch between the project card and compact table representations. */
export function ProjectViewToggle({ view, onChange }: { view: ProjectView; onChange: (view: ProjectView) => void }) {
  return <div role="group" aria-label="Project view" className="flex rounded-2xl border p-1" style={{ borderColor: "var(--border)", backgroundColor: "var(--muted)" }}>{([{ value: "grid", label: "Grid", icon: Grid2X2 }, { value: "table", label: "Table", icon: List }] as const).map(({ value, label, icon: Icon }) => <button key={value} type="button" onClick={() => onChange(value)} aria-pressed={view === value} className={cn("inline-flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-medium transition-all", view === value ? "shadow-[var(--shadow-sm)]" : "hover:opacity-75")} style={{ backgroundColor: view === value ? "var(--card)" : "transparent", color: view === value ? "var(--foreground)" : "var(--muted-foreground)" }}><Icon aria-hidden="true" className="h-4 w-4" /><span className="sr-only sm:not-sr-only">{label}</span></button>)}</div>;
}
