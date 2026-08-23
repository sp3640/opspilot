"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Pencil, Trash2 } from "lucide-react";

import { ErrorState, PageHeader } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { Button } from "@/components/ui/button";
import { useApplication } from "@/hooks/use-applications";
import { useIncident } from "@/hooks/use-incidents";
import { useProjects } from "@/hooks/use-projects";
import { useTeam } from "@/hooks/use-teams";
import { PROJECT_LOOKUP_QUERY } from "@/lib/constants";
import { useHasPermission } from "@/store/auth-store";

import { DeleteIncidentDialog } from "./delete-incident-dialog";
import { EditIncidentModal } from "./edit-incident-modal";
import { IncidentAlerts } from "./incident-alerts";
import { IncidentAssignment } from "./incident-assignment";
import { IncidentCombinedTimeline } from "./incident-combined-timeline";
import { IncidentComments } from "./incident-comments";
import { IncidentContextChain } from "./incident-context-chain";
import { IncidentDeploymentsPanel } from "./incident-deployments-panel";
import { IncidentK8sEventsPanel } from "./incident-k8s-events-panel";
import { IncidentLogsPanel } from "./incident-logs-panel";
import { IncidentMetricsPanel } from "./incident-metrics-panel";
import { IncidentSummary } from "./incident-summary";
import { IncidentTimeline } from "./incident-timeline";

/**
 * The full incident investigation console (Phase 17): every piece of
 * evidence connected to an incident laid out as one scrollable page, so
 * an engineer can investigate without hopping between the Alerts, Metrics,
 * Logs, Deployments, and Audit screens individually. "Timeline" is the
 * combined, meaningful story (created/fired/deployed/rolled back/
 * acknowledged/commented/resolved); "Audit" is the incident's raw,
 * unfiltered change log - the two are deliberately kept distinct.
 */
export function IncidentConsole({ incidentID }: { incidentID: number }) {
  const router = useRouter();
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const canManageIncidents = useHasPermission("incident:manage");

  const { data: incident, error, isError, isLoading, refetch } = useIncident(incidentID);
  const { data: projectsData } = useProjects(PROJECT_LOOKUP_QUERY);
  const { data: application } = useApplication(incident?.applicationId ?? null);
  const { data: ownerTeam } = useTeam(incident?.ownerTeamId ?? null);

  if (isLoading) {
    return (
      <div className="mx-auto max-w-[1200px] space-y-6">
        <PageHeader
          title="Loading incident…"
          breadcrumb={[{ label: "Overview", href: "/" }, { label: "Incidents", href: "/incidents" }, { label: "…" }]}
        />
      </div>
    );
  }

  if (isError || !incident) {
    return (
      <div className="mx-auto max-w-[1200px] space-y-6">
        <PageHeader
          title="Incident not found"
          breadcrumb={[{ label: "Overview", href: "/" }, { label: "Incidents", href: "/incidents" }, { label: "Not found" }]}
        />
        <ErrorState
          description={error instanceof Error ? error.message : "Unable to load this incident."}
          onRetry={() => {
            void refetch();
          }}
        />
      </div>
    );
  }

  const projectName = (projectsData?.items ?? []).find((project) => project.id === incident.projectId)?.name ?? "Unknown Project";
  const applicationName = application?.name ?? "Not set";
  const ownerTeamName = ownerTeam?.name ?? "Not set";

  return (
    <div className="mx-auto max-w-[1200px] space-y-6 lg:space-y-8">
      <PageHeader
        title={incident.title}
        description="Incident investigation workspace"
        breadcrumb={[
          { label: "Overview", href: "/" },
          { label: "Incidents", href: "/incidents" },
          { label: incident.title },
        ]}
        actions={
          canManageIncidents ? (
            <div className="flex items-center gap-2">
              <Button type="button" variant="secondary" onClick={() => setEditModalOpen(true)}>
                <Pencil aria-hidden="true" className="h-4 w-4" />
                Edit
              </Button>
              <Button
                type="button"
                variant="ghost"
                onClick={() => setDeleteDialogOpen(true)}
                className="text-[var(--danger)]"
              >
                <Trash2 aria-hidden="true" className="h-4 w-4" />
                Delete
              </Button>
            </div>
          ) : undefined
        }
      />

      <SectionCard title="Incident Summary">
        <IncidentSummary
          incident={incident}
          projectName={projectName}
          applicationName={applicationName}
          ownerTeamName={ownerTeamName}
        />
      </SectionCard>

      <SectionCard title="Assignment" description="Who owns this incident, and whether it has been acknowledged.">
        <IncidentAssignment incident={incident} />
      </SectionCard>

      <SectionCard title="Affected Services" description="Application, deployment, cluster, and namespace connected to this incident.">
        <IncidentContextChain incident={incident} />
      </SectionCard>

      <SectionCard title="Alerts" description="Alerts explicitly attached to this incident.">
        <IncidentAlerts incident={incident} />
      </SectionCard>

      <SectionCard title="Timeline" description="Every meaningful event, merged and ordered chronologically.">
        <IncidentCombinedTimeline incident={incident} />
      </SectionCard>

      <SectionCard title="Metrics" description="Historical metrics for the affected application's cluster.">
        <IncidentMetricsPanel incident={incident} />
      </SectionCard>

      <SectionCard title="Logs" description="Aggregated runtime logs for the affected application.">
        <IncidentLogsPanel incident={incident} />
      </SectionCard>

      <SectionCard title="Deployments" description="Deployment history for the affected application.">
        <IncidentDeploymentsPanel incident={incident} />
      </SectionCard>

      <SectionCard title="Kubernetes Events" description="Live cluster events for the affected application's namespace.">
        <IncidentK8sEventsPanel incident={incident} />
      </SectionCard>

      <SectionCard title="Comments" description="Discussion and updates from the response team.">
        <IncidentComments incidentId={incident.id} />
      </SectionCard>

      <SectionCard title="Audit" description="The incident's complete, unfiltered change log.">
        <IncidentTimeline incident={incident} />
      </SectionCard>

      <EditIncidentModal open={editModalOpen} incident={incident} onClose={() => setEditModalOpen(false)} />
      <DeleteIncidentDialog
        open={deleteDialogOpen}
        incident={incident}
        onClose={() => setDeleteDialogOpen(false)}
        onSuccess={() => router.push("/incidents")}
      />
    </div>
  );
}

