"use client";

import { useEffect, useState } from "react";
import { Activity, Ban, ClipboardList, RotateCcw, Rocket, type LucideIcon, X } from "lucide-react";

import { StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useHasPermission } from "@/store/auth-store";
import type { DeploymentResponse } from "@/types/deployment-api";

import { DeploymentCancel } from "./deployment-cancel";
import { DeploymentHistory } from "./deployment-history";
import { DeploymentRollback } from "./deployment-rollback";

const allTabs: ReadonlyArray<{ label: string; icon: LucideIcon }> = [
  { label: "Overview", icon: Activity },
  { label: "History", icon: ClipboardList },
  { label: "Rollback", icon: RotateCcw },
  { label: "Cancel", icon: Ban },
];

/** Right-side deployment context. Only Overview renders real content today. */
export function DeploymentDetailsDrawer({
  deployment,
  open,
  onClose,
}: {
  deployment: DeploymentResponse | null;
  open: boolean;
  onClose: () => void;
}) {
  const canManageDeployments = useHasPermission("deployment:manage");
  const tabs = canManageDeployments ? allTabs : allTabs.filter((tab) => tab.label !== "Rollback" && tab.label !== "Cancel");
  const [activeTab, setActiveTab] = useState("Overview");

  useEffect(() => {
    setActiveTab("Overview");
  }, [deployment?.id]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    if (open) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [open, onClose]);

  if (!open || !deployment) return null;

  const titleId = "deployment-details-title";
  const tabListId = "deployment-details-tablist";
  const activeTabKey = activeTab.toLowerCase().replace(/\s+/g, "-");
  const activePanelId = `deployment-details-panel-${activeTabKey}`;

  const overviewStats = [
    { label: "Image", value: deployment.image },
    { label: "Image tag", value: deployment.imageTag || "-" },
    { label: "Status", value: deployment.status },
    { label: "Environment", value: deployment.environment || "-" },
    { label: "Target cluster", value: deployment.targetClusterId },
    { label: "Started", value: formatDateTime(deployment.startedAt) },
    { label: "Completed", value: formatDateTime(deployment.completedAt) },
    { label: "Created", value: formatDate(deployment.createdAt) },
    { label: "Updated", value: formatDate(deployment.updatedAt) },
  ];

  const placeholderMessage =
    activeTab === "History"
      ? "Deployment history will be implemented next."
      : activeTab === "Rollback"
        ? "Rollback functionality will be implemented next."
        : "Deployment cancellation will be implemented next.";

  return (
    <div
      className="fixed inset-0 z-[60] flex justify-end bg-[color:color-mix(in_srgb,var(--background)_72%,transparent)]"
      role="presentation"
      onMouseDown={onClose}
    >
      <aside
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        onMouseDown={(event) => event.stopPropagation()}
        className="flex h-full w-full max-w-xl flex-col border-l shadow-[var(--shadow-lg)]"
        style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
      >
        <header className="border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 items-center gap-3">
              <div
                className="rounded-2xl p-3"
                style={{
                  color: "var(--primary)",
                  backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)",
                }}
              >
                <Rocket aria-hidden="true" className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <h2 id={titleId} className="truncate text-xl font-semibold tracking-tight">{deploymentName(deployment)}</h2>
                <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                  Deployment workspace
                </p>
              </div>
            </div>
            <Button
              type="button"
              variant="ghost"
              onClick={onClose}
              className="h-9 w-9 rounded-xl p-0"
              aria-label="Close deployment details"
            >
              <X aria-hidden="true" className="h-4 w-4" />
            </Button>
          </div>
          <div className="mt-5 flex gap-2">
            <StatusBadge variant={getStatusVariant(deployment.status)}>{deployment.status}</StatusBadge>
            <span
              className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
              style={{
                color: "var(--muted-foreground)",
                backgroundColor: "color-mix(in srgb, var(--muted-foreground) 12%, transparent)",
              }}
            >
              {deployment.environment}
            </span>
          </div>
        </header>

        <div className="overflow-x-auto border-b px-3" style={{ borderColor: "var(--border)" }}>
          <div className="flex gap-1" role="tablist" id={tabListId} aria-label="Deployment details sections">
            {tabs.map(({ label, icon: TabIcon }, index) => {
              const isActive = activeTab === label;
              const tabSlug = label.toLowerCase().replace(/\s+/g, "-");
              const tabId = `deployment-details-tab-${tabSlug}`;
              const panelId = `deployment-details-panel-${tabSlug}`;

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
                    const nextTabButton = document.getElementById(`deployment-details-tab-${nextTab.label.toLowerCase().replace(/\s+/g, "-")}`);
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
          aria-labelledby={`deployment-details-tab-${activeTabKey}`}
          aria-describedby={tabListId}
          className="flex-1 overflow-y-auto p-5"
        >
          {activeTab === "Overview" ? (
            <dl className="grid grid-cols-2 gap-3">
              {overviewStats.map((item) => (
                <div
                  key={item.label}
                  className="rounded-2xl border p-4"
                  style={{
                    borderColor: "var(--border)",
                    backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)",
                  }}
                >
                  <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{item.label}</dt>
                  <dd className="mt-1 break-all font-semibold">{item.value}</dd>
                </div>
              ))}
            </dl>
          ) : activeTab === "History" ? (
            <DeploymentHistory deploymentId={deployment.id} />
          ) : activeTab === "Rollback" ? (
            <DeploymentRollback deployment={deployment} />
          ) : activeTab === "Cancel" ? (
            <DeploymentCancel deployment={deployment} />
          ) : (
            <div
              className="flex min-h-56 items-center justify-center rounded-2xl border p-6 text-center text-sm"
              style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}
            >
              {placeholderMessage}
            </div>
          )}
        </div>
      </aside>
    </div>
  );
}

function deploymentName(deployment: DeploymentResponse): string {
  return deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image;
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

function formatDateTime(value?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}
