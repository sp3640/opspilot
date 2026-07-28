import { ArrowUpRight } from "lucide-react";

import { activities } from "@/data/activity";
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
  return (
    <SectionCard title="Recent activity" description="Latest changes across your platform" action={<button type="button" className="inline-flex items-center gap-1 text-sm font-medium transition-opacity hover:opacity-75" style={{ color: "var(--primary)" }}>View all <ArrowUpRight aria-hidden="true" className="h-4 w-4" /></button>}>
      <ol className="relative space-y-1 before:absolute before:bottom-5 before:left-[9px] before:top-5 before:w-px before:bg-[var(--border)]">
        {activities.map((activity) => (
          <li key={activity.id} className="relative grid grid-cols-[20px_minmax(0,1fr)] gap-3 rounded-2xl p-3 transition-colors hover:bg-[var(--muted)]">
            <span className="relative z-10 mt-1.5 h-2.5 w-2.5 rounded-full ring-4 ring-[var(--card)]" style={{ backgroundColor: statusColors[activity.status] }} />
            <div className="min-w-0"><div className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1"><h3 className="font-medium">{activity.title}</h3><time className="text-xs tabular-nums" style={{ color: "var(--muted-foreground)" }}>{activity.time}</time></div><p className="mt-1 text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>{activity.description}</p></div>
          </li>
        ))}
      </ol>
    </SectionCard>
  );
}
