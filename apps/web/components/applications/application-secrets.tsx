"use client";

import { useMemo, useState } from "react";
import { ShieldCheck } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { useSecretsByApplication } from "@/hooks/use-secrets";

export function ApplicationSecrets({ applicationId }: { applicationId: string }) {
  const [selectedSecretKey, setSelectedSecretKey] = useState<string | null>(null);

  const params = useMemo(() => ({}), []);
  const { data, error, isError, isLoading, refetch } = useSecretsByApplication(applicationId, params);

  const secrets = data?.items ?? [];

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load secrets. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (secrets.length === 0) {
    return (
      <EmptyState
        icon={ShieldCheck}
        title="No secrets found"
        description="Secrets for this application will appear here once they are created."
      />
    );
  }

  return (
    <div className="space-y-3">
      {secrets.map((secret) => {
        const secretKey = `${secret.namespace}/${secret.name}`;
        const isSelected = selectedSecretKey === secretKey;

        return (
          <button
            key={secretKey}
            type="button"
            onClick={() => setSelectedSecretKey(secretKey)}
            aria-pressed={isSelected}
            className="w-full rounded-2xl border p-4 text-left transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
            style={{ borderColor: isSelected ? "var(--primary)" : "var(--border)", backgroundColor: "var(--card)" }}
          >
            <div className="flex items-start justify-between gap-3">
              <p className="min-w-0 truncate font-semibold">{secret.name}</p>
              <span
                className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
                style={{
                  color: "var(--muted-foreground)",
                  backgroundColor: "color-mix(in srgb, var(--muted-foreground) 12%, transparent)",
                }}
              >
                {secret.type}
              </span>
            </div>
            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Namespace</dt>
                <dd className="mt-1 break-all">{secret.namespace}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Keys</dt>
                <dd className="mt-1">{secret.dataKeyCount}</dd>
              </div>
              <div>
                <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Age</dt>
                <dd className="mt-1">{secret.age || "-"}</dd>
              </div>
            </dl>
          </button>
        );
      })}
    </div>
  );
}
