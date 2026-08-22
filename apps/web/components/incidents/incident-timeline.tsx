"use client";

import { useMemo } from "react";
import { History } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useAuditLogs } from "@/hooks/use-audit";
import { formatDuration } from "@/lib/alert-correlation";
import type { IncidentResponse } from "@/types/incident-api";

const AUDIT_QUERY = { page: 1, limit: 50, sort: "created_at" as const, order: "asc" as const };

/**
 * The incident's real audit trail (GET /incidents/:id/audit-logs, already
 * implemented before this phase) - every title/description/severity/status/
 * application/owner-team change actually recorded - plus created/resolved/
 * duration computed directly from the incident's own timestamps.
 */
export function IncidentTimeline({ incident }: { incident: IncidentResponse }) {
  const { data, error, isError, isLoading, refetch } = useAuditLogs("incident", incident.id, AUDIT_QUERY);

  const summaryItems = useMemo(
    () => [
      { label: "Created", value: formatDateTime(incident.createdAt) },
      { label: "Resolved", value: incident.resolvedAt ? formatDateTime(incident.resolvedAt) : "Not resolved" },
      { label: "Duration", value: formatDuration(incident.createdAt, incident.resolvedAt ?? undefined) },
    ],
    [incident.createdAt, incident.resolvedAt]
  );

  return (
    <div className="space-y-6">
      <dl className="grid grid-cols-3 gap-3">
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
              description={error instanceof Error ? error.message : "Unable to load the incident's history."}
              onRetry={() => {
                void refetch();
              }}
            />
          ) : isLoading ? (
            <ListSkeleton rows={4} />
          ) : (data?.items ?? []).length === 0 ? (
            <EmptyState icon={History} title="No recorded history" description="No audit entries have been recorded for this incident yet." />
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
  if (action === "CREATE") return "Incident created";
  if (!fieldName) return "Incident updated";
  if (!oldValue) return `${fieldName} set to ${newValue}`;
  return `${fieldName} changed from ${oldValue} to ${newValue}`;
}

function formatDateTime(value?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}
