"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Boxes,
  CheckCircle2,
  ClipboardList,
  Cpu,
  Globe,
  Layers3,
  Network,
  Pencil,
  Server,
  ShieldCheck,
  ShieldPlus,
  Trash2,
  X,
  type LucideIcon,
} from "lucide-react";

import { ErrorState } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useCluster, useSetDefaultCluster, useValidateCluster } from "@/hooks/use-clusters";
import { useHasPermission } from "@/store/auth-store";
import type { ClusterValidationResponse } from "@/types/cluster-api";

import { ClusterConfigMaps } from "./cluster-configmaps";
import { ClusterDeployments } from "./cluster-deployments";
import { ClusterIngresses } from "./cluster-ingresses";
import { ClusterNamespaces } from "./cluster-namespaces";
import { ClusterNodes } from "./cluster-nodes";
import { ClusterPods } from "./cluster-pods";
import { ClusterReplicaSets } from "./cluster-replicasets";
import { ClusterSecrets } from "./cluster-secrets";
import { ClusterServices } from "./cluster-services";
import { ClusterProviderBadge, ClusterStatusBadge } from "./cluster-status";
import { DeleteClusterDialog } from "./delete-cluster-dialog";
import { EditClusterModal } from "./edit-cluster-modal";

const tabs: ReadonlyArray<{ label: string; icon: LucideIcon }> = [
  { label: "Overview", icon: CheckCircle2 },
  { label: "Nodes", icon: Cpu },
  { label: "Namespaces", icon: Boxes },
  { label: "Pods", icon: Boxes },
  { label: "Deployments", icon: Server },
  { label: "ReplicaSets", icon: Layers3 },
  { label: "Services", icon: Network },
  { label: "Ingresses", icon: Globe },
  { label: "ConfigMaps", icon: ClipboardList },
  { label: "Secrets", icon: ShieldCheck },
];

export function ClusterDetailsDrawer({
  clusterID,
  projectNameById,
  onClose,
}: {
  clusterID: string | null;
  projectNameById: Map<string, string>;
  onClose: () => void;
}) {
  const [activeTab, setActiveTab] = useState("Overview");
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [validationResult, setValidationResult] = useState<ClusterValidationResponse | null>(null);
  const canManageClusters = useHasPermission("cluster:manage");
  const { data: cluster, error, isError, isLoading, refetch } = useCluster(clusterID);
  const validateCluster = useValidateCluster();
  const setDefaultCluster = useSetDefaultCluster();

  useEffect(() => {
    setActiveTab("Overview");
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

  const tabListId = "cluster-details-tablist";
  const activeTabKey = activeTab.toLowerCase();
  const activePanelId = `cluster-details-panel-${activeTabKey}`;

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
              {canManageClusters ? (
                <>
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
                </>
              ) : null}
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

          {canManageClusters ? (
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
          ) : null}
        </header>

        <div className="overflow-x-auto border-b px-3" style={{ borderColor: "var(--border)" }}>
          <div className="flex gap-1" role="tablist" id={tabListId} aria-label="Cluster details sections">
            {tabs.map(({ label, icon: TabIcon }, index) => {
              const isActive = activeTab === label;
              const tabId = `cluster-details-tab-${label.toLowerCase()}`;
              const panelId = `cluster-details-panel-${label.toLowerCase()}`;

              return (
                <button
                  key={label}
                  id={tabId}
                  type="button"
                  role="tab"
                  tabIndex={isActive ? 0 : -1}
                  aria-selected={isActive}
                  aria-controls={panelId}
                  onClick={() => setActiveTab(label)}
                  onKeyDown={(event) => {
                    if (!"ArrowLeft ArrowRight Home End".includes(event.key)) return;
                    event.preventDefault();

                    const maxIndex = tabs.length - 1;
                    let nextIndex = index;
                    if (event.key === "ArrowRight") nextIndex = index === maxIndex ? 0 : index + 1;
                    if (event.key === "ArrowLeft") nextIndex = index === 0 ? maxIndex : index - 1;
                    if (event.key === "Home") nextIndex = 0;
                    if (event.key === "End") nextIndex = maxIndex;

                    const nextTab = tabs[nextIndex];
                    if (!nextTab) return;
                    setActiveTab(nextTab.label);
                    const nextTabButton = document.getElementById(`cluster-details-tab-${nextTab.label.toLowerCase()}`);
                    nextTabButton?.focus();
                  }}
                  className="inline-flex shrink-0 items-center gap-1.5 border-b-2 px-3 py-3 text-sm font-medium transition-colors"
                  style={{
                    color: isActive ? "var(--primary)" : "var(--muted-foreground)",
                    borderColor: isActive ? "var(--primary)" : "transparent",
                  }}
                >
                  <TabIcon aria-hidden="true" className="h-3.5 w-3.5" />
                  {label}
                </button>
              );
            })}
          </div>
        </div>

        <div
          id={activePanelId}
          role="tabpanel"
          aria-labelledby={`cluster-details-tab-${activeTabKey}`}
          aria-describedby={tabListId}
          className="flex-1 overflow-y-auto p-5"
        >
          {activeTab === "Nodes" ? (
            <ClusterNodes clusterId={cluster.id} />
          ) : activeTab === "Namespaces" ? (
            <ClusterNamespaces clusterId={cluster.id} />
          ) : activeTab === "Pods" ? (
            <ClusterPods clusterId={cluster.id} />
          ) : activeTab === "Deployments" ? (
            <ClusterDeployments clusterId={cluster.id} />
          ) : activeTab === "ReplicaSets" ? (
            <ClusterReplicaSets clusterId={cluster.id} />
          ) : activeTab === "Services" ? (
            <ClusterServices clusterId={cluster.id} />
          ) : activeTab === "Ingresses" ? (
            <ClusterIngresses clusterId={cluster.id} />
          ) : activeTab === "ConfigMaps" ? (
            <ClusterConfigMaps clusterId={cluster.id} />
          ) : activeTab === "Secrets" ? (
            <ClusterSecrets clusterId={cluster.id} />
          ) : (
            <div className="space-y-6">
          <section>
            <h3 className="text-sm font-semibold">Connection</h3>
            <dl className="mt-3 grid gap-3">
              <DrawerStat label="Project" value={projectDisplay} />
              <DrawerStat label="Provider" value={cluster.provider} />
              <DrawerStat label="Status" value={cluster.status} />
              <DrawerStat label="Connection type" value={cluster.connectionType || "-"} />
              <DrawerStat label="API endpoint" value={cluster.apiEndpoint || "-"} />
              <DrawerStat label="Kubernetes version" value={cluster.kubernetesVersion || "-"} />
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
                <DrawerStat label="Kubernetes version" value={validationResult.kubernetesVersion || "-"} />
                <DrawerStat label="API server" value={validationResult.apiServerUrl || "-"} />
                <DrawerStat label="Latency" value={`${validationResult.latencyMs} ms`} />
                <DrawerStat label="Validation time" value={formatDateTime(validationResult.validatedAt)} />
                <DrawerStat label="Validation error" value={validationResult.error || "-"} />
              </dl>
            ) : (
              <p className="mt-2 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Run Validate cluster to check live connectivity, Kubernetes version, and API reachability.
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
          )}
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
