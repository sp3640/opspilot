'use client';

import { ArrowUpRight } from "lucide-react";
import { useDashboardActivity } from "@/hooks/use-dashboard-data";
import type { ActivityStatus } from "@/types/activity";

import { SectionCard } from "./section-card";

const statusColors: Record<ActivityStatus, string> = {
  success: "var(--success)",
  warning: "var(--warning)",
  info: "var(--primary)",
  error: "var(--danger)",
};

/** Timestamped platform events displayed as an accessible activity timeline. */
export function RecentActivity() {
  const { data: activities, isLoading } = useDashboardActivity();

  if (isLoading) {
    return (
      <SectionCard
        title="Recent activity"
        description="Latest changes across your platform"
        action={
          <button
            type="button"
            disabled
            className="inline-flex items-center gap-1 text-sm font-medium transition-opacity hover:opacity-75 opacity-50"
            style={{ color: "var(--primary)" }}
          >
            View all <ArrowUpRight aria-hidden="true" className="h-4 w-4" />
          </button>
        }
      >
        <div className="space-y-3">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-12 rounded-2xl bg-gray-700 animate-pulse" />
          ))}
        </div>
      </SectionCard>
    );
  }

  if (!activities || activities.length === 0) {
    return (
      <SectionCard
        title="Recent activity"
        description="Latest changes across your platform"
        action={
          <button
            type="button"
            className="inline-flex items-center gap-1 text-sm font-medium transition-opacity hover:opacity-75"
            style={{ color: "var(--primary)" }}
          >
            View all <ArrowUpRight aria-hidden="true" className="h-4 w-4" />
          </button>
        }
      >
        <div
          className="rounded-2xl p-6 text-center"
          style={{ backgroundColor: "color-mix(in srgb, var(--muted) 50%, transparent)" }}
        >
          <p style={{ color: "var(--muted-foreground)" }}>No recent activity</p>
        </div>
      </SectionCard>
    );
  }

  return (
    <SectionCard
      title="Recent activity"
      description="Latest changes across your platform"
      action={
        <button
          type="button"
          className="inline-flex items-center gap-1 text-sm font-medium transition-opacity hover:opacity-75"
          style={{ color: "var(--primary)" }}
        >
          View all <ArrowUpRight aria-hidden="true" className="h-4 w-4" />
        </button>
      }
    >
      <ol className="relative space-y-1 before:absolute before:bottom-5 before:left-[9px] before:top-5 before:w-px before:bg-[var(--border)]">
        {activities.map((activity) => (
          <li
            key={activity.id}
            className="relative grid grid-cols-[20px_minmax(0,1fr)] gap-3 rounded-2xl p-3 transition-colors hover:bg-[var(--muted)]"
          >
            <span
              className="relative z-10 mt-1.5 h-2.5 w-2.5 rounded-full ring-4 ring-[var(--card)]"
              style={{
                backgroundColor: statusColors[activity.status as ActivityStatus],
              }}
            />
            <div className="min-w-0">
              <div className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
                <h3 className="font-medium">{activity.title}</h3>
                <time
                  className="text-xs tabular-nums"
                  style={{ color: "var(--muted-foreground)" }}
                >
                  {new Date(activity.timestamp).toLocaleDateString()}
                </time>
              </div>
              <p className="mt-1 text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>
                {activity.description}
              </p>
            </div>
          </li>
        ))}
      </ol>
    </SectionCard>
  );
}
