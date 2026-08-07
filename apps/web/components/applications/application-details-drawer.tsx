"use client";

import { useEffect, useState } from "react";
import {
  Activity,
  Boxes,
  ClipboardList,
  Cpu,
  Layers3,
  List,
  Network,
  Rocket,
  Server,
  ShieldCheck,
  type LucideIcon,
  X,
} from "lucide-react";

import { StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";
import type { ApplicationResponse } from "@/types/application-api";

import { ApplicationConfigMaps } from "./application-configmaps";
import { ApplicationDeployments } from "./application-deployments";
import { ApplicationEvents } from "./application-events";
import { ApplicationPods } from "./application-pods";
import { ApplicationReplicaSets } from "./application-replicasets";
import { ApplicationRuntimeDeployments } from "./application-runtime-deployments";
import { ApplicationSecrets } from "./application-secrets";
import { ApplicationServices } from "./application-services";

const tabs: ReadonlyArray<{ label: string; icon: LucideIcon }> = [
  { label: "Overview", icon: Activity },
  { label: "Deployments", icon: Rocket },
  { label: "Pods", icon: Boxes },
  { label: "Services", icon: Network },
  { label: "ConfigMaps", icon: ClipboardList },
  { label: "Secrets", icon: ShieldCheck },
  { label: "ReplicaSets", icon: Layers3 },
  { label: "Runtime Deployments", icon: Server },
  { label: "Events", icon: Cpu },
  { label: "Logs", icon: List },
];

/** Right-side application context. Only Overview renders real content today. */
export function ApplicationDetailsDrawer({
  application,
  open,
  onClose,
}: {
  application: ApplicationResponse | null;
  open: boolean;
  onClose: () => void;
}) {
  const [activeTab, setActiveTab] = useState("Overview");

  useEffect(() => {
    setActiveTab("Overview");
  }, [application?.id]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    if (open) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [open, onClose]);

  if (!open || !application) return null;

  const titleId = "application-details-title";
  const tabListId = "application-details-tablist";
  const activeTabKey = activeTab.toLowerCase().replace(/\s+/g, "-");
  const activePanelId = `application-details-panel-${activeTabKey}`;

  const overviewStats = [
    { label: "Name", value: application.name },
    { label: "Runtime", value: application.runtime },
    { label: "Status", value: application.status },
    { label: "Repository URL", value: application.repositoryUrl || "-" },
    { label: "Default branch", value: application.defaultBranch || "-" },
    { label: "Port", value: String(application.port) },
    { label: "Environment", value: application.environment || "-" },
    { label: "Created", value: formatDate(application.createdAt) },
    { label: "Updated", value: formatDate(application.updatedAt) },
  ];

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
                <h2 id={titleId} className="truncate text-xl font-semibold tracking-tight">{application.name}</h2>
                <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                  Application workspace
                </p>
              </div>
            </div>
            <Button
              type="button"
              variant="ghost"
              onClick={onClose}
              className="h-9 w-9 rounded-xl p-0"
              aria-label="Close application details"
            >
              <X aria-hidden="true" className="h-4 w-4" />
            </Button>
          </div>
          <div className="mt-5 flex gap-2">
            <StatusBadge variant={getStatusVariant(application.status)}>{application.status}</StatusBadge>
            <span
              className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
              style={{
                color: "var(--muted-foreground)",
                backgroundColor: "color-mix(in srgb, var(--muted-foreground) 12%, transparent)",
              }}
            >
              {application.runtime}
            </span>
          </div>
        </header>

        <div className="overflow-x-auto border-b px-3" style={{ borderColor: "var(--border)" }}>
          <div className="flex gap-1" role="tablist" id={tabListId} aria-label="Application details sections">
            {tabs.map(({ label, icon: TabIcon }, index) => {
              const isActive = activeTab === label;
              const tabSlug = label.toLowerCase().replace(/\s+/g, "-");
              const tabId = `application-details-tab-${tabSlug}`;
              const panelId = `application-details-panel-${tabSlug}`;

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
                    const nextTabButton = document.getElementById(`application-details-tab-${nextTab.label.toLowerCase().replace(/\s+/g, "-")}`);
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
          aria-labelledby={`application-details-tab-${activeTabKey}`}
          aria-describedby={tabListId}
          className="flex-1 overflow-y-auto p-5"
        >
          {activeTab === "Overview" ? (
            <div className="space-y-6">
              <div>
                <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
                  Description
                </h3>
                <p className="mt-2 text-sm leading-6">
                  {application.description || "No description provided"}
                </p>
              </div>
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
            </div>
          ) : activeTab === "Deployments" ? (
            <ApplicationDeployments applicationId={application.id} />
          ) : activeTab === "Pods" ? (
            <ApplicationPods applicationId={application.id} />
          ) : activeTab === "Services" ? (
            <ApplicationServices applicationId={application.id} />
          ) : activeTab === "ConfigMaps" ? (
            <ApplicationConfigMaps applicationId={application.id} />
          ) : activeTab === "Secrets" ? (
            <ApplicationSecrets applicationId={application.id} />
          ) : activeTab === "ReplicaSets" ? (
            <ApplicationReplicaSets applicationId={application.id} />
          ) : activeTab === "Runtime Deployments" ? (
            <ApplicationRuntimeDeployments applicationId={application.id} />
          ) : activeTab === "Events" ? (
            <ApplicationEvents applicationId={application.id} />
          ) : (
            <div
              className="flex min-h-56 items-center justify-center rounded-2xl border p-6 text-center text-sm"
              style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}
            >
              {activeTab} data is not available for {application.name}.
            </div>
          )}
        </div>
      </aside>
    </div>
  );
}

function getStatusVariant(status: string): "success" | "info" | "archived" {
  switch (status) {
    case "Ready":
      return "success";
    case "Archived":
      return "archived";
    default:
      return "info";
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
