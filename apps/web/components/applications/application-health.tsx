"use client";

import { useMemo } from "react";
import axios from "axios";
import { Boxes, Rocket } from "lucide-react";

import { ClusterStatusBadge } from "@/components/clusters/cluster-status";
import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useCluster } from "@/hooks/use-clusters";
import { useLatestDeployment } from "@/hooks/use-deployments";
import { usePodsByApplication } from "@/hooks/use-pods";
import { useRuntimeDeploymentsByApplication } from "@/hooks/use-runtime-deployments";
import { matchRuntimeDeployment } from "@/lib/deployment-replica-state";
import type { PodResponse } from "@/types/pod-api";

import { ApplicationHealthScore } from "./application-health-score";

/**
 * The primary "is this application healthy" summary. The unified health
 * score (Phase 21) - real, deterministic, and documented per-factor - has
 * replaced the earlier ad-hoc client-side approximation that lived here
 * (manually re-deriving deployment/pod/alert/incident state, plus an
 * always-empty Prometheus-metrics placeholder). See
 * components/applications/application-health-score.tsx and the backend's
 * internal/health package for how it's computed.
 *
 * Below the score, "Runtime" connects the chain this data comes from:
 * Application -> Deployment -> Cluster -> Namespace -> Pods, plus real
 * replica state (ready/available) matched from the live k8s Deployment
 * object via matchRuntimeDeployment - "Not available" when no confident
 * match exists, never a guess.
 */
export function ApplicationHealth({
  applicationId,
  status,
  environment,
}: {
  applicationId: string;
  status: string;
  environment: string;
}) {
  const {
    data: deployment,
    error: deploymentError,
    isError: isDeploymentError,
    isLoading: isDeploymentLoading,
    refetch: refetchDeployment,
  } = useLatestDeployment(applicationId);

  const notDeployedYet = axios.isAxiosError(deploymentError) && deploymentError.response?.status === 404;

  const { data: cluster, isLoading: isClusterLoading } = useCluster(deployment?.targetClusterId ?? null);

  const podsParams = useMemo(() => ({ namespace: deployment?.namespace ?? "" }), [deployment?.namespace]);
  const {
    data: podsData,
    isLoading: isPodsLoading,
  } = usePodsByApplication(applicationId, podsParams, Boolean(deployment?.namespace));
  const pods = useMemo(() => podsData?.items ?? [], [podsData]);
  const podHealth = useMemo(() => summarizePodHealth(pods), [pods]);

  const { data: runtimeData } = useRuntimeDeploymentsByApplication(
    applicationId,
    { namespace: deployment?.namespace ?? "" },
    Boolean(deployment?.namespace)
  );
  const runtimeMatch = deployment ? matchRuntimeDeployment(deployment, runtimeData?.items ?? []) : null;

  return (
    <div className="space-y-6">
      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Overview
        </h3>
        <div className="mt-3 flex flex-wrap gap-2">
          <StatusBadge variant={getApplicationStatusVariant(status)}>{status}</StatusBadge>
          {environment ? (
            <span
              className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
              style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}
            >
              {environment}
            </span>
          ) : (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
              No environment set for this application.
            </p>
          )}
        </div>
      </section>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Health Score
        </h3>
        <div className="mt-3">
          <ApplicationHealthScore applicationId={applicationId} />
        </div>
      </section>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Runtime
        </h3>
        <div className="mt-3">
          {isDeploymentLoading ? (
            <ListSkeleton rows={2} />
          ) : notDeployedYet ? (
            <EmptyState
              icon={Rocket}
              title="Not deployed yet"
              description="Deploy this application to see its runtime cluster, namespace, and pods here."
            />
          ) : isDeploymentError ? (
            <ErrorState
              description={
                deploymentError instanceof Error ? deploymentError.message : "Unable to load runtime information."
              }
              onRetry={() => {
                void refetchDeployment();
              }}
            />
          ) : deployment ? (
            <dl className="grid grid-cols-2 gap-3">
              <HealthStat label="Cluster">
                {isClusterLoading ? (
                  <Loading />
                ) : cluster ? (
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-semibold">{cluster.name}</span>
                    <ClusterStatusBadge status={cluster.status} />
                  </div>
                ) : (
                  <span className="break-all font-semibold">{deployment.targetClusterId}</span>
                )}
              </HealthStat>
              <HealthStat label="Namespace">
                <span className="break-all font-semibold">{deployment.namespace}</span>
              </HealthStat>
              <HealthStat label="Deployment">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="break-all font-semibold">
                    {deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image}
                  </span>
                  <StatusBadge variant={getDeploymentStatusVariant(deployment.status)}>{deployment.status}</StatusBadge>
                </div>
              </HealthStat>
              <HealthStat label="Pods">
                {isPodsLoading ? (
                  <Loading />
                ) : pods.length === 0 ? (
                  <NotAvailable label="No pods found" />
                ) : (
                  <div className="flex flex-wrap items-center gap-2">
                    <Boxes aria-hidden="true" className="h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
                    <span className="font-semibold">{pods.length} pod{pods.length === 1 ? "" : "s"}</span>
                    <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>{podHealth.breakdown}</span>
                  </div>
                )}
              </HealthStat>
              <HealthStat label="Replica state">
                {runtimeMatch ? (
                  <span className="font-semibold">
                    {runtimeMatch.readyReplicas}/{runtimeMatch.replicas} ready, {runtimeMatch.availableReplicas} available
                  </span>
                ) : (
                  <NotAvailable label="Not available" />
                )}
              </HealthStat>
            </dl>
          ) : null}
        </div>
      </section>
    </div>
  );
}

function HealthStat({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div
      className="rounded-2xl border p-4"
      style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
    >
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</dt>
      <dd className="mt-1">{children}</dd>
    </div>
  );
}

function Loading() {
  return <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>Loading...</span>;
}

function NotAvailable({ label }: { label: string }) {
  return <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>{label}</span>;
}

function summarizePodHealth(pods: PodResponse[]): { breakdown: string } {
  let healthy = 0;
  let degraded = 0;
  let unhealthy = 0;

  for (const pod of pods) {
    if (pod.phase === "Running" || pod.phase === "Succeeded") {
      healthy += pod.restartCount > 0 ? 0 : 1;
      if (pod.restartCount > 0) degraded += 1;
    } else if (pod.phase === "Pending") {
      degraded += 1;
    } else {
      unhealthy += 1;
    }
  }

  return { breakdown: `${healthy} healthy, ${degraded} degraded, ${unhealthy} unhealthy` };
}

function getApplicationStatusVariant(status: string): "success" | "info" | "archived" {
  switch (status) {
    case "Ready":
      return "success";
    case "Archived":
      return "archived";
    default:
      return "info";
  }
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
