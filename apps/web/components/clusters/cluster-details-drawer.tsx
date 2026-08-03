"use client";

import { useEffect, useMemo, useState } from "react";
import {
  CheckCircle2,
  Pencil,
  RefreshCw,
  ShieldCheck,
  ShieldPlus,
  Trash2,
  X,
} from "lucide-react";

import { ErrorState } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useCluster, useSetDefaultCluster, useValidateCluster } from "@/hooks/use-clusters";
import type { ClusterValidationResponse } from "@/types/cluster-api";

import { ClusterProviderBadge, ClusterStatusBadge } from "./cluster-status";
import { DeleteClusterDialog } from "./delete-cluster-dialog";
import { EditClusterModal } from "./edit-cluster-modal";

export function ClusterDetailsDrawer({
  clusterID,
  projectNameById,
  onClose,
}: {
  clusterID: string | null;
  projectNameById: Map<string, string>;
  onClose: () => void;
}) {
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [validationResult, setValidationResult] = useState<ClusterValidationResponse | null>(null);
  const { data: cluster, error, isError, isLoading, refetch } = useCluster(clusterID);
  const validateCluster = useValidateCluster();
  const setDefaultCluster = useSetDefaultCluster();

  useEffect(() => {
    setEditModalOpen(false);
    setDeleteDialogOpen(false);
    setValidationResult(null);
  }, [clusterID]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      if (editModalOpen || deleteDialogOpen) return;
      onClose();
    };
    if (clusterID) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [clusterID, onClose, editModalOpen, deleteDialogOpen]);

  const validating = validateCluster.isPending;
  const settingDefault = setDefaultCluster.isPending;

  const projectDisplay = useMemo(() => {
    if (!cluster) return "-";
    const projectName = projectNameById.get(cluster.projectId);
    if (!projectName) return cluster.projectId;
    return `${projectName} (${cluster.projectId})`;
  }, [cluster, projectNameById]);

  const runValidation = async () => {
    if (!cluster) return;
    const result = await validateCluster.mutateAsync(cluster.id);
    setValidationResult(result);
    await refetch();
  };

  const handleSetDefault = async () => {
    if (!cluster || cluster.isDefault) return;
    await setDefaultCluster.mutateAsync(cluster.id);
    await refetch();
  };

  if (!clusterID) return null;

  if (isLoading) {
    return (
      <DrawerFrame onClose={onClose} ariaLabel="Loading cluster details">
        <div className="flex flex-1 items-center justify-center p-5 text-sm" style={{ color: "var(--muted-foreground)" }}>
          Loading cluster details...
        </div>
      </DrawerFrame>
    );
  }

  if (isError || !cluster) {
    return (
      <DrawerFrame onClose={onClose} ariaLabel="Cluster details unavailable">
        <div className="p-5">
          <ErrorState
            description={error instanceof Error ? error.message : "Unable to load cluster details."}
            onRetry={() => {
              void refetch();
            }}
          />
        </div>
      </DrawerFrame>
    );
  }

  return (
    <div
      className="fixed inset-0 z-[60] flex justify-end bg-[color:color-mix(in_srgb,var(--background)_72%,transparent)]"
      role="presentation"
      onMouseDown={onClose}
    >
      <aside
        role="dialog"
        aria-modal="true"
        aria-labelledby="cluster-details-title"
        onMouseDown={(event) => event.stopPropagation()}
        className="flex h-full w-full max-w-xl flex-col border-l shadow-[var(--shadow-lg)]"
        style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
      >
        <header className="border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <h2 id="cluster-details-title" className="truncate text-xl font-semibold tracking-tight">
                  {cluster.name}
                </h2>
                {cluster.isDefault ? (
                  <span className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide" style={{ color: "var(--success)", backgroundColor: "color-mix(in srgb, var(--success) 12%, transparent)" }}>
                    <ShieldCheck aria-hidden={true} className="h-3 w-3" />
                    Default
                  </span>
                ) : null}
              </div>
              <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Cluster workspace
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="secondary"
                onClick={() => setEditModalOpen(true)}
                className="px-3"
                aria-label="Edit cluster"
              >
                <Pencil aria-hidden="true" className="h-4 w-4" />
                <span className="hidden sm:inline">Edit</span>
              </Button>
              <Button
                type="button"
                variant="ghost"
                onClick={() => setDeleteDialogOpen(true)}
                className="px-3 text-[var(--danger)]"
                aria-label="Delete cluster"
              >
                <Trash2 aria-hidden="true" className="h-4 w-4" />
                <span className="hidden sm:inline">Delete</span>
              </Button>
              <Button
                type="button"
                variant="ghost"
                onClick={onClose}
                className="h-9 w-9 rounded-xl p-0"
                aria-label="Close cluster details"
              >
                <X aria-hidden="true" className="h-4 w-4" />
              </Button>
            </div>
          </div>

          <div className="mt-4 flex flex-wrap gap-2">
            <ClusterStatusBadge status={cluster.status} />
            <ClusterProviderBadge provider={cluster.provider} />
          </div>

          <div className="mt-4 flex flex-wrap gap-2">
            <Button
              type="button"
              variant="secondary"
              onClick={() => {
                void runValidation();
              }}
              loading={validating}
              disabled={validating}
            >
              <CheckCircle2 aria-hidden={true} className="h-4 w-4" />
              Validate cluster
            </Button>
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                void runValidation();
              }}
              loading={validating}
              disabled={validating}
            >
              <RefreshCw aria-hidden={true} className="h-4 w-4" />
              Refresh validation
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                void handleSetDefault();
              }}
              loading={settingDefault}
              disabled={settingDefault || cluster.isDefault}
            >
              <ShieldPlus aria-hidden={true} className="h-4 w-4" />
              Set default cluster
            </Button>
          </div>
        </header>

        <div className="flex-1 space-y-6 overflow-y-auto p-5">
          <section>
            <h3 className="text-sm font-semibold">Connection</h3>
            <dl className="mt-3 grid gap-3">
              <DrawerStat label="Project" value={projectDisplay} />
              <DrawerStat label="Provider" value={cluster.provider} />
              <DrawerStat label="Status" value={cluster.status} />
              <DrawerStat label="Connection type" value={cluster.connectionType || "-"} />
              <DrawerStat label="API endpoint" value={cluster.apiEndpoint || "-"} />
              <DrawerStat label="Version" value={cluster.version || "-"} />
              <DrawerStat label="Region" value={cluster.region || "-"} />
            </dl>
          </section>

          <section>
            <h3 className="text-sm font-semibold">Validation</h3>
            <dl className="mt-3 grid gap-3">
              <DrawerStat label="Last validation" value={formatDateTime(cluster.lastValidatedAt)} />
              <DrawerStat label="Validation error" value={cluster.validationError || "-"} />
              <DrawerStat label="Last discovery" value={formatDateTime(cluster.lastDiscoveryAt)} />
            </dl>
          </section>

          <section>
            <h3 className="text-sm font-semibold">Validation result</h3>
            {validationResult ? (
              <dl className="mt-3 grid gap-3">
                <DrawerStat label="Connected" value={validationResult.connected ? "Yes" : "No"} />
                <DrawerStat label="Version" value={validationResult.clusterVersion || "-"} />
                <DrawerStat label="API server" value={validationResult.apiServerUrl || "-"} />
                <DrawerStat label="Latency" value={`${validationResult.latencyMs} ms`} />
                <DrawerStat label="Validation time" value={formatDateTime(validationResult.validatedAt)} />
                <DrawerStat label="Validation error" value={validationResult.error || "-"} />
              </dl>
            ) : (
              <p className="mt-2 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Run Validate cluster or Refresh validation to view the latest validation result.
              </p>
            )}
          </section>

          <section>
            <h3 className="text-sm font-semibold">Metadata</h3>
            <pre
              className="mt-3 overflow-x-auto rounded-2xl border p-3 text-xs"
              style={{
                borderColor: "var(--border)",
                backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)",
              }}
            >
              {JSON.stringify(cluster.metadata ?? {}, null, 2)}
            </pre>
          </section>

          <section>
            <h3 className="text-sm font-semibold">Timeline</h3>
            <dl className="mt-3 grid gap-3">
              <DrawerStat label="Created" value={formatDateTime(cluster.createdAt)} />
              <DrawerStat label="Updated" value={formatDateTime(cluster.updatedAt)} />
            </dl>
          </section>
        </div>
      </aside>

      <EditClusterModal open={editModalOpen} cluster={cluster} onClose={() => setEditModalOpen(false)} />
      <DeleteClusterDialog
        open={deleteDialogOpen}
        cluster={cluster}
        onClose={() => setDeleteDialogOpen(false)}
        onSuccess={onClose}
      />
    </div>
  );
}

function DrawerFrame({
  ariaLabel,
  children,
  onClose,
}: {
  ariaLabel: string;
  children: React.ReactNode;
  onClose: () => void;
}) {
  return (
    <div
      className="fixed inset-0 z-[60] flex justify-end bg-[color:color-mix(in_srgb,var(--background)_72%,transparent)]"
      role="presentation"
      onMouseDown={onClose}
    >
      <aside
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel}
        onMouseDown={(event) => event.stopPropagation()}
        className="flex h-full w-full max-w-xl flex-col border-l shadow-[var(--shadow-lg)]"
        style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
      >
        {children}
      </aside>
    </div>
  );
}

function DrawerStat({ label, value }: { label: string; value: string }) {
  return (
    <div
      className="rounded-2xl border p-4"
      style={{
        borderColor: "var(--border)",
        backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)",
      }}
    >
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        {label}
      </dt>
      <dd className="mt-1 break-all font-semibold">{value}</dd>
    </div>
  );
}

function formatDateTime(value?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}
