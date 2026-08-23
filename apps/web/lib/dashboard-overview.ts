export type OverallHealthState = "HEALTHY" | "DEGRADED" | "CRITICAL" | "UNKNOWN";

export type DashboardCounts = {
  applications: number | null;
  clusters: number | null;
  activeAlerts: number | null;
  criticalAlerts: number | null;
  activeIncidents: number | null;
  recentDeployments: number | null;
  failedDeployments: number | null;
  unhealthyPods: number | null;
};

export type ProblemItem = {
  id: string;
  title: string;
  description: string;
  href: string;
};

export type ActivityItem = {
  id: string;
  timestamp: string;
  title: string;
  href: string;
};

export type DashboardOverview = {
  overallHealth: OverallHealthState;
  counts: DashboardCounts;
  criticalProblems: ProblemItem[];
  warnings: ProblemItem[];
  recentActivity: ActivityItem[];
};

const ACTIVE_ALERT_STATUSES = new Set(["OPEN", "ACKNOWLEDGED", "INVESTIGATING"]);
const ACTIVE_INCIDENT_STATUSES = new Set(["OPEN", "INVESTIGATING", "MITIGATING"]);
const CRITICAL_INCIDENT_SEVERITIES = new Set(["P0", "P1"]);

type AlertLike = { id: number; title: string; severity: string; status: string; firstSeenAt: string };
type IncidentLike = { id: number; title: string; severity: string; status: string; createdAt: string };
type DeploymentLike = { id: string; image: string; imageTag: string; status: string; createdAt: string };
type ClusterLike = { id: string; name: string; status: string };
type ClusterPodHealthLike = { clusterId: string; clusterName: string; unhealthyCount: number; checked: boolean };

const RECENT_ACTIVITY_DEFAULT_LIMIT = 10;

/**
 * Builds the operations dashboard's overview from already-fetched, real
 * data only. Every input is optional/omittable independently - a source the
 * caller could not load contributes nothing (its count stays null, its
 * problems are simply absent) rather than being guessed at. "Overall
 * health" and every count/problem below are deterministic functions of
 * whatever real data was actually provided.
 */
export function buildDashboardOverview(input: {
  applications?: { total: number };
  clusters?: ClusterLike[];
  alerts?: AlertLike[];
  incidents?: IncidentLike[];
  recentDeployments?: DeploymentLike[];
  unhealthyPodsByCluster?: ClusterPodHealthLike[];
  recentActivityLimit?: number;
}): DashboardOverview {
  const criticalProblems: ProblemItem[] = [];
  const warnings: ProblemItem[] = [];
  const activity: ActivityItem[] = [];

  // Clusters
  const disconnectedClusters = (input.clusters ?? []).filter((cluster) => cluster.status === "INVALID");
  for (const cluster of disconnectedClusters) {
    criticalProblems.push({
      id: `cluster-${cluster.id}`,
      title: `${cluster.name} disconnected`,
      description: "This cluster failed its most recent connectivity validation.",
      href: "/clusters",
    });
  }

  // Alerts
  const activeAlerts = (input.alerts ?? []).filter((alert) => ACTIVE_ALERT_STATUSES.has(alert.status));
  const criticalAlerts = activeAlerts.filter((alert) => alert.severity === "CRITICAL");
  for (const alert of criticalAlerts) {
    criticalProblems.push({
      id: `alert-${alert.id}`,
      title: alert.title,
      description: "Critical alert",
      href: "/alerts",
    });
  }
  for (const alert of activeAlerts.filter((alert) => alert.severity !== "CRITICAL")) {
    warnings.push({
      id: `alert-${alert.id}`,
      title: alert.title,
      description: `${alert.severity} alert`,
      href: "/alerts",
    });
  }
  for (const alert of input.alerts ?? []) {
    activity.push({ id: `alert-${alert.id}`, timestamp: alert.firstSeenAt, title: `Alert fired: ${alert.title}`, href: "/alerts" });
  }

  // Incidents
  const activeIncidents = (input.incidents ?? []).filter((incident) => ACTIVE_INCIDENT_STATUSES.has(incident.status));
  for (const incident of activeIncidents) {
    const item: ProblemItem = {
      id: `incident-${incident.id}`,
      title: incident.title,
      description: `Incident — ${incident.severity}`,
      href: `/incidents/${incident.id}`,
    };
    if (CRITICAL_INCIDENT_SEVERITIES.has(incident.severity)) {
      criticalProblems.push(item);
    } else {
      warnings.push(item);
    }
  }
  for (const incident of input.incidents ?? []) {
    activity.push({
      id: `incident-${incident.id}`,
      timestamp: incident.createdAt,
      title: `Incident opened: ${incident.title}`,
      href: `/incidents/${incident.id}`,
    });
  }

  // Deployments
  const failedDeployments = (input.recentDeployments ?? []).filter((deployment) => deployment.status === "Failed");
  for (const deployment of failedDeployments) {
    criticalProblems.push({
      id: `deployment-${deployment.id}`,
      title: `Deployment ${deploymentLabel(deployment)} failed`,
      description: "The most recent execution of this deployment failed.",
      href: "/projects",
    });
  }
  for (const deployment of input.recentDeployments ?? []) {
    activity.push({
      id: `deployment-${deployment.id}`,
      timestamp: deployment.createdAt,
      title: `Deployment ${deploymentLabel(deployment)} — ${deployment.status}`,
      href: "/projects",
    });
  }

  // Pods
  let unhealthyPodsTotal: number | null = null;
  if (input.unhealthyPodsByCluster && input.unhealthyPodsByCluster.some((entry) => entry.checked)) {
    unhealthyPodsTotal = 0;
    for (const entry of input.unhealthyPodsByCluster) {
      if (!entry.checked) continue;
      unhealthyPodsTotal += entry.unhealthyCount;
      if (entry.unhealthyCount > 0) {
        warnings.push({
          id: `pods-${entry.clusterId}`,
          title: `${entry.unhealthyCount} unhealthy pod(s) in ${entry.clusterName}`,
          description: "Pods that are not Running/Succeeded, or have restarted.",
          href: "/clusters",
        });
      }
    }
  }

  activity.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
  const recentActivity = activity.slice(0, input.recentActivityLimit ?? RECENT_ACTIVITY_DEFAULT_LIMIT);

  const counts: DashboardCounts = {
    applications: input.applications ? input.applications.total : null,
    clusters: input.clusters ? input.clusters.length : null,
    activeAlerts: input.alerts ? activeAlerts.length : null,
    criticalAlerts: input.alerts ? criticalAlerts.length : null,
    activeIncidents: input.incidents ? activeIncidents.length : null,
    recentDeployments: input.recentDeployments ? input.recentDeployments.length : null,
    failedDeployments: input.recentDeployments ? failedDeployments.length : null,
    unhealthyPods: unhealthyPodsTotal,
  };

  return {
    overallHealth: computeOverallHealth(input, criticalProblems, warnings),
    counts,
    criticalProblems,
    warnings,
    recentActivity,
  };
}

function computeOverallHealth(
  input: {
    applications?: { total: number };
    clusters?: ClusterLike[];
    alerts?: AlertLike[];
    incidents?: IncidentLike[];
    recentDeployments?: DeploymentLike[];
    unhealthyPodsByCluster?: ClusterPodHealthLike[];
  },
  criticalProblems: ProblemItem[],
  warnings: ProblemItem[]
): OverallHealthState {
  const anyRealDataLoaded = Boolean(
    input.clusters ||
      input.alerts ||
      input.incidents ||
      input.applications ||
      input.recentDeployments ||
      input.unhealthyPodsByCluster?.some((entry) => entry.checked)
  );
  if (!anyRealDataLoaded) return "UNKNOWN";

  if (criticalProblems.length > 0) return "CRITICAL";
  if (warnings.length > 0) return "DEGRADED";
  return "HEALTHY";
}

function deploymentLabel(deployment: DeploymentLike): string {
  return deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image;
}
