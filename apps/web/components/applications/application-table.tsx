"use client";

import { ChevronRight, Pencil, Trash2 } from "lucide-react";

import { StatusBadge, TableSkeleton } from "@/components/common";
import { Button } from "@/components/ui/button";
import type { ApplicationResponse } from "@/types/application-api";

type ApplicationTableProps = {
  applications: ApplicationResponse[];
  loading?: boolean;
  onRowClick?: (application: ApplicationResponse) => void;
  onEdit?: (application: ApplicationResponse) => void;
  onDelete?: (application: ApplicationResponse) => void;
};

export function ApplicationTable({ applications, loading, onRowClick, onEdit, onDelete }: ApplicationTableProps) {
  if (loading) {
    return <TableSkeleton columns={8} />;
  }

  const hasActions = Boolean(onEdit || onDelete);

  return (
    <div className="overflow-x-auto rounded-2xl border" style={{ borderColor: "var(--border)" }}>
      <table className="w-full min-w-[1120px] text-left text-sm">
        <caption className="sr-only">Applications list with runtime, status, repository, and activity metadata</caption>
        <thead
          className="text-xs uppercase tracking-[0.1em]"
          style={{
            color: "var(--muted-foreground)",
            backgroundColor: "color-mix(in srgb, var(--muted) 55%, transparent)",
          }}
        >
          <tr>
            <th scope="col" className="px-5 py-3 font-semibold">Name</th>
            <th scope="col" className="px-5 py-3 font-semibold">Runtime</th>
            <th scope="col" className="px-5 py-3 font-semibold">Status</th>
            <th scope="col" className="px-5 py-3 font-semibold">Repository</th>
            <th scope="col" className="px-5 py-3 font-semibold">Branch</th>
            <th scope="col" className="px-5 py-3 font-semibold">Port</th>
            <th scope="col" className="px-5 py-3 font-semibold">Environment</th>
            <th scope="col" className="px-5 py-3 font-semibold">Updated</th>
            <th scope="col" className="w-12 px-5 py-3"><span className="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          {applications.map((application) => (
            <tr
              key={application.id}
              role="button"
              tabIndex={0}
              aria-label={`Open application details for ${application.name}`}
              onClick={() => onRowClick?.(application)}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onRowClick?.(application);
                }
              }}
              className="cursor-pointer border-t transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
              style={{ borderColor: "var(--border)" }}
            >
              <td className="px-5 py-4 font-medium">{application.name}</td>
              <td className="px-5 py-4">{application.runtime}</td>
              <td className="px-5 py-4"><StatusBadge variant={getStatusVariant(application.status)}>{application.status}</StatusBadge></td>
              <td className="max-w-[280px] truncate px-5 py-4" title={application.repositoryUrl || ""}>
                {application.repositoryUrl || "-"}
              </td>
              <td className="px-5 py-4">{application.defaultBranch || "-"}</td>
              <td className="px-5 py-4">{application.port}</td>
              <td className="px-5 py-4">{application.environment || "-"}</td>
              <td className="px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                {formatDate(application.updatedAt)}
              </td>
              <td className="px-5 py-4">
                {hasActions ? (
                  <div className="flex items-center gap-1">
                    {onEdit && (
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={(event) => {
                          event.stopPropagation();
                          onEdit(application);
                        }}
                        className="h-9 w-9 rounded-xl p-0"
                        aria-label={`Edit ${application.name}`}
                      >
                        <Pencil aria-hidden="true" className="h-4 w-4" />
                      </Button>
                    )}
                    {onDelete && (
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={(event) => {
                          event.stopPropagation();
                          onDelete(application);
                        }}
                        className="h-9 w-9 rounded-xl p-0"
                        aria-label={`Delete ${application.name}`}
                      >
                        <Trash2 aria-hidden="true" className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                ) : (
                  <ChevronRight aria-hidden="true" className="h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function getStatusVariant(status: string): "success" | "info" | "archived" {
  switch (status) {
    case "Ready":
      return "success";
    case "Archived":
      return "archived";
    default:
      return "info";
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
