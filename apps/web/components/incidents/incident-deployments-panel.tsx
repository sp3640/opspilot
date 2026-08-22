"use client";

import { useMemo, useState } from "react";
import { Rocket } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { DeploymentDetailsDrawer } from "@/components/deployments/deployment-details-drawer";
import { useDeploymentsByApplication } from "@/hooks/use-deployments";
import type { IncidentResponse } from "@/types/incident-api";

const DEPLOYMENT_LOOKUP = { page: 1, limit: 20, sort: "created_at" as const, order: "desc" as const };

/** Incident -> Deployments: every deployment on record for the affected application. */
export function IncidentDeploymentsPanel({ incident }: { incident: IncidentResponse }) {
  const hasApplication = Boolean(incident.applicationId);
  const [selectedDeploymentIndex, setSelectedDeploymentIndex] = useState<number | null>(null);
  const { data, error, isError, isLoading, refetch } = useDeploymentsByApplication(
    incident.applicationId ?? null,
    DEPLOYMENT_LOOKUP,
    hasApplication
  );
  const deployments = useMemo(() => data?.items ?? [], [data?.items]);

  if (!hasApplication) {
    return (
      <EmptyState
        icon={Rocket}
        title="No deployments available"
        description="This incident has no affected application to list deployments for."
      />
    );
  }

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load deployments."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={3} />;
  }

  if (deployments.length === 0) {
    return (
      <EmptyState
        icon={Rocket}
        title="No deployments recorded"
        description="This application has no deployment history yet."
      />
    );
  }

  return (
    <>
      <ul className="space-y-2">
        {deployments.map((deployment, index) => (
          <li key={deployment.id} className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
            <button
              type="button"
              onClick={() => setSelectedDeploymentIndex(index)}
              className="flex w-full items-start justify-between gap-3 text-left"
            >
              <div className="min-w-0">
                <p className="truncate text-sm font-medium underline-offset-2 hover:underline" style={{ color: "var(--primary)" }}>
                  {deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image}
                </p>
                <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  {deployment.environment} · {deployment.namespace} · {formatDateTime(deployment.createdAt)}
                </p>
              </div>
              <StatusBadge variant={deploymentStatusVariant(deployment.status)}>{deployment.status}</StatusBadge>
            </button>
          </li>
        ))}
      </ul>

      <DeploymentDetailsDrawer
        deployment={selectedDeploymentIndex !== null ? (deployments[selectedDeploymentIndex] ?? null) : null}
        open={selectedDeploymentIndex !== null}
        onClose={() => setSelectedDeploymentIndex(null)}
      />
    </>
  );
}

function deploymentStatusVariant(status: string): "success" | "warning" | "critical" | "archived" | "info" {
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

function formatDateTime(value: string) {
  return new Date(value).toLocaleString();
}
