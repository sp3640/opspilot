"use client";

import { useState } from "react";
import { Ban, History, RotateCcw, Rocket, ShieldAlert, XCircle } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useAuditLogs } from "@/hooks/use-audit";
import { useHasPermission } from "@/store/auth-store";
import type { DeploymentResponse } from "@/types/deployment-api";

import { DeploymentCancel } from "./deployment-cancel";
import { DeploymentRedeploy } from "./deployment-redeploy";
import { DeploymentRollback } from "./deployment-rollback";

const AUDIT_QUERY = { page: 1, limit: 20, sort: "created_at" as const, order: "desc" as const };

type RemediationAction = "redeploy" | "rollback" | "cancel";

const ACTIONS: ReadonlyArray<{ key: RemediationAction; label: string; icon: typeof Rocket }> = [
  { key: "redeploy", label: "Redeploy", icon: Rocket },
  { key: "rollback", label: "Rollback", icon: RotateCcw },
  { key: "cancel", label: "Cancel", icon: Ban },
];

/**
 * Authorized remediation (Phase 20): every safe operational action OpsPilot
 * can genuinely take against this deployment, in one place. Redeploy and
 * Rollback both reuse the same, already-wired executor that applies to the
 * live cluster; Cancel is honestly disclosed as database-only, since there
 * is no existing capability to stop a running apply. Restart pod/workload
 * and Scale workload are listed as explicitly unsupported rather than
 * silently omitted or implemented against nothing - there is no backend
 * capability for either today.
 */
export function DeploymentRemediation({ deployment }: { deployment: DeploymentResponse }) {
  const canRemediate = useHasPermission("deployment:manage");
  const [selectedAction, setSelectedAction] = useState<RemediationAction>("redeploy");

  return (
    <div className="space-y-6">
      {canRemediate ? (
        <section>
          <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
            Available actions
          </h3>
          <div className="mt-3 flex gap-2">
            {ACTIONS.map(({ key, label, icon: Icon }) => {
              const isActive = selectedAction === key;
              return (
                <button
                  key={key}
                  type="button"
                  onClick={() => setSelectedAction(key)}
                  aria-pressed={isActive}
                  className="inline-flex flex-1 items-center justify-center gap-1.5 rounded-2xl border px-3 py-2.5 text-sm font-medium transition-colors"
                  style={{
                    borderColor: isActive ? "var(--primary)" : "var(--border)",
                    color: isActive ? "var(--primary)" : "var(--foreground)",
                    backgroundColor: isActive ? "color-mix(in srgb, var(--primary) 10%, transparent)" : "var(--card)",
                  }}
                >
                  <Icon aria-hidden="true" className="h-4 w-4" />
                  {label}
                </button>
              );
            })}
          </div>

          <div className="mt-4">
            {selectedAction === "redeploy" ? (
              <DeploymentRedeploy deployment={deployment} />
            ) : selectedAction === "rollback" ? (
              <DeploymentRollback deployment={deployment} />
            ) : (
              <DeploymentCancel deployment={deployment} />
            )}
          </div>
        </section>
      ) : (
        <EmptyState
          icon={ShieldAlert}
          title="No remediation permission"
          description="You have read-only access to deployments. Remediation actions require the deployment:manage permission."
        />
      )}

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Not currently supported
        </h3>
        <ul className="mt-3 space-y-2">
          <UnsupportedAction
            label="Restart pod"
            reason="No backend capability exists yet to restart an individual pod - only read access to pods is implemented."
          />
          <UnsupportedAction
            label="Scale workload"
            reason="No backend capability exists yet to scale a live workload independently of redeploying it."
          />
        </ul>
      </section>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Recent remediation activity
        </h3>
        <div className="mt-3">
          <RemediationAuditTrail deploymentId={deployment.id} />
        </div>
      </section>
    </div>
  );
}

function UnsupportedAction({ label, reason }: { label: string; reason: string }) {
  return (
    <li className="flex items-start gap-3 rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
      <XCircle aria-hidden="true" className="mt-0.5 h-4 w-4 shrink-0" style={{ color: "var(--muted-foreground)" }} />
      <div>
        <p className="text-sm font-medium">{label}</p>
        <p className="mt-0.5 text-xs" style={{ color: "var(--muted-foreground)" }}>{reason}</p>
      </div>
    </li>
  );
}

function RemediationAuditTrail({ deploymentId }: { deploymentId: string }) {
  const { data, error, isError, isLoading, refetch } = useAuditLogs("deployment", deploymentId, AUDIT_QUERY);
  const entries = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load this deployment's audit trail."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={3} />;
  }

  if (entries.length === 0) {
    return (
      <EmptyState
        icon={History}
        title="No remediation activity yet"
        description="Actions taken on this deployment (rollback, cancel, redeploy execution) will appear here."
      />
    );
  }

  return (
    <ol className="space-y-2">
      {entries.map((entry) => (
        <li key={entry.id} className="rounded-2xl border p-3 text-sm" style={{ borderColor: "var(--border)" }}>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="font-medium">{describeAuditEntry(entry.field_name, entry.old_value, entry.new_value)}</span>
            <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>{formatDateTime(entry.created_at)}</span>
          </div>
        </li>
      ))}
    </ol>
  );
}

function describeAuditEntry(fieldName: string, oldValue: string, newValue: string): string {
  if (!fieldName) return "Deployment updated";
  if (fieldName === "revision") return `Rollback: revision ${oldValue} → ${newValue}`;
  if (!oldValue) return `${fieldName} set to ${newValue}`;
  return `${fieldName} changed from ${oldValue} to ${newValue}`;
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString();
}
