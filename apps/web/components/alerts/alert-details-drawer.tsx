"use client";

import { useEffect, useState } from "react";
import {
  Activity,
  Check,
  FileText,
  Gauge,
  History,
  Link2,
  ListTree,
  Pencil,
  Plus,
  RotateCcw,
  Share2,
  Trash2,
  X,
  type LucideIcon,
} from "lucide-react";
import { toast } from "sonner";

import { ErrorState } from "@/components/common";
import { Button } from "@/components/ui/button";
import {
  useAcknowledgeAlert,
  useAlert,
  useAttachAlertIncident,
  useReopenAlert,
  useResolveAlert,
} from "@/hooks/use-alerts";
import { useIncidents } from "@/hooks/use-incidents";
import { incidentService } from "@/services/incident-service";
import { useHasPermission } from "@/store/auth-store";
import type { ProjectResponse } from "@/types/project-api";

import { AlertContextChain } from "./alert-context-chain";
import { AlertLogsPanel } from "./alert-logs-panel";
import { AlertMetricsPanel } from "./alert-metrics-panel";
import { AlertRelated } from "./alert-related";
import { AlertSeverityBadge } from "./alert-severity-badge";
import { AlertStatusBadge } from "./alert-status-badge";
import { AlertTimeline } from "./alert-timeline";
import { DeleteAlertDialog } from "./delete-alert-dialog";
import { EditAlertModal } from "./edit-alert-modal";

const tabs: ReadonlyArray<{ label: string; icon: LucideIcon }> = [
  { label: "Overview", icon: Activity },
  { label: "Context", icon: ListTree },
  { label: "Metrics", icon: Gauge },
  { label: "Logs", icon: FileText },
  { label: "Timeline", icon: History },
  { label: "Related", icon: Share2 },
];

/**
 * Alert details, restructured (Phase 15) so an engineer can investigate
 * without leaving this drawer: Overview keeps the original acknowledge/
 * resolve/reopen/attach-incident actions and permission gates completely
 * unchanged; the new tabs add the affected-resource chain, relevant
 * metrics/logs, a real audit-log-backed timeline, and related alerts/
 * incidents - each only ever shown when a real identifier supports it.
 */
export function AlertDetailsDrawer({
  alertID,
  projects,
  projectNameById,
  onClose,
}: {
  alertID: number | null;
  projects: ProjectResponse[];
  projectNameById: Map<string, string>;
  onClose: () => void;
}) {
  const { data: alert, isLoading, isError, error, refetch } = useAlert(alertID);
  const [activeTab, setActiveTab] = useState("Overview");
  const [edit, setEdit] = useState(false);
  const [remove, setRemove] = useState(false);
  const [linking, setLinking] = useState(false);

  const { data: incidentsData } = useIncidents({ page: 1, limit: 100, sort: "updated_at", order: "desc", projectId: alert?.projectId });
  const acknowledge = useAcknowledgeAlert();
  const resolve = useResolveAlert();
  const reopen = useReopenAlert();
  const attach = useAttachAlertIncident();
  const canManageAlerts = useHasPermission("alert:manage");

  useEffect(() => {
    setActiveTab("Overview");
    setEdit(false);
    setRemove(false);
    setLinking(false);
  }, [alertID]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    if (alertID) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [alertID, onClose]);

  if (!alertID) return null;

  const tabListId = "alert-details-tablist";
  const activeTabKey = activeTab.toLowerCase().replace(/\s+/g, "-");
  const activePanelId = `alert-details-panel-${activeTabKey}`;

  return (
    <div className="fixed inset-0 z-50 bg-black/45" role="presentation" onMouseDown={onClose}>
      <aside
        role="dialog"
        aria-modal="true"
        aria-label="Alert details"
        onMouseDown={(event) => event.stopPropagation()}
        className="ml-auto flex h-full w-full max-w-xl flex-col border-l"
        style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
      >
        <header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="min-w-0">
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Alert details</p>
            <h2 className="mt-1 truncate text-xl font-semibold">{alert?.title || "Loading alert"}</h2>
            {alert ? (
              <div className="mt-3 flex flex-wrap gap-2">
                <AlertSeverityBadge severity={alert.severity} />
                <AlertStatusBadge status={alert.status} />
              </div>
            ) : null}
          </div>
          <Button type="button" variant="ghost" className="h-9 w-9 p-0" onClick={onClose} aria-label="Close alert details">
            <X className="h-4 w-4" />
          </Button>
        </header>

        {isError || !alert ? (
          <div className="p-5">
            {isLoading ? (
              <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Loading alert…</p>
            ) : (
              <ErrorState
                description={error instanceof Error ? error.message : "Unable to load alert."}
                onRetry={() => {
                  void refetch();
                }}
              />
            )}
          </div>
        ) : (
          <>
            <div className="overflow-x-auto border-b px-3" style={{ borderColor: "var(--border)" }}>
              <div className="flex gap-1" role="tablist" id={tabListId} aria-label="Alert details sections">
                {tabs.map(({ label, icon: TabIcon }, index) => {
                  const isActive = activeTab === label;
                  const tabSlug = label.toLowerCase().replace(/\s+/g, "-");
                  const tabId = `alert-details-tab-${tabSlug}`;
                  const panelId = `alert-details-panel-${tabSlug}`;

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
                        const nextTabButton = document.getElementById(`alert-details-tab-${nextTab.label.toLowerCase().replace(/\s+/g, "-")}`);
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
              aria-labelledby={`alert-details-tab-${activeTabKey}`}
              aria-describedby={tabListId}
              className="flex-1 overflow-y-auto p-5"
            >
              {activeTab === "Context" ? (
                <AlertContextChain alert={alert} />
              ) : activeTab === "Metrics" ? (
                <AlertMetricsPanel alert={alert} />
              ) : activeTab === "Logs" ? (
                <AlertLogsPanel alert={alert} />
              ) : activeTab === "Timeline" ? (
                <AlertTimeline alert={alert} />
              ) : activeTab === "Related" ? (
                <AlertRelated alert={alert} />
              ) : (
                <div className="space-y-6">
                  <Details
                    title="General"
                    values={[
                      ["Title", alert.title],
                      ["Message", alert.description],
                      ["Rule", alert.source],
                      ["Occurrences", String(alert.occurrenceCount)],
                      ["Fingerprint", alert.fingerprint],
                    ]}
                  />
                  <Details
                    title="Resource"
                    values={[
                      ["Resource", `${alert.resourceType}: ${alert.resourceId}`],
                      ["Project", projectNameById.get(alert.projectId) || alert.projectId],
                    ]}
                  />
                  <Details
                    title="Acknowledgement"
                    values={[["Acknowledged", alert.acknowledgedAt ? time(alert.acknowledgedAt) : "Not acknowledged"]]}
                  />
                  <Details title="Resolution" values={[["Resolved", alert.resolvedAt ? time(alert.resolvedAt) : "Not resolved"]]} />

                  <section>
                    <h3 className="text-sm font-semibold">Incident linkage</h3>
                    {alert.incidentId ? (
                      <div className="mt-3 flex items-center justify-between rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
                        <span className="text-sm">Incident #{alert.incidentId}</span>
                        <Button
                          type="button"
                          variant="secondary"
                          className="px-3 py-2"
                          onClick={() => {
                            window.location.assign("/incidents");
                          }}
                        >
                          Open incident
                        </Button>
                      </div>
                    ) : canManageAlerts ? (
                      <div className="mt-3 flex flex-wrap gap-2">
                        <Button type="button" variant="secondary" onClick={() => setLinking(true)}>
                          <Link2 className="h-4 w-4" />
                          Link existing incident
                        </Button>
                        <Button
                          type="button"
                          variant="secondary"
                          onClick={async () => {
                            try {
                              const incident = await incidentService.createIncident({
                                title: alert.title,
                                description: alert.description,
                                severity: severityToIncident(alert.severity),
                                status: "OPEN",
                                project_id: alert.projectId,
                              });
                              await attach.mutateAsync({ id: alert.id, incidentId: incident.id });
                              toast.success("Incident created and linked to alert.");
                            } catch {
                              toast.error("Unable to create and link incident.");
                            }
                          }}
                          loading={attach.isPending}
                        >
                          <Plus className="h-4 w-4" />
                          Create incident
                        </Button>
                      </div>
                    ) : null}
                  </section>

                  {linking ? (
                    <section className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
                      <label className="text-sm font-medium">Choose incident</label>
                      <div className="mt-2 flex gap-2">
                        <select
                          id="alert-link-incident"
                          className="h-10 flex-1 rounded-xl border bg-transparent px-3 text-sm"
                          style={{ borderColor: "var(--border)" }}
                        >
                          <option value="">Select incident</option>
                          {(incidentsData?.items ?? []).map((incident) => (
                            <option key={incident.id} value={incident.id}>{incident.title}</option>
                          ))}
                        </select>
                        <Button
                          type="button"
                          onClick={() => {
                            const element = document.getElementById("alert-link-incident") as HTMLSelectElement | null;
                            const id = Number(element?.value);
                            if (id) {
                              void attach.mutateAsync({ id: alert.id, incidentId: id }).then(() => setLinking(false));
                            }
                          }}
                          loading={attach.isPending}
                        >
                          Link
                        </Button>
                      </div>
                    </section>
                  ) : null}

                  <section>
                    <h3 className="text-sm font-semibold">Timestamps</h3>
                    <dl className="mt-3 grid gap-3 sm:grid-cols-2">
                      <Item label="First seen" value={time(alert.firstSeenAt)} />
                      <Item label="Last seen" value={time(alert.lastSeenAt)} />
                      <Item label="Created" value={time(alert.createdAt)} />
                      <Item label="Updated" value={time(alert.updatedAt)} />
                      <Item label="Acknowledged" value={alert.acknowledgedAt ? time(alert.acknowledgedAt) : "-"} />
                      <Item label="Resolved" value={alert.resolvedAt ? time(alert.resolvedAt) : "-"} />
                    </dl>
                  </section>

                  <Json title="Metadata" value={alert.metadata} />
                </div>
              )}
            </div>

            <footer className="flex flex-wrap gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
              {canManageAlerts ? (
                <>
                  <Button
                    variant="secondary"
                    onClick={() => void acknowledge.mutateAsync(alert.id)}
                    disabled={alert.status === "ACKNOWLEDGED" || alert.status === "RESOLVED"}
                    loading={acknowledge.isPending}
                  >
                    <Check className="h-4 w-4" />
                    Acknowledge
                  </Button>
                  <Button
                    variant="secondary"
                    onClick={() => void resolve.mutateAsync(alert.id)}
                    disabled={alert.status === "RESOLVED"}
                    loading={resolve.isPending}
                  >
                    <Check className="h-4 w-4" />
                    Resolve
                  </Button>
                  <Button
                    variant="secondary"
                    onClick={() => void reopen.mutateAsync(alert.id)}
                    disabled={alert.status !== "RESOLVED"}
                    loading={reopen.isPending}
                  >
                    <RotateCcw className="h-4 w-4" />
                    Reopen
                  </Button>
                  <Button variant="secondary" onClick={() => setEdit(true)}>
                    <Pencil className="h-4 w-4" />
                    Edit
                  </Button>
                  <Button variant="danger" onClick={() => setRemove(true)}>
                    <Trash2 className="h-4 w-4" />
                    Delete
                  </Button>
                </>
              ) : null}
            </footer>

            <EditAlertModal open={edit} alert={alert} projects={projects} onClose={() => setEdit(false)} />
            <DeleteAlertDialog open={remove} alert={alert} onClose={() => setRemove(false)} onSuccess={onClose} />
          </>
        )}
      </aside>
    </div>
  );
}

function Details({ title, values }: { title: string; values: [string, string][] }) {
  return (
    <section>
      <h3 className="text-sm font-semibold">{title}</h3>
      <dl className="mt-3 grid gap-3 sm:grid-cols-2">
        {values.map(([label, value]) => (
          <Item key={label} label={label} value={value || "-"} />
        ))}
      </dl>
    </section>
  );
}

function Item({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</dt>
      <dd className="mt-1 break-words text-sm">{value}</dd>
    </div>
  );
}

function Json({ title, value }: { title: string; value: Record<string, unknown> }) {
  return (
    <section>
      <h3 className="text-sm font-semibold">{title}</h3>
      <pre
        className="mt-3 overflow-x-auto rounded-2xl border p-3 text-xs"
        style={{ borderColor: "var(--border)", backgroundColor: "var(--muted)" }}
      >
        {JSON.stringify(value ?? {}, null, 2)}
      </pre>
    </section>
  );
}

function time(value: string) {
  return new Date(value).toLocaleString();
}

function severityToIncident(severity: string) {
  return ({ CRITICAL: "P0", HIGH: "P1", MEDIUM: "P2", LOW: "P3", INFO: "P4" } as Record<string, string>)[severity] ?? "P2";
}
