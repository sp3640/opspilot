"use client";

import axios from "axios";
import { Gauge } from "lucide-react";

import { CLUSTER_HISTORICAL_CHART_SPECS } from "@/components/clusters/cluster-metrics";
import { EmptyState, ListSkeleton } from "@/components/common";
import { ScopedMetricsDashboard } from "@/components/metrics/scoped-metrics-dashboard";
import { useLatestDeployment } from "@/hooks/use-deployments";
import type { IncidentResponse } from "@/types/incident-api";

/**
 * Relevant metrics for the incident's affected application's cluster,
 * reusing the same historical dashboard the Cluster Metrics tab uses
 * (Phase 13). Only shown when the incident has an affected application with
 * a real deployment on record - otherwise there is no cluster to scope a
 * metrics view to.
 */
export function IncidentMetricsPanel({ incident }: { incident: IncidentResponse }) {
  const hasApplication = Boolean(incident.applicationId);
  const { data: deployment, error, isLoading } = useLatestDeployment(incident.applicationId ?? null, hasApplication);
  const notDeployed = axios.isAxiosError(error) && error.response?.status === 404;

  if (!hasApplication) {
    return (
      <EmptyState
        icon={Gauge}
        title="No metrics available"
        description="This incident has no affected application to scope a metrics view to."
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={3} />;
  }

  if (notDeployed || !deployment?.targetClusterId) {
    return (
      <EmptyState
        icon={Gauge}
        title="No metrics available"
        description="The affected application has no deployment on record yet."
      />
    );
  }

  return (
    <ScopedMetricsDashboard
      projectId={incident.projectId}
      clusterId={deployment.targetClusterId}
      specs={CLUSTER_HISTORICAL_CHART_SPECS}
    />
  );
}
