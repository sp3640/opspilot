"use client";

import { AlertTriangle, CheckCircle2, HelpCircle, ShieldAlert } from "lucide-react";

import type { OverallHealthState } from "@/lib/dashboard-overview";

const BANNER_CONFIG: Record<OverallHealthState, { icon: typeof CheckCircle2; label: string; message: string; color: string }> = {
  HEALTHY: {
    icon: CheckCircle2,
    label: "Healthy",
    message: "No critical problems or active incidents were found across your organization.",
    color: "var(--success)",
  },
  DEGRADED: {
    icon: AlertTriangle,
    label: "Degraded",
    message: "Some non-critical problems need attention. See Warnings below.",
    color: "var(--warning)",
  },
  CRITICAL: {
    icon: ShieldAlert,
    label: "Critical",
    message: "Critical problems require immediate attention. See below.",
    color: "var(--danger)",
  },
  UNKNOWN: {
    icon: HelpCircle,
    label: "Unknown",
    message: "Not enough data has loaded yet to determine organization health.",
    color: "var(--muted-foreground)",
  },
};

/** "Is my organization healthy right now?" - the single, deterministic answer at the top of the dashboard. */
export function DashboardHealthBanner({ state }: { state: OverallHealthState }) {
  const config = BANNER_CONFIG[state];
  const Icon = config.icon;

  return (
    <div
      className="flex items-start gap-4 rounded-3xl border p-5 shadow-[var(--shadow-sm)]"
      style={{ borderColor: config.color, backgroundColor: `color-mix(in srgb, ${config.color} 8%, transparent)` }}
    >
      <div
        className="rounded-2xl p-3"
        style={{ color: config.color, backgroundColor: `color-mix(in srgb, ${config.color} 15%, transparent)` }}
      >
        <Icon aria-hidden="true" className="h-6 w-6" />
      </div>
      <div>
        <p className="text-xs font-semibold uppercase tracking-wider" style={{ color: config.color }}>
          Organization health
        </p>
        <h2 className="mt-1 text-2xl font-semibold tracking-tight">{config.label}</h2>
        <p className="mt-1 text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>
          {config.message}
        </p>
      </div>
    </div>
  );
}
