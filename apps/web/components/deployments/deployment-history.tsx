"use client";

import { ClipboardList } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useDeploymentHistory } from "@/hooks/use-deployments";

export function DeploymentHistory({ deploymentId }: { deploymentId: string }) {
  const { data, error, isError, isLoading, refetch } = useDeploymentHistory(deploymentId);

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
      {history.map((item) => (
        <li
          key={item.id}
          className="rounded-2xl border p-4"
          style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
        >
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="font-semibold">Revision {item.revision}</p>
              {item.triggeredBy ? (
                <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  Triggered by #{item.triggeredBy}
                </p>
              ) : null}
            </div>
            <StatusBadge variant={getStatusVariant(item.status)}>{item.status}</StatusBadge>
          </div>
          <p className="mt-3 text-sm leading-6">{item.changeSummary || "No summary provided"}</p>
          <p className="mt-3 text-xs" style={{ color: "var(--muted-foreground)" }}>
            {formatDate(item.createdAt)}
          </p>
        </li>
      ))}
    </ol>
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
