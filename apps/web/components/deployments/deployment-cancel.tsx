"use client";

import { useState } from "react";
import { AlertTriangle, Ban } from "lucide-react";

import { EmptyState, ErrorState } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useCancelDeployment } from "@/hooks/use-deployments";
import type { DeploymentResponse } from "@/types/deployment-api";

const CANCELLABLE_STATUSES = ["Pending", "Queued", "Running"];

export function DeploymentCancel({ deployment }: { deployment: DeploymentResponse }) {
  const cancel = useCancelDeployment();
  const [confirmed, setConfirmed] = useState(false);

  if (!CANCELLABLE_STATUSES.includes(deployment.status)) {
    return (
      <EmptyState
        icon={Ban}
        title="Deployment cannot be cancelled"
        description="Only pending, queued, or running deployments can be cancelled."
      />
    );
  }

  const handleCancel = async () => {
    try {
      await cancel.mutateAsync(deployment.id);
      setConfirmed(false);
    } catch {
      // Error is surfaced via the mutation's error toast and the inline ErrorState below.
    }
  };

  return (
    <div className="space-y-5">
      <div
        className="flex items-start gap-3 rounded-2xl border p-4 text-sm"
        style={{
          backgroundColor: "color-mix(in srgb, var(--warning) 12%, transparent)",
          borderColor: "var(--warning)",
          color: "var(--warning)",
        }}
      >
        <AlertTriangle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" />
        <p className="leading-6">
          Cancelling stops this in-progress deployment. The target environment may be left in a partially updated state,
          and this action cannot be undone.
        </p>
      </div>

      <dl className="grid grid-cols-2 gap-3">
        <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Image</dt>
          <dd className="mt-1 break-all font-semibold">{deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image}</dd>
        </div>
        <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Environment</dt>
          <dd className="mt-1 break-all font-semibold">{deployment.environment || "-"}</dd>
        </div>
        <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Namespace</dt>
          <dd className="mt-1 break-all font-semibold">{deployment.namespace || "-"}</dd>
        </div>
        <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}>
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Current status</dt>
          <dd className="mt-1 break-all font-semibold">{deployment.status}</dd>
        </div>
      </dl>

      {cancel.isError ? (
        <ErrorState
          description={cancel.error instanceof Error ? cancel.error.message : "Unable to cancel deployment. Please try again."}
          onRetry={() => {
            void handleCancel();
          }}
        />
      ) : null}

      <label className="flex items-start gap-2 text-sm">
        <input
          type="checkbox"
          checked={confirmed}
          onChange={(event) => setConfirmed(event.target.checked)}
          disabled={cancel.isPending}
          className="mt-0.5 h-4 w-4"
        />
        <span>I understand this will cancel the deployment and cannot be undone.</span>
      </label>

      <div className="flex justify-end">
        <Button
          type="button"
          variant="danger"
          onClick={handleCancel}
          loading={cancel.isPending}
          disabled={!confirmed || cancel.isPending}
        >
          <Ban aria-hidden="true" className="h-4 w-4" />
          Cancel deployment
        </Button>
      </div>
    </div>
  );
}
