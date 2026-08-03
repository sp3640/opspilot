"use client";

import { Ellipsis, ExternalLink, ShieldCheck } from "lucide-react";

import { Button } from "@/components/ui/button";

import { ClusterProviderBadge, ClusterStatusBadge } from "./cluster-status";
import type { Cluster } from "./types";

type ClusterCardProps = {
  cluster: Cluster;
  onOpen: (clusterID: string) => void;
};

export function ClusterCard({ cluster, onOpen }: ClusterCardProps) {
  return (
    <article
      onClick={() => onOpen(cluster.id)}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onOpen(cluster.id);
        }
      }}
      role="button"
      tabIndex={0}
      aria-label={`Open cluster details for ${cluster.name}`}
      className="group cursor-pointer rounded-3xl border p-5 shadow-[var(--shadow-sm)] transition-all duration-300 hover:-translate-y-1 hover:shadow-[var(--shadow-md)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <h2 className="truncate font-semibold tracking-tight">{cluster.name}</h2>
            {cluster.isDefault ? (
              <span className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide" style={{ color: "var(--success)", backgroundColor: "color-mix(in srgb, var(--success) 12%, transparent)" }}>
                <ShieldCheck aria-hidden={true} className="h-3 w-3" />
                Default
              </span>
            ) : null}
          </div>
          <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
            {cluster.apiEndpoint || "No API endpoint configured"}
          </p>
        </div>
        <Button
          type="button"
          variant="ghost"
          onClick={(event) => {
            event.stopPropagation();
            onOpen(cluster.id);
          }}
          className="h-9 w-9 rounded-xl p-0"
          aria-label={`Open ${cluster.name} actions`}
        >
          <Ellipsis aria-hidden="true" className="h-4 w-4" />
        </Button>
      </div>

      <div className="mt-5 flex flex-wrap gap-2">
        <ClusterStatusBadge status={cluster.status} />
        <ClusterProviderBadge provider={cluster.provider} />
      </div>

      <dl className="mt-5 grid grid-cols-2 gap-3 border-y py-4 text-sm" style={{ borderColor: "var(--border)" }}>
        <Meta label="Version" value={cluster.version || "-"} />
        <Meta label="Region" value={cluster.region || "-"} />
        <Meta label="Last validation" value={formatDateTime(cluster.lastValidatedAt)} />
        <Meta label="Last discovery" value={formatDateTime(cluster.lastDiscoveryAt)} />
      </dl>

      <div className="mt-4 flex items-center justify-between gap-3">
        <span className="inline-flex items-center gap-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Project {cluster.projectId}
        </span>
        <span className="inline-flex items-center gap-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Updated {new Date(cluster.updatedAt).toLocaleDateString()}
          <ExternalLink aria-hidden="true" className="h-3 w-3 opacity-0 transition-opacity group-hover:opacity-100" />
        </span>
      </div>
    </article>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        {label}
      </dt>
      <dd className="mt-1 truncate font-semibold">{value}</dd>
    </div>
  );
}

function formatDateTime(value?: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}
