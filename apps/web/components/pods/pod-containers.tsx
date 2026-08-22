"use client";

import { EmptyState, StatusBadge } from "@/components/common";
import { Boxes } from "lucide-react";
import type { ContainerStatusResponse, PodResponse } from "@/types/pod-api";

/** Per-container detail: image, state, restarts, resources, and probes. */
export function PodContainers({ pod }: { pod: PodResponse }) {
  const containers = pod.containerStatuses ?? [];

  if (containers.length === 0) {
    return (
      <EmptyState
        icon={Boxes}
        title="No container statuses"
        description="Container status is reported once Kubernetes schedules this pod."
      />
    );
  }

  const metricsByContainer = new Map((pod.metrics?.containers ?? []).map((metric) => [metric.name, metric]));

  return (
    <div className="space-y-4">
      {containers.map((container) => {
        const metrics = metricsByContainer.get(container.name);
        const unhealthy = isContainerUnhealthy(container);

        return (
          <div
            key={container.name}
            className="rounded-2xl border p-4"
            style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
          >
            <div className="flex items-start justify-between gap-3">
              <p className="min-w-0 truncate font-semibold">{container.name}</p>
              <StatusBadge variant={container.ready ? "success" : unhealthy ? "critical" : "warning"}>
                {container.state}
              </StatusBadge>
            </div>
            <p className="mt-1 break-all text-xs" style={{ color: "var(--muted-foreground)" }}>{container.image}</p>

            {container.stateMessage ? (
              <p
                className="mt-3 rounded-xl border p-3 text-xs"
                style={{ borderColor: "var(--danger)", color: "var(--danger)", backgroundColor: "color-mix(in srgb, var(--danger) 8%, transparent)" }}
              >
                {container.stateMessage}
              </p>
            ) : null}

            {container.lastTerminationReason ? (
              <p className="mt-2 text-xs" style={{ color: "var(--muted-foreground)" }}>
                Last terminated: {container.lastTerminationReason}
                {typeof container.lastTerminationExitCode === "number" ? ` (exit code ${container.lastTerminationExitCode})` : ""}
                {container.lastTerminationFinishedAt ? ` at ${new Date(container.lastTerminationFinishedAt).toLocaleString()}` : ""}
              </p>
            ) : null}

            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm sm:grid-cols-3">
              <Stat label="Ready">{container.ready ? "Yes" : "No"}</Stat>
              <Stat label="Started">{container.started ? "Yes" : "No"}</Stat>
              <Stat label="Restarts">{container.restartCount}</Stat>
              {typeof container.exitCode === "number" ? <Stat label="Exit code">{container.exitCode}</Stat> : null}
              <Stat label="Readiness probe">{container.hasReadinessProbe ? "Configured" : "Not configured"}</Stat>
              <Stat label="Liveness probe">{container.hasLivenessProbe ? "Configured" : "Not configured"}</Stat>
              <Stat label="CPU">{formatResource(metrics?.cpu, container.cpuRequest, container.cpuLimit)}</Stat>
              <Stat label="Memory">{formatResource(metrics?.memory, container.memoryRequest, container.memoryLimit)}</Stat>
            </dl>
          </div>
        );
      })}
    </div>
  );
}

function Stat({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</dt>
      <dd className="mt-1">{children}</dd>
    </div>
  );
}

function formatResource(usage: string | undefined, request: string | undefined, limit: string | undefined): string {
  const parts: string[] = [];
  if (usage) parts.push(`${usage} used`);
  if (request) parts.push(`${request} requested`);
  if (limit) parts.push(`${limit} limit`);

  return parts.length > 0 ? parts.join(" / ") : "Not available";
}

function isContainerUnhealthy(container: ContainerStatusResponse): boolean {
  return container.state === "CrashLoopBackOff" || container.state === "OOMKilled" || container.stateReason === "CrashLoopBackOff";
}
