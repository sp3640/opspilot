"use client";

import { useMemo } from "react";
import { useQueries } from "@tanstack/react-query";

import { ErrorState, PageHeader } from "@/components/common";
import { useAlerts } from "@/hooks/use-alerts";
import { useApplications } from "@/hooks/use-applications";
import { useClusters } from "@/hooks/use-clusters";
import { useDeployments } from "@/hooks/use-deployments";
import { useIncidents } from "@/hooks/use-incidents";
import { buildDashboardOverview } from "@/lib/dashboard-overview";
import { podService } from "@/services/pod-service";
import type { PodResponse } from "@/types/pod-api";

import { DashboardHealthBanner } from "./dashboard-health-banner";
import { DashboardProblems } from "./dashboard-problems";
import { DashboardRecentActivity } from "./dashboard-recent-activity";
import { DashboardStatGrid } from "./dashboard-stat-grid";

const APPLICATIONS_COUNT_QUERY = { page: 1, limit: 1 };
const ORG_LOOKUP_QUERY = { page: 1, limit: 100 };
const RECENT_DEPLOYMENTS_QUERY = { page: 1, limit: 20, sort: "created_at" as const, order: "desc" as const };

// Checking pod health requires one live-cluster call per cluster (there is
// no org-wide/bulk pod endpoint) - capped so a large organization's
// dashboard never fans out to dozens of clusters on every load.
const MAX_CLUSTERS_CHECKED_FOR_PODS = 6;

/**
 * The operations dashboard (Phase 22): "Is my organization healthy right
 * now?" answered from real, already-available data only. Every count and
 * problem is deterministically derived by lib/dashboard-overview.ts from
 * whatever actually loaded - a source that fails or was never attempted
 * contributes nothing, never a fabricated number.
 */
export function DashboardWorkspace() {
  const applicationsQuery = useApplications(APPLICATIONS_COUNT_QUERY);
  const clustersQuery = useClusters(ORG_LOOKUP_QUERY);
  const alertsQuery = useAlerts(ORG_LOOKUP_QUERY);
  const incidentsQuery = useIncidents(ORG_LOOKUP_QUERY);
  const deploymentsQuery = useDeployments(RECENT_DEPLOYMENTS_QUERY);

  const clusters = useMemo(() => clustersQuery.data?.items ?? [], [clustersQuery.data]);
  const clustersToCheckForPods = useMemo(() => clusters.slice(0, MAX_CLUSTERS_CHECKED_FOR_PODS), [clusters]);

  const podQueries = useQueries({
    queries: clustersToCheckForPods.map((cluster) => ({
      queryKey: ["pods", "cluster-list", cluster.id, {}],
      queryFn: () => podService.listPodsByCluster(cluster.id, {}),
      enabled: Boolean(cluster.id),
    })),
  });

  const unhealthyPodsByCluster = useMemo(
    () =>
      clustersToCheckForPods.map((cluster, index) => {
        const query = podQueries[index];
        const checked = Boolean(query?.isSuccess);
        const unhealthyCount = checked ? (query?.data?.items ?? []).filter(isPodUnhealthy).length : 0;
        return { clusterId: cluster.id, clusterName: cluster.name, unhealthyCount, checked };
      }),
    [clustersToCheckForPods, podQueries]
  );

  const isCoreError = applicationsQuery.isError || clustersQuery.isError || alertsQuery.isError || incidentsQuery.isError;
  const isCoreLoading =
    applicationsQuery.isLoading || clustersQuery.isLoading || alertsQuery.isLoading || incidentsQuery.isLoading;

  const overview = useMemo(
    () =>
      buildDashboardOverview({
        applications: applicationsQuery.data ? { total: applicationsQuery.data.total } : undefined,
        clusters: clustersQuery.data?.items,
        alerts: alertsQuery.data?.items,
        incidents: incidentsQuery.data?.items,
        recentDeployments: deploymentsQuery.data?.items,
        unhealthyPodsByCluster: clustersToCheckForPods.length > 0 ? unhealthyPodsByCluster : undefined,
      }),
    [applicationsQuery.data, clustersQuery.data, alertsQuery.data, incidentsQuery.data, deploymentsQuery.data, clustersToCheckForPods, unhealthyPodsByCluster]
  );

  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <PageHeader
        title="Operations Dashboard"
        description="Is my organization healthy right now?"
        breadcrumb={[{ label: "Overview" }]}
      />

      {isCoreError ? (
        <ErrorState
          description="Unable to load one or more parts of the operations dashboard."
          onRetry={() => {
            void applicationsQuery.refetch();
            void clustersQuery.refetch();
            void alertsQuery.refetch();
            void incidentsQuery.refetch();
          }}
        />
      ) : isCoreLoading ? (
        <DashboardSkeleton />
      ) : (
        <>
          <DashboardHealthBanner state={overview.overallHealth} />
          <DashboardProblems criticalProblems={overview.criticalProblems} warnings={overview.warnings} />
          <DashboardStatGrid counts={overview.counts} />
          <DashboardRecentActivity items={overview.recentActivity} />
        </>
      )}
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div className="h-24 animate-pulse rounded-3xl" style={{ backgroundColor: "var(--muted)" }} />
      <div className="grid gap-4 lg:grid-cols-2">
        <div className="h-40 animate-pulse rounded-3xl" style={{ backgroundColor: "var(--muted)" }} />
        <div className="h-40 animate-pulse rounded-3xl" style={{ backgroundColor: "var(--muted)" }} />
      </div>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {Array.from({ length: 8 }, (_, index) => (
          <div key={index} className="h-24 animate-pulse rounded-2xl" style={{ backgroundColor: "var(--muted)" }} />
        ))}
      </div>
    </div>
  );
}

function isPodUnhealthy(pod: PodResponse): boolean {
  if (pod.phase === "Failed" || pod.phase === "Unknown") return true;
  if (pod.restartCount > 0) return true;
  if (pod.phase === "Running" && !pod.ready) return true;
  return false;
}
