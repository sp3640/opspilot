'use client';

import { AlertTriangle, FileText, FolderKanban, Users } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useDashboardStats } from "@/hooks/use-dashboard-data";

import { MetricBadge } from "./metric-badge";

type Metric = {
  title: string;
  value: string;
  description: string;
  trend: string;
  trendDirection: "up" | "down";
  trendTone: "success" | "warning";
  icon: LucideIcon;
};

/** High-level operational counts shown immediately below the dashboard summary. */
export function MetricsGrid() {
  const { data: stats, isLoading } = useDashboardStats();

  const metrics: Metric[] = stats
    ? [
        {
          title: "Projects",
          value: stats.projects.toString(),
          description: `${stats.projectsDeploying || 0} actively deploying`,
          trend: "16.7%",
          trendDirection: "up",
          trendTone: "success",
          icon: FolderKanban,
        },
        {
          title: "Incidents",
          value: stats.incidents.total.toString(),
          description: `${stats.incidentsRequiringAttention || stats.incidents.critical || 0} requires attention`,
          trend: `${stats.incidents.resolved} resolved`,
          trendDirection: "down",
          trendTone: "success",
          icon: AlertTriangle,
        },
        {
          title: "Audit logs",
          value: stats.auditLogs.toString(),
          description: "Events captured today",
          trend: "8.0%",
          trendDirection: "up",
          trendTone: "success",
          icon: FileText,
        },
        {
          title: "Users",
          value: stats.users.toString(),
          description: `${stats.newUsersThisWeek || 0} new this week`,
          trend: "20.0%",
          trendDirection: "up",
          trendTone: "success",
          icon: Users,
        },
      ]
    : [];

  if (isLoading) {
    return (
      <section aria-label="Operational metrics" className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <article
            key={i}
            className="rounded-2xl border p-5 shadow-[var(--shadow-sm)]"
            style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
          >
            <div className="h-8 w-16 rounded bg-gray-700 animate-pulse" />
            <div className="mt-6 h-6 w-20 rounded bg-gray-700 animate-pulse" />
          </article>
        ))}
      </section>
    );
  }

  return (
    <section aria-label="Operational metrics" className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {metrics.map(({ icon: Icon, ...metric }) => (
        <article
          key={metric.title}
          className="group relative overflow-hidden rounded-2xl border p-5 shadow-[var(--shadow-sm)] transition duration-300 hover:-translate-y-1 hover:shadow-[var(--shadow-md)]"
          style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
        >
          <div
            aria-hidden="true"
            className="absolute -right-8 -top-8 h-24 w-24 rounded-full blur-2xl transition-opacity duration-300 group-hover:opacity-100"
            style={{ backgroundColor: "color-mix(in srgb, var(--primary) 20%, transparent)", opacity: 0.55 }}
          />
          <div className="relative flex items-start justify-between gap-4">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.14em]" style={{ color: "var(--muted-foreground)" }}>{metric.title}</p>
              <p className="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">{metric.value}</p>
            </div>
            <div className="rounded-2xl p-3" style={{ backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)", color: "var(--primary)" }}>
              <Icon aria-hidden="true" className="h-5 w-5" />
            </div>
          </div>
          <div className="relative mt-5 flex items-center justify-between gap-2">
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>{metric.description}</p>
            <MetricBadge value={metric.trend} direction={metric.trendDirection} tone={metric.trendTone} />
          </div>
        </article>
      ))}
    </section>
  );
}
