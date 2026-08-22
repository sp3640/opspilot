"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp, ClipboardList } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useDeploymentHistory, useDeploymentHistoryRevision } from "@/hooks/use-deployments";

export function DeploymentHistory({ deploymentId }: { deploymentId: string }) {
  const { data, error, isError, isLoading, refetch } = useDeploymentHistory(deploymentId);
  const [expandedRevision, setExpandedRevision] = useState<number | null>(null);

  const history = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load deployment history. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (history.length === 0) {
    return (
      <EmptyState
        icon={ClipboardList}
        title="No deployment history"
        description="Revisions for this deployment will appear here as it is updated."
      />
    );
  }

  return (
    <ol className="space-y-3">
      {history.map((item) => {
        const isExpanded = expandedRevision === item.revision;

        return (
          <li
            key={item.id}
            className="rounded-2xl border p-4"
            style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
          >
            <button
              type="button"
              onClick={() => setExpandedRevision(isExpanded ? null : item.revision)}
              aria-expanded={isExpanded}
              className="flex w-full items-start justify-between gap-3 text-left"
            >
              <div className="min-w-0">
                <p className="font-semibold">Revision {item.revision}</p>
                {item.triggeredBy ? (
                  <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                    Triggered by #{item.triggeredBy}
                  </p>
                ) : null}
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <StatusBadge variant={getStatusVariant(item.status)}>{item.status}</StatusBadge>
                {isExpanded ? (
                  <ChevronUp aria-hidden="true" className="h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
                ) : (
                  <ChevronDown aria-hidden="true" className="h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
                )}
              </div>
            </button>
            <p className="mt-3 text-sm leading-6">{item.changeSummary || "No summary provided"}</p>
            <p className="mt-3 text-xs" style={{ color: "var(--muted-foreground)" }}>
              {formatDate(item.createdAt)}
            </p>

            {isExpanded ? (
              <RevisionSnapshot deploymentId={deploymentId} revision={item.revision} />
            ) : null}
          </li>
        );
      })}
    </ol>
  );
}

/** Fetches GET /deployments/:id/history/:revision to show the full config snapshot at that revision. */
function RevisionSnapshot({ deploymentId, revision }: { deploymentId: string; revision: number }) {
  const { data, error, isError, isLoading } = useDeploymentHistoryRevision(deploymentId, revision);

  if (isLoading) {
    return (
      <div className="mt-4 border-t pt-4" style={{ borderColor: "var(--border)" }}>
        <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>Loading revision snapshot…</p>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="mt-4 border-t pt-4" style={{ borderColor: "var(--border)" }}>
        <p className="text-xs" style={{ color: "var(--danger)" }}>
          {error instanceof Error ? error.message : "Unable to load this revision's snapshot."}
        </p>
      </div>
    );
  }

  const fields: ReadonlyArray<[string, string]> = [
    ["Image", data.imageTag ? `${data.image}:${data.imageTag}` : data.image],
    ["Environment", data.environment],
    ["Namespace", data.namespace],
    ["Replica count", String(data.replicaCount)],
    ["Strategy", data.deploymentStrategy],
    ["Commit", data.commitSha || "Not available"],
    ["Author", data.author || "Not available"],
  ];

  return (
    <dl className="mt-4 grid grid-cols-2 gap-3 border-t pt-4 text-sm" style={{ borderColor: "var(--border)" }}>
      {fields.map(([label, value]) => (
        <div key={label}>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</dt>
          <dd className="mt-1 break-all">{value}</dd>
        </div>
      ))}
    </dl>
  );
}

function getStatusVariant(status: string): "success" | "warning" | "critical" | "archived" | "info" {
  switch (status) {
    case "Succeeded":
      return "success";
    case "Failed":
      return "critical";
    case "Pending":
    case "Queued":
    case "RolledBack":
      return "warning";
    case "Cancelled":
      return "archived";
    default:
      return "info";
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
