"use client";

import { AlertTriangle, CheckCircle2, HelpCircle, ShieldAlert } from "lucide-react";

import type { OverallHealthState } from "@/lib/dashboard-overview";

const BANNER_CONFIG: Record<OverallHealthState, { icon: typeof CheckCircle2; label: string; message: string; color: string; text: string; bg: string; border: string }> = {
  HEALTHY: {
    icon: CheckCircle2,
    label: "Healthy",
    message: "No critical problems or active incidents were found across your organization.",
    color: "var(--success)",
    text: "var(--success-text)",
    bg: "var(--success-bg)",
    border: "var(--success-border)",
  },
  DEGRADED: {
    icon: AlertTriangle,
    label: "Degraded",
    message: "Some non-critical problems need attention. See Warnings below.",
    color: "var(--warning)",
    text: "var(--warning-text)",
    bg: "var(--warning-bg)",
    border: "var(--warning-border)",
  },
  CRITICAL: {
    icon: ShieldAlert,
    label: "Critical",
    message: "Critical problems require immediate attention. See below.",
    color: "var(--danger)",
    text: "var(--danger-text)",
    bg: "var(--danger-bg)",
    border: "var(--danger-border)",
  },
  UNKNOWN: {
    icon: HelpCircle,
    label: "Unknown",
    message: "Not enough data has loaded yet to determine organization health.",
    color: "var(--muted-foreground)",
    text: "var(--muted-foreground)",
    bg: "var(--card)",
    border: "var(--border)",
  },
};

/** "Is my organization healthy right now?" - the single, deterministic answer at the top of the dashboard. */
export function DashboardHealthBanner({ state }: { state: OverallHealthState }) {
  const config = BANNER_CONFIG[state];
  const Icon = config.icon;

  return (
    <div
      className="flex items-start gap-4 rounded-3xl border p-5 shadow-[var(--shadow-sm)]"
      style={{ borderColor: config.border, backgroundColor: config.bg }}
    >
      <div
        className="rounded-2xl p-3"
        style={{ color: config.color, backgroundColor: `color-mix(in srgb, ${config.color} 14%, transparent)` }}
      >
        <Icon aria-hidden="true" className="h-6 w-6" />
      </div>
      <div>
        <p className="text-xs font-semibold uppercase tracking-wider" style={{ color: config.text }}>
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
