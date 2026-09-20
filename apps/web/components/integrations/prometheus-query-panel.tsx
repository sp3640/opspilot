"use client";

import { useState } from "react";
import { AlertTriangle, Play } from "lucide-react";

import { Button } from "@/components/ui/button";
import { usePrometheusQuery } from "@/hooks/use-prometheus";
import { getPrometheusQueryErrorMessage } from "@/lib/prometheus-query-error";
import type { IntegrationStatus } from "@/types/integration-api";
import type { PrometheusCorrelationResult, PrometheusSample } from "@/types/prometheus-api";

/**
 * Small, ad hoc PromQL query interface for a connected Prometheus
 * integration (Sprint 30). Deliberately not a Grafana replacement: one
 * query box, one "Run query" action, and the backend's own normalized
 * instant-query result - no dashboards, no continuous polling.
 */
export function PrometheusQueryPanel({
  integrationId,
  integrationStatus,
}: {
  integrationId: string;
  integrationStatus: IntegrationStatus;
}) {
  const [queryText, setQueryText] = useState("");
  const [submittedQuery, setSubmittedQuery] = useState("");
  const isConnected = integrationStatus === "CONNECTED";
  const hasSubmitted = submittedQuery.trim().length > 0;

  const { data, error, isError, isFetching } = usePrometheusQuery(
    integrationId,
    submittedQuery,
    isConnected && hasSubmitted
  );

  const handleRunQuery = () => {
    if (!isConnected || !queryText.trim()) return;
    setSubmittedQuery(queryText.trim());
  };

  return (
    <div className="mt-3 space-y-3 border-t pt-3" style={{ borderColor: "var(--border)" }}>
      <span className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
        PromQL query
      </span>

      {!isConnected ? (
        <p className="flex items-center gap-1.5 text-sm" style={{ color: "var(--muted-foreground)" }}>
          <AlertTriangle aria-hidden="true" className="h-4 w-4" />
          Connect this integration before running Prometheus queries.
        </p>
      ) : (
        <>
          <div className="flex flex-wrap gap-2">
            <input
              value={queryText}
              onChange={(event) => setQueryText(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") handleRunQuery();
              }}
              placeholder="rate(http_requests_total[5m])"
              aria-label="PromQL query"
              className="h-10 min-w-64 flex-1 rounded-xl border bg-transparent px-3 font-mono text-sm outline-none focus:border-[var(--primary)]"
              style={{ borderColor: "var(--border)" }}
            />
            <Button type="button" loading={isFetching} disabled={!queryText.trim()} onClick={handleRunQuery}>
              <Play aria-hidden="true" className="h-4 w-4" />
              Run query
            </Button>
          </div>

          <PrometheusQueryResult hasSubmitted={hasSubmitted} isFetching={isFetching} isError={isError} error={error} samples={data?.samples} correlation={data?.correlation} />
        </>
      )}
    </div>
  );
}

function PrometheusQueryResult({
  hasSubmitted,
  isFetching,
  isError,
  error,
  samples,
  correlation,
}: {
  hasSubmitted: boolean;
  isFetching: boolean;
  isError: boolean;
  error: unknown;
  samples?: PrometheusSample[];
  correlation?: PrometheusCorrelationResult;
}) {
  if (!hasSubmitted) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        Enter a PromQL query and run it to see results.
      </p>
    );
  }

  if (isFetching) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }} role="status">
        Running query&hellip;
      </p>
    );
  }

  if (isError) {
    return (
      <p className="flex items-center gap-1.5 text-sm" style={{ color: "var(--danger-text)" }} role="alert">
        <AlertTriangle aria-hidden="true" className="h-4 w-4" />
        {getPrometheusQueryErrorMessage(error)}
      </p>
    );
  }

  if (!samples) return null;

  if (samples.length === 0) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        No data returned for this query.
      </p>
    );
  }

  return (
    <div className="space-y-3">
      <div className="overflow-x-auto rounded-xl border" style={{ borderColor: "var(--border)" }}>
        <table className="w-full text-left text-sm">
          <thead>
            <tr style={{ color: "var(--muted-foreground)" }}>
              <th className="px-3 py-2 font-medium">Metric</th>
              <th className="px-3 py-2 font-medium">Labels</th>
              <th className="px-3 py-2 font-medium">Value</th>
              <th className="px-3 py-2 font-medium">Timestamp</th>
            </tr>
          </thead>
          <tbody>
            {samples.map((sample, index) => (
              <tr key={`${sample.timestamp}-${index}`} className="border-t" style={{ borderColor: "var(--border)" }}>
                <td className="px-3 py-2 font-mono">{sample.labels?.__name__ ?? "—"}</td>
                <td className="px-3 py-2 font-mono text-xs">{formatLabels(sample.labels)}</td>
                <td className="px-3 py-2">{sample.value}</td>
                <td className="px-3 py-2">{new Date(sample.timestamp).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {correlation && <PrometheusCorrelationSummary correlation={correlation} />}
    </div>
  );
}

function formatLabels(labels?: Record<string, string>): string {
  if (!labels) return "—";
  const entries = Object.entries(labels).filter(([key]) => key !== "__name__");
  if (entries.length === 0) return "—";
  return entries.map(([key, value]) => `${key}="${value}"`).join(", ");
}

function PrometheusCorrelationSummary({ correlation }: { correlation: PrometheusCorrelationResult }) {
  if (!correlation.available || (!correlation.resource && !correlation.application)) {
    return (
      <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        Correlation unavailable
      </p>
    );
  }

  return (
    <div className="space-y-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
      {correlation.resource && (
        <p>
          Resource: {correlation.resource.kind} {correlation.resource.name}
          {correlation.resource.namespace ? ` (${correlation.resource.namespace})` : ""}
        </p>
      )}
      {correlation.application && <p>Application: {correlation.application.name}</p>}
    </div>
  );
}
