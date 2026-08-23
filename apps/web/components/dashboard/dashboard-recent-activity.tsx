"use client";

import Link from "next/link";
import { History } from "lucide-react";

import { EmptyState } from "@/components/common";
import type { ActivityItem } from "@/lib/dashboard-overview";

import { SectionCard } from "./section-card";

/** A real, merged feed of alerts firing, incidents opening, and deployments running — sorted chronologically, nothing synthesized. */
export function DashboardRecentActivity({ items }: { items: ActivityItem[] }) {
  return (
    <SectionCard title="Recent Activity" description="Alerts, incidents, and deployments across your organization.">
      {items.length === 0 ? (
        <EmptyState icon={History} title="No recent activity" description="Nothing has happened recently across your organization." />
      ) : (
        <ol className="space-y-2">
          {items.map((item) => (
            <li key={item.id}>
              <Link
                href={item.href}
                className="flex items-center justify-between gap-3 rounded-2xl border p-3 text-sm transition-colors hover:bg-[var(--muted)]"
                style={{ borderColor: "var(--border)" }}
              >
                <span className="min-w-0 truncate">{item.title}</span>
                <span className="shrink-0 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  {formatDateTime(item.timestamp)}
                </span>
              </Link>
            </li>
          ))}
        </ol>
      )}
    </SectionCard>
  );
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString();
}
