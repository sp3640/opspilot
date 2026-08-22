"use client";

import { useMemo } from "react";
import { Layers3, Rocket } from "lucide-react";

import { EmptyState, StatusBadge } from "@/components/common";
import { useApplicationsByProject } from "@/hooks/use-applications";
import { useDeploymentsByProject } from "@/hooks/use-project-deployments";
import { useAlerts } from "@/hooks/use-alerts";
import { useIncidents } from "@/hooks/use-incidents";
import { useProjectTeams } from "@/hooks/use-project-teams";
import { useTeams } from "@/hooks/use-teams";
import { PAGINATION_MAX_PAGE_SIZE } from "@/lib/constants/pagination";
import type { ProjectResponse } from "@/types/project-api";

const TEAM_LOOKUP_QUERY = { page: 1, limit: 100, sort: "name" as const, order: "asc" as const };
const APPLICATIONS_ROLLUP_QUERY = { page: 1, limit: PAGINATION_MAX_PAGE_SIZE, sort: "name" as const, order: "asc" as const };
const RECENT_DEPLOYMENTS_QUERY = { page: 1, limit: 5, sort: "created_at" as const, order: "desc" as const };

const ACTIVE_INCIDENT_STATUSES = ["OPEN", "INVESTIGATING"];
const ACTIVE_ALERT_STATUSES = ["OPEN", "ACKNOWLEDGED", "INVESTIGATING"];

export function ProjectOverview({ project }: { project: ProjectResponse }) {
  const { data: owningTeamsData, isLoading: isOwningTeamsLoading } = useProjectTeams(project.id);
  const { data: teamsLookupData } = useTeams(TEAM_LOOKUP_QUERY);
  const owningTeamNames = useMemo(() => {
    const teamById = new Map((teamsLookupData?.items ?? []).map((team) => [team.id, team.name]));
    return (owningTeamsData?.items ?? []).map((assignment) => teamById.get(assignment.teamId) ?? assignment.teamId);
  }, [owningTeamsData, teamsLookupData]);

  const { data: applicationsData, isLoading: isApplicationsLoading } = useApplicationsByProject(
    project.id,
    APPLICATIONS_ROLLUP_QUERY
  );
  const applications = applicationsData?.items ?? [];
  const environments = useMemo(
    () => Array.from(new Set(applications.map((application) => application.environment).filter(Boolean))),
    [applications]
  );

  const { data: incidentsData, isLoading: isIncidentsLoading } = useIncidents({
    page: 1,
    limit: PAGINATION_MAX_PAGE_SIZE,
    projectId: project.id,
  });
  const activeIncidentCount = (incidentsData?.items ?? []).filter((incident) =>
    ACTIVE_INCIDENT_STATUSES.includes(incident.status)
  ).length;

  const { data: alertsData, isLoading: isAlertsLoading } = useAlerts({
    page: 1,
    limit: PAGINATION_MAX_PAGE_SIZE,
    projectId: project.id,
  });
  const activeAlertCount = (alertsData?.items ?? []).filter((alert) => ACTIVE_ALERT_STATUSES.includes(alert.status)).length;

  const { data: deploymentsData, isLoading: isDeploymentsLoading } = useDeploymentsByProject(
    project.id,
    RECENT_DEPLOYMENTS_QUERY
  );
  const recentDeployments = deploymentsData?.items ?? [];

  return (
    <div className="space-y-6">
      <p className="text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>
        {project.description || "No description provided"}
      </p>

      <dl className="grid grid-cols-2 gap-3">
        <DrawerStat label="Services" value={String(project.services)} />
        <DrawerStat label="Members" value={String(project.members)} />
        <DrawerStat label="Created" value={formatDate(project.createdAt)} />
        <DrawerStat label="Project owner" value={project.owner.name} />
      </dl>

      <div>
        <h3 className="text-sm font-semibold">Project identifier</h3>
        <p className="mt-3 rounded-2xl border p-3 text-sm" style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}>
          {project.slug}
        </p>
      </div>

      <div>
        <h3 className="text-sm font-semibold">Owning team</h3>
        <div className="mt-3 flex flex-wrap gap-2">
          {isOwningTeamsLoading ? (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Loading owning team...</p>
          ) : owningTeamNames.length === 0 ? (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>No team assigned</p>
          ) : (
            owningTeamNames.map((name) => <TeamChip key={name} name={name} />)
          )}
        </div>
      </div>

      <dl className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <DrawerStat
          label="Applications"
          value={isApplicationsLoading ? "…" : String(applicationsData?.total ?? applications.length)}
        />
        <DrawerStat
          label="Active alerts"
          value={isAlertsLoading ? "…" : String(activeAlertCount)}
        />
        <DrawerStat
          label="Active incidents"
          value={isIncidentsLoading ? "…" : String(activeIncidentCount)}
        />
        <DrawerStat
          label="Environments"
          value={isApplicationsLoading ? "…" : String(environments.length)}
        />
      </dl>

      <div>
        <h3 className="text-sm font-semibold">Environments</h3>
        <div className="mt-3 flex flex-wrap gap-2">
          {isApplicationsLoading ? (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Loading environments...</p>
          ) : environments.length === 0 ? (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
              No environments yet — add applications to see their environments here.
            </p>
          ) : (
            environments.map((environment) => (
              <span
                key={environment}
                className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
                style={{ color: "var(--foreground)", backgroundColor: "var(--muted)" }}
              >
                {environment}
              </span>
            ))
          )}
        </div>
      </div>

      <div>
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold">Recent deployments</h3>
        </div>
        <div className="mt-3">
          {isDeploymentsLoading ? (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Loading recent deployments...</p>
          ) : recentDeployments.length === 0 ? (
            <EmptyState
              icon={Rocket}
              title="No deployments yet"
              description="Deployments created for applications in this project will appear here."
            />
          ) : (
            <ul className="space-y-2">
              {recentDeployments.map((deployment) => (
                <li
                  key={deployment.id}
                  className="flex items-center justify-between gap-3 rounded-2xl border p-3"
                  style={{ borderColor: "var(--border)" }}
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">
                      {deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image}
                    </p>
                    <p className="mt-1 truncate text-xs" style={{ color: "var(--muted-foreground)" }}>
                      {deployment.environment || "-"} · {formatDate(deployment.createdAt)}
                    </p>
                  </div>
                  <StatusBadge variant={getDeploymentStatusVariant(deployment.status)}>{deployment.status}</StatusBadge>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {applications.length > 0 ? (
        <div>
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold">Applications</h3>
          </div>
          <ul className="mt-3 space-y-2">
            {applications.slice(0, 5).map((application) => (
              <li
                key={application.id}
                className="flex items-center justify-between gap-3 rounded-2xl border p-3"
                style={{ borderColor: "var(--border)" }}
              >
                <div className="flex min-w-0 items-center gap-2">
                  <Layers3 aria-hidden="true" className="h-4 w-4 shrink-0" style={{ color: "var(--muted-foreground)" }} />
                  <p className="truncate text-sm font-medium">{application.name}</p>
                </div>
                <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>{application.environment || "-"}</span>
              </li>
            ))}
          </ul>
        </div>
      ) : null}
    </div>
  );
}

function TeamChip({ name }: { name: string }) {
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
      style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}
    >
      {name}
    </span>
  );
}

function DrawerStat({ label, value }: { label: string; value: string }) {
  return (
    <div
      className="rounded-2xl border p-4"
      style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
    >
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</dt>
      <dd className="mt-1 font-semibold">{value}</dd>
    </div>
  );
}

function getDeploymentStatusVariant(status: string): "success" | "warning" | "critical" | "archived" | "info" {
  switch (status) {
    case "Succeeded":
      return "success";
    case "Failed":
      return "critical";
    case "Pending":
    case "Queued":
    case "RolledBack":
      return "warning";
    case "Cancelled":
      return "archived";
    default:
      return "info";
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
