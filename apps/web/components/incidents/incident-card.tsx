"use client";

import { AlertCircle, Clock } from "lucide-react";

import { StatusBadge } from "@/components/common";
import type { IncidentResponse } from "@/types/incident-api";

type IncidentCardProps = {
  incident: IncidentResponse;
  projectName: string;
  onSelect?: (id: number) => void;
};

export function IncidentCard({ incident, projectName, onSelect }: IncidentCardProps) {
  const severityColor = getSeverityColor(incident.severity);
  const statusLabel = getStatusLabel(incident.status);
  const statusVariant = getStatusVariant(incident.status);

  return (
    <article
      className="cursor-pointer rounded-3xl border p-5 shadow-[var(--shadow-sm)] transition-all duration-300 hover:shadow-[var(--shadow-md)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
      onClick={() => onSelect?.(incident.id)}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelect?.(incident.id);
        }
      }}
      role="button"
      tabIndex={0}
      aria-label={`Open details for incident: ${incident.title}`}
    >
      <div className="flex items-start justify-between gap-3">
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
            <h2 className="truncate font-semibold tracking-tight">{incident.title}</h2>
            <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
              {projectName}
            </p>
          </div>
        </div>
      </div>

      <p className="mt-5 line-clamp-2 min-h-10 text-sm leading-5" style={{ color: "var(--muted-foreground)" }}>
        {incident.description || "No description provided"}
      </p>

      <div className="mt-5 flex flex-wrap gap-2">
        <StatusBadge variant={statusVariant}>{statusLabel}</StatusBadge>
        <div
          className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
          style={{ backgroundColor: `color-mix(in srgb, ${severityColor} 12%, transparent)`, color: severityColor }}
        >
          {incident.severity}
        </div>
      </div>

      <div className="mt-4 flex items-center justify-between gap-3 border-t pt-4 text-xs" style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}>
        <div className="flex items-center gap-1">
          <Clock aria-hidden="true" className="h-3.5 w-3.5" />
          Updated {formatDate(incident.updatedAt)}
        </div>
      </div>
    </article>
  );
}

function getSeverityColor(severity: string): string {
  const colors: Record<string, string> = {
    P0: "#dc2626",
    P1: "#ea580c",
    P2: "#f59e0b",
    P3: "#eab308",
    P4: "#84cc16",
  };
  return colors[severity] || "var(--muted-foreground)";
}

function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    OPEN: "Open",
    INVESTIGATING: "Investigating",
    RESOLVED: "Resolved",
  };
  return labels[status] || status;
}

function getStatusVariant(status: string): "info" | "warning" | "success" {
  const variants: Record<string, "info" | "warning" | "success"> = {
    OPEN: "info",
    INVESTIGATING: "warning",
    RESOLVED: "success",
  };
  return variants[status] || "info";
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
