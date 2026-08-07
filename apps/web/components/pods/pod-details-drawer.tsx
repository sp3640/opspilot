"use client";

import { useEffect, useState } from "react";
import { Activity, Boxes, Cpu, List, type LucideIcon, X } from "lucide-react";

import { StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";
import type { PodResponse } from "@/types/pod-api";

import { PodLogs } from "./pod-logs";

const tabs: ReadonlyArray<{ label: string; icon: LucideIcon }> = [
  { label: "Overview", icon: Activity },
  { label: "Logs", icon: List },
  { label: "Events", icon: Cpu },
];

/** Right-side pod context. Only Overview renders real content today. */
export function PodDetailsDrawer({
  pod,
  applicationId,
  open,
  onClose,
}: {
  pod: PodResponse | null;
  applicationId: string;
  open: boolean;
  onClose: () => void;
}) {
  const [activeTab, setActiveTab] = useState("Overview");

  useEffect(() => {
    setActiveTab("Overview");
  }, [pod?.namespace, pod?.name]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    if (open) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [open, onClose]);

  if (!open || !pod) return null;

  const titleId = "pod-details-title";
  const tabListId = "pod-details-tablist";
  const activeTabKey = activeTab.toLowerCase().replace(/\s+/g, "-");
  const activePanelId = `pod-details-panel-${activeTabKey}`;

  const overviewStats = [
    { label: "Name", value: pod.name },
    { label: "Namespace", value: pod.namespace },
    { label: "Phase", value: pod.phase },
    { label: "Node", value: pod.nodeName || "-" },
    { label: "Pod IP", value: pod.podIP || "-" },
    { label: "Host IP", value: pod.hostIP || "-" },
    { label: "Restart count", value: String(pod.restartCount) },
    { label: "Start time", value: formatDateTime(pod.startTime) },
    { label: "Age", value: pod.age || "-" },
  ];

  const labelEntries = Object.entries(pod.labels ?? {});
  const ownerReferences = pod.ownerReferences ?? [];
  const containerStatuses = pod.containerStatuses ?? [];
  const conditions = pod.conditions ?? [];

  const placeholderMessage =
    activeTab === "Logs" ? "Pod logs will be implemented next." : "Pod events will be implemented next.";

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
                <Boxes aria-hidden="true" className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <h2 id={titleId} className="truncate text-xl font-semibold tracking-tight">{pod.name}</h2>
                <p className="mt-1 truncate text-sm" style={{ color: "var(--muted-foreground)" }}>
                  {pod.namespace}
                </p>
              </div>
            </div>
            <Button
              type="button"
              variant="ghost"
              onClick={onClose}
              className="h-9 w-9 rounded-xl p-0"
              aria-label="Close pod details"
            >
              <X aria-hidden="true" className="h-4 w-4" />
            </Button>
          </div>
          <div className="mt-5 flex gap-2">
            <StatusBadge variant={getStatusVariant(pod.phase)}>{pod.phase}</StatusBadge>
            <span
              className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
              style={{
                color: "var(--muted-foreground)",
                backgroundColor: "color-mix(in srgb, var(--muted-foreground) 12%, transparent)",
              }}
            >
              {pod.ready ? "Ready" : "Not ready"}
            </span>
          </div>
        </header>

        <div className="overflow-x-auto border-b px-3" style={{ borderColor: "var(--border)" }}>
          <div className="flex gap-1" role="tablist" id={tabListId} aria-label="Pod details sections">
            {tabs.map(({ label, icon: TabIcon }, index) => {
              const isActive = activeTab === label;
              const tabSlug = label.toLowerCase().replace(/\s+/g, "-");
              const tabId = `pod-details-tab-${tabSlug}`;
              const panelId = `pod-details-panel-${tabSlug}`;

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
                    const nextTabButton = document.getElementById(`pod-details-tab-${nextTab.label.toLowerCase().replace(/\s+/g, "-")}`);
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
          aria-labelledby={`pod-details-tab-${activeTabKey}`}
          aria-describedby={tabListId}
          className="flex-1 overflow-y-auto p-5"
        >
          {activeTab === "Overview" ? (
            <div className="space-y-6">
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

              <section>
                <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
                  Labels
                </h3>
                {labelEntries.length === 0 ? (
                  <p className="mt-2 text-sm" style={{ color: "var(--muted-foreground)" }}>No labels</p>
                ) : (
                  <div className="mt-2 flex flex-wrap gap-2">
                    {labelEntries.map(([key, value]) => (
                      <span
                        key={key}
                        className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
                        style={{
                          color: "var(--muted-foreground)",
                          backgroundColor: "color-mix(in srgb, var(--muted-foreground) 12%, transparent)",
                        }}
                      >
                        {key}={value}
                      </span>
                    ))}
                  </div>
                )}
              </section>

              <section>
                <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
                  Owner references
                </h3>
                {ownerReferences.length === 0 ? (
                  <p className="mt-2 text-sm" style={{ color: "var(--muted-foreground)" }}>No owner references</p>
                ) : (
                  <ul className="mt-2 space-y-2">
                    {ownerReferences.map((ref) => (
                      <li
                        key={ref.uid}
                        className="rounded-2xl border p-3 text-sm"
                        style={{ borderColor: "var(--border)" }}
                      >
                        <p className="font-medium">{ref.kind} · {ref.name}</p>
                        <p className="mt-1 break-all text-xs" style={{ color: "var(--muted-foreground)" }}>{ref.apiVersion}</p>
                      </li>
                    ))}
                  </ul>
                )}
              </section>

              <section>
                <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
                  Container statuses
                </h3>
                {containerStatuses.length === 0 ? (
                  <p className="mt-2 text-sm" style={{ color: "var(--muted-foreground)" }}>No container statuses</p>
                ) : (
                  <ul className="mt-2 space-y-2">
                    {containerStatuses.map((container) => (
                      <li
                        key={container.name}
                        className="rounded-2xl border p-3 text-sm"
                        style={{ borderColor: "var(--border)" }}
                      >
                        <div className="flex items-start justify-between gap-3">
                          <p className="min-w-0 truncate font-medium">{container.name}</p>
                          <StatusBadge variant={container.ready ? "success" : "critical"}>
                            {container.ready ? "Ready" : "Not ready"}
                          </StatusBadge>
                        </div>
                        <p className="mt-1 break-all text-xs" style={{ color: "var(--muted-foreground)" }}>{container.image}</p>
                        <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                          State: {container.state} · Restarts: {container.restartCount}
                        </p>
                      </li>
                    ))}
                  </ul>
                )}
              </section>

              <section>
                <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
                  Conditions
                </h3>
                {conditions.length === 0 ? (
                  <p className="mt-2 text-sm" style={{ color: "var(--muted-foreground)" }}>No conditions</p>
                ) : (
                  <ul className="mt-2 space-y-2">
                    {conditions.map((condition) => (
                      <li
                        key={condition.type}
                        className="rounded-2xl border p-3 text-sm"
                        style={{ borderColor: "var(--border)" }}
                      >
                        <p className="font-medium">{condition.type}: {condition.status}</p>
                        {condition.reason ? (
                          <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>{condition.reason}</p>
                        ) : null}
                        {condition.message ? (
                          <p className="mt-1 break-all text-xs" style={{ color: "var(--muted-foreground)" }}>{condition.message}</p>
                        ) : null}
                      </li>
                    ))}
                  </ul>
                )}
              </section>
            </div>
          ) : activeTab === "Logs" ? (
            <PodLogs applicationId={applicationId} pod={pod} />
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

function getStatusVariant(phase: string): "success" | "warning" | "critical" | "info" {
  switch (phase) {
    case "Running":
    case "Succeeded":
      return "success";
    case "Pending":
      return "warning";
    case "Failed":
      return "critical";
    default:
      return "info";
  }
}

function formatDateTime(value?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}
