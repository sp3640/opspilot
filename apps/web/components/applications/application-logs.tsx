"use client";

import { useMemo, useState } from "react";
import axios from "axios";
import { Database, Rocket } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { PodLogs } from "@/components/pods/pod-logs";
import { useLatestDeployment } from "@/hooks/use-deployments";
import { usePodsByApplication } from "@/hooks/use-pods";
import { classifyPodHealth } from "@/lib/pod-health";
import type { PodResponse } from "@/types/pod-api";

const ALL_PODS_KEY = "__all__";
const MAX_AGGREGATED_PODS = 5;

/**
 * Application -> Pod -> Container -> Logs.
 *
 * "Aggregated" here means collating each mapped runtime pod's own Live
 * Kubernetes Pod Logs panel in one place (each pod keeps its own container
 * selection, search, time range, and live tail) - not a true merged/sorted
 * multi-pod log stream. There is no centralized log store behind this: see
 * the "Centralized Historical Logs" section below, which is honestly a
 * placeholder until a real log platform (e.g. Loki) is introduced.
 */
export function ApplicationLogs({ applicationId }: { applicationId: string }) {
  const {
    data: deployment,
    error: deploymentError,
    isError: isDeploymentError,
    isLoading: isDeploymentLoading,
    refetch: refetchDeployment,
  } = useLatestDeployment(applicationId);

  const notDeployedYet = axios.isAxiosError(deploymentError) && deploymentError.response?.status === 404;

  const podsParams = useMemo(() => ({ namespace: deployment?.namespace ?? "" }), [deployment?.namespace]);
  const {
    data: podsData,
    error: podsError,
    isError: isPodsError,
    isLoading: isPodsLoading,
    refetch: refetchPods,
  } = usePodsByApplication(applicationId, podsParams, Boolean(deployment?.namespace));
  const pods = useMemo(() => podsData?.items ?? [], [podsData]);

  const [selectedPodKey, setSelectedPodKey] = useState<string>(ALL_PODS_KEY);
  const selectedPod = pods.find((pod) => `${pod.namespace}/${pod.name}` === selectedPodKey) ?? null;
  const aggregatedPods = pods.slice(0, MAX_AGGREGATED_PODS);

  return (
    <div className="space-y-8">
      <section>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
            Live Kubernetes Pod Logs
          </h3>
          {pods.length > 0 ? (
            <select
              aria-label="Pod"
              value={selectedPodKey}
              onChange={(event) => setSelectedPodKey(event.target.value)}
              className="h-10 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]"
              style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
            >
              <option value={ALL_PODS_KEY}>All running pods ({pods.length})</option>
              {pods.map((pod) => (
                <option key={`${pod.namespace}/${pod.name}`} value={`${pod.namespace}/${pod.name}`}>
                  {pod.name}
                </option>
              ))}
            </select>
          ) : null}
        </div>
        <p className="mt-2 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Fetched directly from the Kubernetes API for this application&apos;s currently running pods - not a
          centralized log store.
        </p>

        <div className="mt-4">
          {isDeploymentLoading ? (
            <ListSkeleton rows={3} />
          ) : notDeployedYet ? (
            <EmptyState
              icon={Rocket}
              title="Not deployed yet"
              description="Deploy this application to see logs from its runtime pods here."
            />
          ) : isDeploymentError ? (
            <ErrorState
              description={deploymentError instanceof Error ? deploymentError.message : "Unable to load runtime information."}
              onRetry={() => { void refetchDeployment(); }}
            />
          ) : isPodsLoading ? (
            <ListSkeleton rows={3} />
          ) : isPodsError ? (
            <ErrorState
              description={podsError instanceof Error ? podsError.message : "Unable to load pods. Please try again."}
              onRetry={() => { void refetchPods(); }}
            />
          ) : pods.length === 0 ? (
            <EmptyState
              icon={Rocket}
              title="No running pods"
              description="This application has no running pods to fetch logs from right now."
            />
          ) : selectedPod ? (
            <PodLogPanel applicationId={applicationId} pod={selectedPod} />
          ) : (
            <div className="space-y-6">
              {aggregatedPods.map((pod) => (
                <PodLogPanel key={`${pod.namespace}/${pod.name}`} applicationId={applicationId} pod={pod} />
              ))}
              {pods.length > aggregatedPods.length ? (
                <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                  Showing logs for {aggregatedPods.length} of {pods.length} pods. Select a specific pod above to view
                  its logs.
                </p>
              ) : null}
            </div>
          )}
        </div>
      </section>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Centralized Historical Logs
        </h3>
        <div className="mt-3">
          <EmptyState
            icon={Database}
            title="Not connected yet"
            description="OpsPilot does not aggregate logs into a centralized store (e.g. Loki or Elasticsearch) yet. The Live Kubernetes Pod Logs above are fetched directly from the cluster and are lost once a pod is deleted."
          />
        </div>
      </section>
    </div>
  );
}

function PodLogPanel({ applicationId, pod }: { applicationId: string; pod: PodResponse }) {
  const health = classifyPodHealth(pod);

  return (
    <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="min-w-0 truncate font-semibold">{pod.name}</p>
        <StatusBadge variant={health.variant}>{health.state}</StatusBadge>
      </div>
      <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>{pod.namespace}</p>
      <div className="mt-4">
        <PodLogs applicationId={applicationId} pod={pod} />
      </div>
    </div>
  );
}
