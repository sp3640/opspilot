"use client";

import { useMemo } from "react";
import { History } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useAuditLogs } from "@/hooks/use-audit";
import { formatDuration } from "@/lib/alert-correlation";
import type { AlertResponse } from "@/types/alert-api";

const AUDIT_QUERY = { page: 1, limit: 50, sort: "created_at" as const, order: "asc" as const };

/**
 * The alert's own audit trail (every acknowledge/resolve/reopen/escalation
 * OpsPilot has actually recorded, via GET /alerts/:id/audit-logs) is the
 * real, timestamped basis for a timeline - not a reconstruction. First
 * seen/last seen/duration/occurrence count come directly from the alert
 * itself, which is the ground truth for those fields regardless of whether
 * any audit entries exist.
 */
export function AlertTimeline({ alert }: { alert: AlertResponse }) {
  const { data, error, isError, isLoading, refetch } = useAuditLogs("alert", alert.id, AUDIT_QUERY);

  const summaryItems = useMemo(
    () => [
      { label: "First seen", value: formatDateTime(alert.firstSeenAt) },
      { label: "Last seen", value: formatDateTime(alert.lastSeenAt) },
      { label: "Duration", value: formatDuration(alert.firstSeenAt, alert.resolvedAt ?? undefined) },
      { label: "Occurrence count", value: String(alert.occurrenceCount) },
    ],
    [alert.firstSeenAt, alert.lastSeenAt, alert.resolvedAt, alert.occurrenceCount]
  );

  return (
    <div className="space-y-6">
      <dl className="grid grid-cols-2 gap-3">
        {summaryItems.map((item) => (
          <div key={item.label} className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{item.label}</dt>
            <dd className="mt-1 font-semibold">{item.value}</dd>
          </div>
        ))}
      </dl>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          History
        </h3>
        <div className="mt-3">
          {isError ? (
            <ErrorState
              description={error instanceof Error ? error.message : "Unable to load the alert's history."}
              onRetry={() => {
                void refetch();
              }}
            />
          ) : isLoading ? (
            <ListSkeleton rows={4} />
          ) : (data?.items ?? []).length === 0 ? (
            <EmptyState icon={History} title="No recorded history" description="No audit entries have been recorded for this alert yet." />
          ) : (
            <ol className="space-y-2">
              {(data?.items ?? []).map((entry) => (
                <li key={entry.id} className="rounded-2xl border p-3 text-sm" style={{ borderColor: "var(--border)" }}>
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span className="font-medium">{describeAuditEntry(entry.action, entry.field_name, entry.old_value, entry.new_value)}</span>
                    <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>{formatDateTime(entry.created_at)}</span>
                  </div>
                </li>
              ))}
            </ol>
          )}
        </div>
      </section>
    </div>
  );
}

function describeAuditEntry(action: string, fieldName: string, oldValue: string, newValue: string): string {
  if (action === "CREATE") return "Alert created";
  if (!fieldName) return "Alert updated";
  if (!oldValue) return `${fieldName} set to ${newValue}`;
  return `${fieldName} changed from ${oldValue} to ${newValue}`;
}

function formatDateTime(value?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}
