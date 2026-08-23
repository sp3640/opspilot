"use client";

import { ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useApplicationHealthScore } from "@/hooks/use-application-health-score";
import type { ApplicationHealthFactorResponse, ApplicationHealthState } from "@/types/application-health-score-api";

/**
 * The unified health score (Phase 21): one deterministic 0-100 number
 * (or "Unknown" when no factor has real data at all) plus the exact factor
 * breakdown that produced it. Every factor always renders, including ones
 * that are permanently Unknown today (Error Rate, Latency, Resource
 * Utilization have no data source anywhere in OpsPilot) - each one's
 * `reason` explains why, so "why did the score change" is always
 * answerable from what's on screen, never hidden in a black box.
 */
export function ApplicationHealthScore({ applicationId }: { applicationId: string }) {
  const { data, error, isError, isLoading, refetch } = useApplicationHealthScore(applicationId);

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load this application's health score."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading || !data) {
    return <ListSkeleton rows={3} />;
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-baseline gap-2">
          <span className="text-3xl font-semibold tracking-tight">{data.score ?? "—"}</span>
          <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>
            {data.score !== undefined ? "/ 100" : "no score available"}
          </span>
        </div>
        <StatusBadge variant={stateVariant(data.state)}>{stateLabel(data.state)}</StatusBadge>
      </div>

      <dl className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        {data.factors.map((factor) => (
          <FactorRow key={factor.key} factor={factor} />
        ))}
      </dl>
    </div>
  );
}

function FactorRow({ factor }: { factor: ApplicationHealthFactorResponse }) {
  return (
    <div className="rounded-2xl border p-3" style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}>
      <div className="flex items-center justify-between gap-2">
        <dt className="text-sm font-medium">{factor.label}</dt>
        <StatusBadge variant={stateVariant(factor.state)}>{stateLabel(factor.state)}</StatusBadge>
      </div>
      <dd className="mt-1.5 text-xs leading-5" style={{ color: "var(--muted-foreground)" }}>
        {factor.reason}
      </dd>
    </div>
  );
}

function stateLabel(state: ApplicationHealthState): string {
  switch (state) {
    case "HEALTHY":
      return "Healthy";
    case "DEGRADED":
      return "Degraded";
    case "CRITICAL":
      return "Critical";
    default:
      return "Unknown";
  }
}

function stateVariant(state: ApplicationHealthState): "success" | "warning" | "critical" | "archived" {
  switch (state) {
    case "HEALTHY":
      return "success";
    case "DEGRADED":
      return "warning";
    case "CRITICAL":
      return "critical";
    default:
      return "archived";
  }
}
