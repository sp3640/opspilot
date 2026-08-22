export const INCIDENT_SEVERITY_VALUES = ["P0", "P1", "P2", "P3", "P4"] as const;
export const INCIDENT_STATUS_VALUES = ["OPEN", "INVESTIGATING", "MITIGATING", "RESOLVED"] as const;

export type IncidentSeverity = (typeof INCIDENT_SEVERITY_VALUES)[number];
export type IncidentStatus = (typeof INCIDENT_STATUS_VALUES)[number];

export const INCIDENT_DEFAULT_SEVERITY: IncidentSeverity = "P0";
export const INCIDENT_DEFAULT_STATUS: IncidentStatus = "OPEN";

export const INCIDENT_SEVERITY_OPTIONS: ReadonlyArray<{
  value: IncidentSeverity;
  label: string;
}> = [
  { value: "P0", label: "P0 - Critical" },
  { value: "P1", label: "P1 - High" },
  { value: "P2", label: "P2 - Medium" },
  { value: "P3", label: "P3 - Low" },
  { value: "P4", label: "P4 - Minimal" },
];

export const INCIDENT_STATUS_OPTIONS: ReadonlyArray<{
  value: IncidentStatus;
  label: string;
}> = [
  { value: "OPEN", label: "Open" },
  { value: "INVESTIGATING", label: "Investigating" },
  { value: "MITIGATING", label: "Mitigating" },
  { value: "RESOLVED", label: "Resolved" },
];

export const INCIDENT_SEVERITY_COLORS: Record<IncidentSeverity, string> = {
  P0: "#dc2626",
  P1: "#ea580c",
  P2: "#f59e0b",
  P3: "#eab308",
  P4: "#84cc16",
};

export const INCIDENT_STATUS_LABELS: Record<IncidentStatus, string> = {
  OPEN: "Open",
  INVESTIGATING: "Investigating",
  MITIGATING: "Mitigating",
  RESOLVED: "Resolved",
};

export const INCIDENT_STATUS_VARIANTS: Record<
  IncidentStatus,
  "info" | "warning" | "success"
> = {
  OPEN: "info",
  INVESTIGATING: "warning",
  MITIGATING: "warning",
  RESOLVED: "success",
};
