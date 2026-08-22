"use client";

import { useState } from "react";
import axios from "axios";
import { Boxes } from "lucide-react";

import { ClusterStatusBadge } from "@/components/clusters/cluster-status";
import { StatusBadge } from "@/components/common";
import { PodDetailsDrawer } from "@/components/pods/pod-details-drawer";
import { useApplication } from "@/hooks/use-applications";
import { useCluster } from "@/hooks/use-clusters";
import { useLatestDeployment } from "@/hooks/use-deployments";
import { usePodsByApplication } from "@/hooks/use-pods";
import { parseResourceIdentifier } from "@/lib/alert-correlation";

/**
 * Affected application -> Affected deployment -> Affected cluster ->
 * Affected namespace -> Affected pod/resource - shared between Alert
 * (Phase 15) and Incident (Phase 16) detail views, since both connect to the
 * same underlying Application -> Deployment -> Cluster -> Namespace -> Pods
 * chain. Every link is only ever shown when a real identifier supports it.
 */
export function AffectedResourceChain({
  applicationId,
  clusterIdFallback,
  clusterNameFallback,
  resourceType,
  resourceId,
}: {
  applicationId: string | null;
  clusterIdFallback?: string | null;
  clusterNameFallback?: string | null;
  resourceType?: string;
  resourceId?: string;
}) {
  const [podDrawerOpen, setPodDrawerOpen] = useState(false);
  const hasApplication = Boolean(applicationId);

  const { data: application, isLoading: isApplicationLoading } = useApplication(applicationId);

  const {
    data: deployment,
    error: deploymentError,
    isLoading: isDeploymentLoading,
  } = useLatestDeployment(applicationId, hasApplication);
  const notDeployed = axios.isAxiosError(deploymentError) && deploymentError.response?.status === 404;

  const clusterId = deployment?.targetClusterId ?? clusterIdFallback ?? null;
  const { data: cluster, isLoading: isClusterLoading } = useCluster(clusterId ?? null);

  const resource = resourceType && resourceId ? parseResourceIdentifier(resourceType, resourceId) : null;
  const namespace = deployment?.namespace ?? resource?.namespace ?? null;

  const podsEnabled = resourceType === "POD" && hasApplication && Boolean(namespace);
  const { data: podsData, isLoading: isPodsLoading } = usePodsByApplication(
    applicationId,
    { namespace: namespace ?? "" },
    podsEnabled
  );
  const matchedPod = (podsData?.items ?? []).find((pod) => pod.name === resource?.name) ?? null;

  return (
    <div className="space-y-3">
      <ChainLink label="Affected application">
        {!hasApplication ? (
          <Unavailable reason="This isn't scoped to a specific application." />
        ) : isApplicationLoading ? (
          <Loading />
        ) : application ? (
          <span className="font-semibold">{application.name}</span>
        ) : (
          <Unavailable reason="The referenced application could not be found." />
        )}
      </ChainLink>

      <ChainLink label="Affected deployment">
        {!hasApplication ? (
          <Unavailable reason="No application is linked." />
        ) : isDeploymentLoading ? (
          <Loading />
        ) : notDeployed ? (
          <Unavailable reason="This application has no deployment on record." />
        ) : deployment ? (
          <div className="flex flex-wrap items-center gap-2">
            <span className="break-all font-semibold">{deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image}</span>
            <StatusBadge variant={deploymentStatusVariant(deployment.status)}>{deployment.status}</StatusBadge>
          </div>
        ) : (
          <Unavailable reason="Unable to determine the deployment for this application." />
        )}
      </ChainLink>

      <ChainLink label="Affected cluster">
        {!clusterId ? (
          <Unavailable reason="No cluster identifier is available." />
        ) : isClusterLoading ? (
          <Loading />
        ) : cluster ? (
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-semibold">{cluster.name}</span>
            <ClusterStatusBadge status={cluster.status} />
          </div>
        ) : (
          <span className="break-all font-semibold">{clusterNameFallback ?? clusterId}</span>
        )}
      </ChainLink>

      <ChainLink label="Affected namespace">
        {namespace ? <span className="break-all font-semibold">{namespace}</span> : <Unavailable reason="No namespace could be determined." />}
      </ChainLink>

      {resourceType ? (
        <ChainLink label={`Affected ${resourceType.toLowerCase()}`}>
          {resourceType === "POD" ? (
            !podsEnabled ? (
              <Unavailable reason="Pod lookup requires a known application and namespace." />
            ) : isPodsLoading ? (
              <Loading />
            ) : matchedPod ? (
              <button
                type="button"
                onClick={() => setPodDrawerOpen(true)}
                className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium underline-offset-2 hover:underline"
                style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}
              >
                <Boxes aria-hidden="true" className="h-3.5 w-3.5" />
                {resource?.name}
              </button>
            ) : (
              <Unavailable reason="This pod could not be found - it may have been deleted or rescheduled." />
            )
          ) : (
            <span className="break-all font-semibold">{resource?.name}</span>
          )}
        </ChainLink>
      ) : null}

      {matchedPod ? (
        <PodDetailsDrawer
          pod={matchedPod}
          applicationId={applicationId ?? ""}
          open={podDrawerOpen}
          onClose={() => setPodDrawerOpen(false)}
        />
      ) : null}
    </div>
  );
}

function ChainLink({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="rounded-2xl border p-3" style={{ borderColor: "var(--border)" }}>
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</dt>
      <dd className="mt-1">{children}</dd>
    </div>
  );
}

function Loading() {
  return <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>Loading...</span>;
}

function Unavailable({ reason }: { reason: string }) {
  return <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>Not available — {reason}</span>;
}

function deploymentStatusVariant(status: string): "success" | "warning" | "critical" | "archived" | "info" {
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
