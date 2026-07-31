"use client";

import { useEffect, useState } from "react";
import { AlertCircle, Pencil, Trash2, X } from "lucide-react";

import { ErrorState, StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useIncident } from "@/hooks/use-incidents";
import {
  INCIDENT_SEVERITY_COLORS,
  INCIDENT_STATUS_LABELS,
  INCIDENT_STATUS_VARIANTS,
  type IncidentSeverity,
  type IncidentStatus,
} from "@/lib/constants";
import { DeleteIncidentDialog } from "./delete-incident-dialog";
import { EditIncidentModal } from "./edit-incident-modal";

/** Read-only incident context loaded from the Incident detail API. */
export function IncidentDetailsDrawer({
  incidentID,
  projectNameById,
  onClose,
}: {
  incidentID: number | null;
  projectNameById: Map<string, string>;
  onClose: () => void;
}) {
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const { data: incident, error, isError, isLoading, refetch } = useIncident(incidentID);

  useEffect(() => {
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

        <div className="flex-1 overflow-y-auto p-5">
          <div className="space-y-6">
            {/* Description */}
            <div>
              <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
                Description
              </h3>
              <p className="mt-2 text-sm leading-6">
                {incident.description || "No description provided"}
              </p>
            </div>

            {/* Metadata Grid */}
            <dl className="grid gap-4">
              <DrawerStat label="Project" value={projectName} />
              <DrawerStat label="Created" value={formatDate(incident.createdAt)} />
              <DrawerStat label="Updated" value={formatDate(incident.updatedAt)} />
            </dl>
          </div>
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

function getSeverityColor(severity: string): string {
  return INCIDENT_SEVERITY_COLORS[severity as IncidentSeverity] || "var(--muted-foreground)";
}

function getStatusLabel(status: string): string {
  return INCIDENT_STATUS_LABELS[status as IncidentStatus] || status;
}

function getStatusVariant(status: string): "info" | "warning" | "success" {
  return INCIDENT_STATUS_VARIANTS[status as IncidentStatus] || "info";
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
