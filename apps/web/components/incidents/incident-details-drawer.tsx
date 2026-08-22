"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  AlertCircle,
  AlertTriangle,
  ClipboardList,
  Compass,
  Files,
  Gauge,
  History,
  ListTree,
  MessageSquare,
  Pencil,
  Trash2,
  X,
  type LucideIcon,
} from "lucide-react";

import { ErrorState, StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useApplication } from "@/hooks/use-applications";
import { useIncident } from "@/hooks/use-incidents";
import { useTeam } from "@/hooks/use-teams";
import { useHasPermission } from "@/store/auth-store";
import {
  INCIDENT_SEVERITY_COLORS,
  INCIDENT_STATUS_LABELS,
  INCIDENT_STATUS_VARIANTS,
  type IncidentSeverity,
  type IncidentStatus,
} from "@/lib/constants";
import { DeleteIncidentDialog } from "./delete-incident-dialog";
import { EditIncidentModal } from "./edit-incident-modal";
import { IncidentAlerts } from "./incident-alerts";
import { IncidentCombinedTimeline } from "./incident-combined-timeline";
import { IncidentComments } from "./incident-comments";
import { IncidentContextChain } from "./incident-context-chain";
import { IncidentLogsPanel } from "./incident-logs-panel";
import { IncidentMetricsPanel } from "./incident-metrics-panel";
import { IncidentSummary } from "./incident-summary";
import { IncidentTimeline } from "./incident-timeline";

const tabs: ReadonlyArray<{ label: string; icon: LucideIcon }> = [
  { label: "Overview", icon: AlertCircle },
  { label: "Context", icon: ListTree },
  { label: "Alerts", icon: AlertTriangle },
  { label: "Metrics", icon: Gauge },
  { label: "Logs", icon: Files },
  { label: "Timeline", icon: History },
  { label: "Comments", icon: MessageSquare },
  { label: "Audit", icon: ClipboardList },
];

/**
 * Incident details drawer, kept for quick preview from the incident grid.
 * Timeline is the combined, meaningful story (Phase 17); Audit is the raw,
 * unfiltered change log. The "Workspace" button opens the full incident
 * investigation console (/incidents/[id]) for deeper cross-evidence work.
 */
export function IncidentDetailsDrawer({
  incidentID,
  projectNameById,
  onClose,
}: {
  incidentID: number | null;
  projectNameById: Map<string, string>;
  onClose: () => void;
}) {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState("Overview");
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const canManageIncidents = useHasPermission("incident:manage");
  const { data: incident, error, isError, isLoading, refetch } = useIncident(incidentID);

  const { data: application } = useApplication(incident?.applicationId ?? null);
  const { data: ownerTeam } = useTeam(incident?.ownerTeamId ?? null);

  useEffect(() => {
    setActiveTab("Overview");
    setEditModalOpen(false);
    setDeleteDialogOpen(false);
  }, [incidentID]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      if (editModalOpen || deleteDialogOpen) return;
      onClose();
    };
    if (incidentID) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [incidentID, onClose, editModalOpen, deleteDialogOpen]);

  if (!incidentID) return null;

  if (isLoading) {
    return (
      <DrawerFrame onClose={onClose} ariaLabel="Loading incident details">
        <div className="flex flex-1 items-center justify-center p-5 text-sm" style={{ color: "var(--muted-foreground)" }}>
          Loading incident details…
        </div>
      </DrawerFrame>
    );
  }

  if (isError || !incident) {
    return (
      <DrawerFrame onClose={onClose} ariaLabel="Incident details unavailable">
        <div className="p-5">
          <ErrorState
            description={error instanceof Error ? error.message : "Unable to load incident details."}
            onRetry={() => {
              void refetch();
            }}
          />
        </div>
      </DrawerFrame>
    );
  }

  const severityColor = getSeverityColor(incident.severity);
  const statusLabel = getStatusLabel(incident.status);
  const statusVariant = getStatusVariant(incident.status);
  const projectName = projectNameById.get(incident.projectId) ?? "Unknown Project";

  const tabListId = "incident-details-tablist";
  const activeTabKey = activeTab.toLowerCase().replace(/\s+/g, "-");
  const activePanelId = `incident-details-panel-${activeTabKey}`;

  return (
    <div
      className="fixed inset-0 z-[60] flex justify-end bg-[color:color-mix(in_srgb,var(--background)_72%,transparent)]"
      role="presentation"
      onMouseDown={onClose}
    >
      <aside
        role="dialog"
        aria-modal="true"
        aria-labelledby="incident-details-title"
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
                  color: severityColor,
                  backgroundColor: `color-mix(in srgb, ${severityColor} 12%, transparent)`,
                }}
              >
                <AlertCircle aria-hidden="true" className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <h2 id="incident-details-title" className="truncate text-xl font-semibold tracking-tight">{incident.title}</h2>
                <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                  Incident workspace
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="secondary"
                onClick={() => router.push(`/incidents/${incident.id}`)}
                className="px-3"
                aria-label="Open full investigation workspace"
              >
                <Compass aria-hidden="true" className="h-4 w-4" />
                <span className="hidden sm:inline">Workspace</span>
              </Button>
              {canManageIncidents ? (
                <>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => setEditModalOpen(true)}
                    className="px-3"
                    aria-label="Edit incident"
                  >
                    <Pencil aria-hidden="true" className="h-4 w-4" />
                    <span className="hidden sm:inline">Edit</span>
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => setDeleteDialogOpen(true)}
                    className="px-3 text-[var(--danger)]"
                    aria-label="Delete incident"
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
                aria-label="Close incident details"
              >
                <X aria-hidden="true" className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <div className="mt-5 flex gap-2">
            <StatusBadge variant={statusVariant}>{statusLabel}</StatusBadge>
            <div
              className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
              style={{
                backgroundColor: `color-mix(in srgb, ${severityColor} 12%, transparent)`,
                color: severityColor,
              }}
            >
              {incident.severity}
            </div>
          </div>
        </header>

        <div className="overflow-x-auto border-b px-3" style={{ borderColor: "var(--border)" }}>
          <div className="flex gap-1" role="tablist" id={tabListId} aria-label="Incident details sections">
            {tabs.map(({ label, icon: TabIcon }, index) => {
              const isActive = activeTab === label;
              const tabSlug = label.toLowerCase().replace(/\s+/g, "-");
              const tabId = `incident-details-tab-${tabSlug}`;
              const panelId = `incident-details-panel-${tabSlug}`;

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
                    const nextTabButton = document.getElementById(`incident-details-tab-${nextTab.label.toLowerCase().replace(/\s+/g, "-")}`);
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
          aria-labelledby={`incident-details-tab-${activeTabKey}`}
          aria-describedby={tabListId}
          className="flex-1 overflow-y-auto p-5"
        >
          {activeTab === "Context" ? (
            <IncidentContextChain incident={incident} />
          ) : activeTab === "Alerts" ? (
            <IncidentAlerts incident={incident} />
          ) : activeTab === "Metrics" ? (
            <IncidentMetricsPanel incident={incident} />
          ) : activeTab === "Logs" ? (
            <IncidentLogsPanel incident={incident} />
          ) : activeTab === "Timeline" ? (
            <IncidentCombinedTimeline incident={incident} />
          ) : activeTab === "Comments" ? (
            <IncidentComments incidentId={incident.id} />
          ) : activeTab === "Audit" ? (
            <IncidentTimeline incident={incident} />
          ) : (
            <IncidentSummary
              incident={incident}
              projectName={projectName}
              applicationName={application?.name ?? "Not set"}
              ownerTeamName={ownerTeam?.name ?? "Not set"}
            />
          )}
        </div>
      </aside>

      {incident && (
        <>
          <EditIncidentModal open={editModalOpen} incident={incident} onClose={() => setEditModalOpen(false)} />
          <DeleteIncidentDialog
            open={deleteDialogOpen}
            incident={incident}
            onClose={() => setDeleteDialogOpen(false)}
            onSuccess={onClose}
          />
        </>
      )}
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

function getSeverityColor(severity: string): string {
  return INCIDENT_SEVERITY_COLORS[severity as IncidentSeverity] || "var(--muted-foreground)";
}

function getStatusLabel(status: string): string {
  return INCIDENT_STATUS_LABELS[status as IncidentStatus] || status;
}

function getStatusVariant(status: string): "info" | "warning" | "success" {
  return INCIDENT_STATUS_VARIANTS[status as IncidentStatus] || "info";
}
