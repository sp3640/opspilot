"use client";

import Link from "next/link";

import type { DashboardCounts } from "@/lib/dashboard-overview";

const STAT_TILES: ReadonlyArray<{ key: keyof DashboardCounts; label: string; href: string; tone?: "danger" }> = [
  { key: "applications", label: "Applications", href: "/projects" },
  { key: "clusters", label: "Clusters", href: "/clusters" },
  { key: "activeAlerts", label: "Active Alerts", href: "/alerts" },
  { key: "criticalAlerts", label: "Critical Alerts", href: "/alerts", tone: "danger" },
  { key: "activeIncidents", label: "Active Incidents", href: "/incidents", tone: "danger" },
  { key: "recentDeployments", label: "Recent Deployments", href: "/projects" },
  { key: "failedDeployments", label: "Failed Deployments", href: "/projects", tone: "danger" },
  { key: "unhealthyPods", label: "Unhealthy Pods", href: "/clusters", tone: "danger" },
];

/**
 * Real counts only - a null value (a source that couldn't be loaded, or
 * wasn't attempted) renders as "—", never a fabricated 0. Every tile links
 * to where that resource actually lives.
 */
export function DashboardStatGrid({ counts }: { counts: DashboardCounts }) {
  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
      {STAT_TILES.map(({ key, label, href, tone }) => {
        const value = counts[key];
        const isProblem = tone === "danger" && typeof value === "number" && value > 0;

        return (
          <Link
            key={key}
            href={href}
            className="rounded-2xl border p-4 transition-colors hover:bg-[var(--muted)]"
            style={{
              borderColor: isProblem ? "var(--danger)" : "var(--border)",
              backgroundColor: isProblem ? "color-mix(in srgb, var(--danger) 6%, transparent)" : "var(--card)",
            }}
          >
            <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</p>
            <p className="mt-1 text-2xl font-semibold tracking-tight" style={{ color: isProblem ? "var(--danger)" : "var(--foreground)" }}>
              {value === null ? "—" : value}
            </p>
          </Link>
        );
      })}
    </div>
  );
}
