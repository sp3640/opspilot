"use client";

import { useMemo } from "react";

import { CLUSTER_HISTORICAL_CHART_SPECS } from "@/components/clusters/cluster-metrics";
import { EmptyState } from "@/components/common";
import { ScopedMetricsDashboard } from "@/components/metrics/scoped-metrics-dashboard";
import { parseAlertMetadata } from "@/lib/alert-correlation";
import type { AlertResponse } from "@/types/alert-api";
import { Gauge } from "lucide-react";

/**
 * Relevant metrics for the alert's cluster, reusing the same historical
 * dashboard the Cluster Metrics tab uses (Phase 13). Only shown when the
 * alert carries a real clusterId - there is no metrics dimension to show
 * otherwise (e.g. a manually-created alert with no condition metadata).
 */
export function AlertMetricsPanel({ alert }: { alert: AlertResponse }) {
  const meta = useMemo(() => parseAlertMetadata(alert.metadata), [alert.metadata]);

  if (!meta.clusterId) {
    return (
      <EmptyState
        icon={Gauge}
        title="No metrics available"
        description="This alert has no cluster identifier to scope a metrics view to."
      />
    );
  }

  return <ScopedMetricsDashboard projectId={alert.projectId} clusterId={meta.clusterId} specs={CLUSTER_HISTORICAL_CHART_SPECS} />;
}
