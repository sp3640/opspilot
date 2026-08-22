"use client";

import { useMemo } from "react";
import { Network } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useServicesByCluster } from "@/hooks/use-services";
import type { ServicePortResponse, ServiceResponse } from "@/types/service-api";

export function ClusterServices({ clusterId }: { clusterId: string }) {
  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = useServicesByCluster(clusterId, params);

  const services = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load services. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (services.length === 0) {
    return (
      <EmptyState
        icon={Network}
        title="No services found"
        description="Services discovered from this cluster will appear here."
      />
    );
  }

  return (
    <div className="space-y-3">
      {services.map((service) => {
        const externalIP = getExternalIP(service);

        return (
          <div
            key={`${service.namespace}/${service.name}`}
            className="w-full rounded-2xl border p-4 text-left"
            style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
          >
            <div className="flex items-start justify-between gap-3">
              <p className="min-w-0 truncate font-semibold">{service.name}</p>
              <span
                className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
                style={{
                  color: "var(--muted-foreground)",
                  backgroundColor: "color-mix(in srgb, var(--muted-foreground) 12%, transparent)",
                }}
              >
                {service.type}
              </span>
            </div>
            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Namespace</dt>
                <dd className="mt-1 break-all">{service.namespace}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Cluster IP</dt>
                <dd className="mt-1 break-all">{service.clusterIP || "-"}</dd>
              </div>
              {externalIP ? (
                <div>
                  <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>External IP</dt>
                  <dd className="mt-1 break-all">{externalIP}</dd>
                </div>
              ) : null}
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Ports</dt>
                <dd className="mt-1 break-all">{formatPorts(service.ports)}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
                <dd className="mt-1">{service.age || "-"}</dd>
              </div>
            </dl>
          </div>
        );
      })}
    </div>
  );
}

function getExternalIP(service: ServiceResponse): string {
  if (service.externalIPs && service.externalIPs.length > 0) {
    return service.externalIPs.join(", ");
  }
  return service.loadBalancerIP ?? "";
}

function formatPorts(ports: ServicePortResponse[]): string {
  if (!ports || ports.length === 0) return "-";
  return ports.map((port) => `${port.port} → ${port.targetPort}/${port.protocol}`).join(", ");
}
