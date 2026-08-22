"use client";

import { FileText } from "lucide-react";

import { ApplicationLogs } from "@/components/applications/application-logs";
import { EmptyState } from "@/components/common";
import type { IncidentResponse } from "@/types/incident-api";

/**
 * Relevant logs, reusing the Application Logs view (Phase 11), which
 * already aggregates across the application's runtime pods. Only available
 * when the incident has an affected application.
 */
export function IncidentLogsPanel({ incident }: { incident: IncidentResponse }) {
  if (!incident.applicationId) {
    return (
      <EmptyState
        icon={FileText}
        title="Logs not available"
        description="This incident has no affected application to fetch logs from."
      />
    );
  }

  return <ApplicationLogs applicationId={incident.applicationId} />;
}
