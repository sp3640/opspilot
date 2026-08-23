"use client";

import Link from "next/link";
import { CheckCircle2, ChevronRight, ShieldAlert, TriangleAlert } from "lucide-react";

import { SectionCard } from "./section-card";
import type { ProblemItem } from "@/lib/dashboard-overview";

/**
 * Problems, not statistics, first: every real critical/warning finding
 * across applications/clusters/alerts/incidents/deployments, each linking
 * directly to where it can be investigated. Rendered even when empty - a
 * calm "nothing wrong" state is itself meaningful, never omitted silently.
 */
export function DashboardProblems({ criticalProblems, warnings }: { criticalProblems: ProblemItem[]; warnings: ProblemItem[] }) {
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <SectionCard title="Critical" description="Requires immediate attention.">
        <ProblemList items={criticalProblems} tone="critical" emptyLabel="No critical problems." />
      </SectionCard>
      <SectionCard title="Warnings" description="Worth investigating soon.">
        <ProblemList items={warnings} tone="warning" emptyLabel="No warnings." />
      </SectionCard>
    </div>
  );
}

function ProblemList({ items, tone, emptyLabel }: { items: ProblemItem[]; tone: "critical" | "warning"; emptyLabel: string }) {
  if (items.length === 0) {
    return (
      <div className="flex items-center gap-2 py-2 text-sm" style={{ color: "var(--muted-foreground)" }}>
        <CheckCircle2 aria-hidden="true" className="h-4 w-4" style={{ color: "var(--success)" }} />
        {emptyLabel}
      </div>
    );
  }

  const color = tone === "critical" ? "var(--danger)" : "var(--warning)";
  const Icon = tone === "critical" ? ShieldAlert : TriangleAlert;

  return (
    <ul className="space-y-2">
      {items.map((item) => (
        <li key={item.id}>
          <Link
            href={item.href}
            className="flex items-start gap-3 rounded-2xl border p-3 transition-colors hover:bg-[var(--muted)]"
            style={{ borderColor: "var(--border)" }}
          >
            <Icon aria-hidden="true" className="mt-0.5 h-4 w-4 shrink-0" style={{ color }} />
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{item.title}</p>
              <p className="mt-0.5 text-xs" style={{ color: "var(--muted-foreground)" }}>
                {item.description}
              </p>
            </div>
            <ChevronRight aria-hidden="true" className="mt-0.5 h-4 w-4 shrink-0" style={{ color: "var(--muted-foreground)" }} />
          </Link>
        </li>
      ))}
    </ul>
  );
}
