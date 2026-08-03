"use client";

import { ChevronRight, ShieldCheck } from "lucide-react";

import { ClusterProviderBadge, ClusterStatusBadge } from "./cluster-status";
import type { Cluster } from "./types";

type ClusterTableProps = {
  clusters: Cluster[];
  onOpen: (clusterID: string) => void;
};

export function ClusterTable({ clusters, onOpen }: ClusterTableProps) {
  return (
    <div className="overflow-x-auto rounded-2xl border" style={{ borderColor: "var(--border)" }}>
      <table className="w-full min-w-[1120px] text-left text-sm">
        <caption className="sr-only">Clusters list with provider, status, endpoint, and validation metadata</caption>
        <thead
          className="text-xs uppercase tracking-[0.1em]"
          style={{
            color: "var(--muted-foreground)",
            backgroundColor: "color-mix(in srgb, var(--muted) 55%, transparent)",
          }}
        >
          <tr>
            <th scope="col" className="px-5 py-3 font-semibold">Name</th>
            <th scope="col" className="px-5 py-3 font-semibold">Provider</th>
            <th scope="col" className="px-5 py-3 font-semibold">Status</th>
            <th scope="col" className="px-5 py-3 font-semibold">API endpoint</th>
            <th scope="col" className="px-5 py-3 font-semibold">Kubernetes version</th>
            <th scope="col" className="px-5 py-3 font-semibold">Region</th>
            <th scope="col" className="px-5 py-3 font-semibold">Last validation</th>
            <th scope="col" className="px-5 py-3 font-semibold">Last discovery</th>
            <th scope="col" className="px-5 py-3 font-semibold">Default</th>
            <th scope="col" className="w-12 px-5 py-3"><span className="sr-only">Open</span></th>
          </tr>
        </thead>
        <tbody>
          {clusters.map((cluster) => (
            <tr
              key={cluster.id}
              role="button"
              tabIndex={0}
              aria-label={`Open cluster details for ${cluster.name}`}
              onClick={() => onOpen(cluster.id)}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onOpen(cluster.id);
                }
              }}
              className="cursor-pointer border-t transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
              style={{ borderColor: "var(--border)" }}
            >
              <td className="px-5 py-4 font-medium">{cluster.name}</td>
              <td className="px-5 py-4"><ClusterProviderBadge provider={cluster.provider} /></td>
              <td className="px-5 py-4"><ClusterStatusBadge status={cluster.status} /></td>
              <td className="max-w-[280px] truncate px-5 py-4" title={cluster.apiEndpoint || ""}>
                {cluster.apiEndpoint || "-"}
              </td>
              <td className="px-5 py-4">{cluster.version || "-"}</td>
              <td className="px-5 py-4">{cluster.region || "-"}</td>
              <td className="px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                {formatDateTime(cluster.lastValidatedAt)}
              </td>
              <td className="px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                {formatDateTime(cluster.lastDiscoveryAt)}
              </td>
              <td className="px-5 py-4">
                {cluster.isDefault ? (
                  <span className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold" style={{ color: "var(--success)", backgroundColor: "color-mix(in srgb, var(--success) 12%, transparent)" }}>
                    <ShieldCheck aria-hidden={true} className="h-3.5 w-3.5" />
                    Default
                  </span>
                ) : (
                  <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>No</span>
                )}
              </td>
              <td className="px-5 py-4">
                <ChevronRight aria-hidden="true" className="h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function formatDateTime(value?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}
