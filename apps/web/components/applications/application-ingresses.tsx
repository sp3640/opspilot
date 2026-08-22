"use client";

import { useMemo, useState } from "react";
import { Globe } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useIngressesByApplication } from "@/hooks/use-ingresses";

export function ApplicationIngresses({ applicationId }: { applicationId: string }) {
  const [selectedIngressKey, setSelectedIngressKey] = useState<string | null>(null);

  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = useIngressesByApplication(applicationId, params);

  const ingresses = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load ingresses. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (ingresses.length === 0) {
    return (
      <EmptyState
        icon={Globe}
        title="No ingresses found"
        description="Ingresses for this application will appear here once they are created."
      />
    );
  }

  return (
    <div className="space-y-3">
      {ingresses.map((ingress) => {
        const ingressKey = `${ingress.namespace}/${ingress.name}`;
        const isSelected = selectedIngressKey === ingressKey;
        const address = ingress.loadBalancerIP || ingress.loadBalancerHostname || "-";

        return (
          <button
            key={ingressKey}
            type="button"
            onClick={() => setSelectedIngressKey(ingressKey)}
            aria-pressed={isSelected}
            className="w-full rounded-2xl border p-4 text-left transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
            style={{ borderColor: isSelected ? "var(--primary)" : "var(--border)", backgroundColor: "var(--card)" }}
          >
            <div className="flex items-start justify-between gap-3">
              <p className="min-w-0 truncate font-semibold">{ingress.name}</p>
              {ingress.ingressClass ? (
                <span
                  className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
                  style={{
                    color: "var(--muted-foreground)",
                    backgroundColor: "color-mix(in srgb, var(--muted-foreground) 12%, transparent)",
                  }}
                >
                  {ingress.ingressClass}
                </span>
              ) : null}
            </div>
            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Namespace</dt>
                <dd className="mt-1 break-all">{ingress.namespace}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Hosts</dt>
                <dd className="mt-1 break-all">{formatList(ingress.hosts)}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Paths</dt>
                <dd className="mt-1 break-all">{formatList(ingress.paths)}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Address</dt>
                <dd className="mt-1 break-all">{address}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>TLS</dt>
                <dd className="mt-1">{ingress.tlsEnabled ? "Enabled" : "Disabled"}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
                <dd className="mt-1">{ingress.age || "-"}</dd>
              </div>
            </dl>
          </button>
        );
      })}
    </div>
  );
}

function formatList(values: string[]): string {
  return values && values.length > 0 ? values.join(", ") : "-";
}
