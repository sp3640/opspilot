"use client";

import { useState } from "react";
import { AlertTriangle, CheckCircle2, Rocket, XCircle } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useCreateDeployment } from "@/hooks/use-deployments";
import type { DeploymentResponse } from "@/types/deployment-api";

/**
 * Redeploys the current configuration as a brand-new deployment record,
 * reusing POST /deployments exactly as-is - the same, already-wired
 * executor path Create always uses, so this genuinely re-applies to the
 * live cluster (unlike Cancel, which is database-only).
 */
export function DeploymentRedeploy({ deployment }: { deployment: DeploymentResponse }) {
  const createDeployment = useCreateDeployment();
  const [confirmed, setConfirmed] = useState(false);
  const [result, setResult] = useState<DeploymentResponse | null>(null);

  const handleRedeploy = async () => {
    try {
      const response = await createDeployment.mutateAsync({
        applicationId: deployment.applicationId,
        projectId: deployment.projectId,
        targetClusterId: deployment.targetClusterId,
        image: deployment.image,
        imageTag: deployment.imageTag || undefined,
        environment: deployment.environment,
        namespace: deployment.namespace,
        replicaCount: deployment.replicaCount,
        deploymentStrategy: deployment.deploymentStrategy,
        commitSha: deployment.commitSha,
        author: deployment.author,
      });
      setConfirmed(false);
      // Verify the result: Create's executor call runs synchronously, so
      // the response already reflects whether the redeploy actually
      // succeeded against the live cluster.
      setResult(response);
    } catch {
      // Error is surfaced via the mutation's error toast.
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
          Redeploying creates a new deployment record with this exact configuration and applies it to the live cluster.
        </p>
      </div>

      {result ? (
        <div
          className="flex items-start gap-3 rounded-2xl border p-4 text-sm"
          style={
            result.status === "Succeeded"
              ? { backgroundColor: "color-mix(in srgb, var(--success) 12%, transparent)", borderColor: "var(--success)", color: "var(--success)" }
              : result.status === "Failed"
                ? { backgroundColor: "color-mix(in srgb, var(--danger) 12%, transparent)", borderColor: "var(--danger)", color: "var(--danger)" }
                : { backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)", borderColor: "var(--border)", color: "var(--foreground)" }
          }
        >
          {result.status === "Succeeded" ? (
            <CheckCircle2 aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" />
          ) : result.status === "Failed" ? (
            <XCircle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" />
          ) : null}
          <p className="leading-6">
            Verified: the new deployment&apos;s status is now <strong>{result.status}</strong>.
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
          <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Replica count</dt>
          <dd className="mt-1 break-all font-semibold">{deployment.replicaCount}</dd>
        </div>
      </dl>

      <label className="flex items-start gap-2 text-sm">
        <input
          type="checkbox"
          checked={confirmed}
          onChange={(event) => setConfirmed(event.target.checked)}
          disabled={createDeployment.isPending}
          className="mt-0.5 h-4 w-4"
        />
        <span>I understand this will trigger a new deployment with this exact configuration.</span>
      </label>

      <div className="flex justify-end">
        <Button
          type="button"
          onClick={handleRedeploy}
          loading={createDeployment.isPending}
          disabled={!confirmed || createDeployment.isPending}
        >
          <Rocket aria-hidden="true" className="h-4 w-4" />
          Redeploy this configuration
        </Button>
      </div>
    </div>
  );
}
