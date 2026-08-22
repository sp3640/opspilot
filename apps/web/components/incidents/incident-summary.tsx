"use client";

import { StatusBadge } from "@/components/common";
import {
  INCIDENT_SEVERITY_COLORS,
  INCIDENT_STATUS_LABELS,
  INCIDENT_STATUS_VARIANTS,
  type IncidentSeverity,
  type IncidentStatus,
} from "@/lib/constants";
import type { IncidentResponse } from "@/types/incident-api";

/**
 * The incident's own facts - severity/status/description plus the
 * project/application/owner-team/created/resolved stats - shared between
 * the details drawer's Overview tab and the full investigation console's
 * Incident Summary section so the two views can never drift apart.
 */
export function IncidentSummary({
  incident,
  projectName,
  applicationName,
  ownerTeamName,
}: {
  incident: IncidentResponse;
  projectName: string;
  applicationName: string;
  ownerTeamName: string;
}) {
  const severityColor = getSeverityColor(incident.severity);
  const statusLabel = getStatusLabel(incident.status);
  const statusVariant = getStatusVariant(incident.status);

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap gap-2">
        <StatusBadge variant={statusVariant}>{statusLabel}</StatusBadge>
        <div
          className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
          style={{ backgroundColor: `color-mix(in srgb, ${severityColor} 12%, transparent)`, color: severityColor }}
        >
          {incident.severity}
        </div>
      </div>

      <div>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Description
        </h3>
        <p className="mt-2 text-sm leading-6">{incident.description || "No description provided"}</p>
      </div>

      <dl className="grid gap-4 sm:grid-cols-2">
        <DrawerStat label="Project" value={projectName} />
        <DrawerStat label="Affected application" value={applicationName} />
        <DrawerStat label="Owner team" value={ownerTeamName} />
        <DrawerStat label="Created" value={formatDate(incident.createdAt)} />
        <DrawerStat label="Updated" value={formatDate(incident.updatedAt)} />
        <DrawerStat label="Resolved" value={incident.resolvedAt ? formatDate(incident.resolvedAt) : "Not resolved"} />
      </dl>
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
