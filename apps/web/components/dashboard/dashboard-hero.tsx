'use client';

import { Activity, CheckCircle2, Cloud, Rocket } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useDashboardSummary } from "@/hooks/use-dashboard-data";

type SummaryItem = [string, string, LucideIcon];

/** The dashboard's at-a-glance operational briefing. */
export function DashboardHero() {
  const { data: summary, isLoading } = useDashboardSummary();

  const summaryItems: SummaryItem[] | null = summary
    ? [
        ["Environment", summary.environment, Cloud],
        ["Last deployment", summary.lastDeployment, Rocket],
        ["Uptime", summary.uptime, Activity],
      ]
    : null;

  const greeting = summary?.greeting || "Good morning";

  return (
    <section
      aria-labelledby="dashboard-heading"
      className="relative overflow-hidden rounded-3xl border px-5 py-6 shadow-[var(--shadow-md)] sm:px-7 sm:py-8"
      style={{
        background: "linear-gradient(135deg, color-mix(in srgb, var(--card) 94%, var(--primary)), var(--card) 58%)",
        borderColor: "color-mix(in srgb, var(--border) 75%, var(--primary))",
      }}
    >
      <div aria-hidden="true" className="absolute -right-20 -top-28 h-64 w-64 rounded-full blur-3xl" style={{ backgroundColor: "color-mix(in srgb, var(--primary) 18%, transparent)" }} />
      <div className="relative grid gap-7 xl:grid-cols-[1.2fr_auto] xl:items-end">
        <div>
          <div className="mb-4 inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-sm font-medium" style={{ borderColor: "color-mix(in srgb, var(--success) 35%, var(--border))", backgroundColor: "color-mix(in srgb, var(--success) 10%, transparent)", color: "var(--success)" }}>
            <CheckCircle2 aria-hidden="true" className="h-4 w-4" />
            {isLoading ? "Loading..." : "All systems operational"}
          </div>
          <h1 id="dashboard-heading" className="text-3xl font-semibold tracking-tight sm:text-4xl">
            {isLoading ? "Loading..." : greeting}
          </h1>
          <p className="mt-3 max-w-2xl text-base leading-7 sm:text-lg" style={{ color: "var(--muted-foreground)" }}>
            {isLoading
              ? "Fetching your platform status..."
              : "Your platform is stable. Three active incidents are being monitored and all core services are operating within their targets."}
          </p>
        </div>

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3 xl:min-w-[530px]">
          {isLoading ? (
            <div className="col-span-full flex items-center justify-center rounded-2xl border p-8" style={{ backgroundColor: "color-mix(in srgb, var(--background) 24%, transparent)", borderColor: "var(--border)" }}>
              <span style={{ color: "var(--muted-foreground)" }}>Loading summary...</span>
            </div>
          ) : summaryItems ? (
            summaryItems.map(([label, value, Icon]) => (
              <div key={label} className="rounded-2xl border p-4" style={{ backgroundColor: "color-mix(in srgb, var(--background) 24%, transparent)", borderColor: "var(--border)" }}>
                <Icon aria-hidden="true" className="h-4 w-4" style={{ color: "var(--primary)" }} />
                <p className="mt-4 text-xs font-semibold uppercase tracking-[0.12em]" style={{ color: "var(--muted-foreground)" }}>{label}</p>
                <p className="mt-1 text-sm font-semibold">{value}</p>
              </div>
            ))
          ) : null}
        </div>
      </div>
    </section>
  );
}
