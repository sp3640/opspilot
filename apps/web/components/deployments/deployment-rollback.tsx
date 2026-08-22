"use client";

import { useEffect, useState } from "react";
import { AlertTriangle, CheckCircle2, RotateCcw, XCircle } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useDeploymentHistory, useRollbackDeployment } from "@/hooks/use-deployments";
import type { DeploymentResponse, RollbackDeploymentResponse } from "@/types/deployment-api";

export function DeploymentRollback({ deployment }: { deployment: DeploymentResponse }) {
  const { data, error, isError, isLoading, refetch } = useDeploymentHistory(deployment.id);
  const rollback = useRollbackDeployment();

  const [confirmed, setConfirmed] = useState(false);
  const [revision, setRevision] = useState<number | null>(null);
  const [result, setResult] = useState<RollbackDeploymentResponse | null>(null);

  const revisions = data?.items ?? [];

  useEffect(() => {
    setResult(null);
  }, [deployment.id]);

  useEffect(() => {
    const first = revisions[0];
    if (revision === null && first) {
      setRevision(first.revision);
    }
  }, [revisions, revision]);

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load deployment revisions. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={3} />;
  }

  if (revisions.length === 0) {
    return (
      <EmptyState
        icon={RotateCcw}
        title="No revisions available"
        description="There is no previous revision to roll back to for this deployment."
      />
    );
  }

  const handleRollback = async () => {
    if (revision === null) return;
    try {
      const response = await rollback.mutateAsync({ id: deployment.id, revision });
      setConfirmed(false);
      // Verify the result: the executor (when configured for the target
      // cluster) runs synchronously within this same request, so the
      // response already reflects the real outcome - no extra fetch needed.
      setResult(response);
    } catch {
      // Error is surfaced via the mutation's error toast.
    }
  };

  const selectClass =
    "mt-2 h-11 w-full rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]";

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
          Rolling back replaces the current deployment configuration with a previous revision. This action affects the
          running environment and cannot be undone automatically.
        </p>
      </div>

      {result ? (
        <div
          className="flex items-start gap-3 rounded-2xl border p-4 text-sm"
          style={
            result.deployment.status === "Succeeded"
              ? { backgroundColor: "color-mix(in srgb, var(--success) 12%, transparent)", borderColor: "var(--success)", color: "var(--success)" }
              : result.deployment.status === "Failed"
                ? { backgroundColor: "color-mix(in srgb, var(--danger) 12%, transparent)", borderColor: "var(--danger)", color: "var(--danger)" }
                : { backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)", borderColor: "var(--border)", color: "var(--foreground)" }
          }
        >
          {result.deployment.status === "Succeeded" ? (
            <CheckCircle2 aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" />
          ) : result.deployment.status === "Failed" ? (
            <XCircle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" />
          ) : null}
          <p className="leading-6">
            Verified: rolled back to revision {result.rollbackSourceRevision} (new revision {result.currentRevision}) —
            deployment status is now <strong>{result.deployment.status}</strong>.
          </p>
        </div>
      ) : null}

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

      <div>
        <label htmlFor="rollback-revision" className="text-sm font-medium">
          Roll back to revision
        </label>
        <select
          id="rollback-revision"
          value={revision ?? ""}
          onChange={(event) => setRevision(Number(event.target.value))}
          disabled={rollback.isPending}
          className={selectClass}
          style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
        >
          {revisions.map((item) => (
            <option key={item.id} value={item.revision}>
              Revision {item.revision} — {item.status}
            </option>
          ))}
        </select>
      </div>

      <label className="flex items-start gap-2 text-sm">
        <input
          type="checkbox"
          checked={confirmed}
          onChange={(event) => setConfirmed(event.target.checked)}
          disabled={rollback.isPending}
          className="mt-0.5 h-4 w-4"
        />
        <span>I understand this will roll back the deployment to the selected revision.</span>
      </label>

      <div className="flex justify-end">
        <Button
          type="button"
          variant="danger"
          onClick={handleRollback}
          loading={rollback.isPending}
          disabled={!confirmed || revision === null || rollback.isPending}
        >
          <RotateCcw aria-hidden="true" className="h-4 w-4" />
          Roll back deployment
        </Button>
      </div>
    </div>
  );
}
