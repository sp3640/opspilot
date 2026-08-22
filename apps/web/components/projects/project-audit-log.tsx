"use client";

import { useState } from "react";
import { ClipboardList } from "lucide-react";

import { EmptyState, ErrorState, StatusBadge, TableSkeleton } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useAuditLogs } from "@/hooks/use-audit";
import { PAGINATION_DEFAULT_PAGE, PAGINATION_DEFAULT_PAGE_SIZE } from "@/lib/constants/pagination";
import type { AuditAction, AuditQueryParams } from "@/types/audit-api";

export function ProjectAuditLog({ projectId }: { projectId: string }) {
  const [page, setPage] = useState(PAGINATION_DEFAULT_PAGE);

  const params: AuditQueryParams = { page, limit: PAGINATION_DEFAULT_PAGE_SIZE, sort: "created_at", order: "desc" };
  const { data, error, isError, isLoading, refetch } = useAuditLogs("project", projectId, params);
  const items = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load audit logs. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <TableSkeleton columns={5} rows={5} />;
  }

  if (items.length === 0) {
    return (
      <EmptyState
        icon={ClipboardList}
        title="No audit activity yet"
        description="Changes made to this project and its resources will appear here."
      />
    );
  }

  return (
    <div className="space-y-4">
      <div className="overflow-x-auto rounded-2xl border" style={{ borderColor: "var(--border)" }}>
        <table className="w-full min-w-[640px] text-left text-sm">
          <thead
            className="text-xs uppercase tracking-[0.1em]"
            style={{ color: "var(--muted-foreground)", backgroundColor: "color-mix(in srgb, var(--muted) 55%, transparent)" }}
          >
            <tr>
              <th className="px-4 py-3 font-semibold">Time</th>
              <th className="px-4 py-3 font-semibold">Action</th>
              <th className="px-4 py-3 font-semibold">Entity</th>
              <th className="px-4 py-3 font-semibold">Field</th>
              <th className="px-4 py-3 font-semibold">User</th>
            </tr>
          </thead>
          <tbody>
            {items.map((item) => (
              <tr key={item.id} className="border-t" style={{ borderColor: "var(--border)" }}>
                <td className="px-4 py-3 whitespace-nowrap">{new Date(item.created_at).toLocaleString()}</td>
                <td className="px-4 py-3">
                  <StatusBadge variant={badgeVariantForAction(item.action)}>{item.action}</StatusBadge>
                </td>
                <td className="px-4 py-3">
                  {item.entity_type} <span style={{ color: "var(--muted-foreground)" }}>#{item.entity_id}</span>
                </td>
                <td className="px-4 py-3">{item.field_name || "—"}</td>
                <td className="px-4 py-3">User {item.user_id}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {data && data.total > 0 ? (
        <nav
          aria-label="Project audit log pagination"
          className="flex flex-col gap-3 border-t pt-4 sm:flex-row sm:items-center sm:justify-between"
          style={{ borderColor: "var(--border)" }}
        >
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
            Showing {Math.min((data.page - 1) * data.limit + 1, data.total)}–{Math.min(data.page * data.limit, data.total)} of{" "}
            {data.total} entries
          </p>
          <div className="flex items-center gap-2">
            <Button type="button" variant="secondary" onClick={() => setPage((current) => current - 1)} disabled={page === 1} className="px-3 py-2">
              Previous
            </Button>
            <span className="px-2 text-sm tabular-nums" style={{ color: "var(--muted-foreground)" }}>
              Page {data.page} of {data.totalPages}
            </span>
            <Button
              type="button"
              variant="secondary"
              onClick={() => setPage((current) => current + 1)}
              disabled={page === data.totalPages}
              className="px-3 py-2"
            >
              Next
            </Button>
          </div>
        </nav>
      ) : null}
    </div>
  );
}

function badgeVariantForAction(action: AuditAction): "success" | "warning" | "critical" {
  switch (action) {
    case "CREATE":
      return "success";
    case "DELETE":
      return "critical";
    default:
      return "warning";
  }
}
