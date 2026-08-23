"use client";

import { useEffect, useState } from "react";
import axios from "axios";

import { StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useHasPermission } from "@/store/auth-store";
import { useApplicationSLO, useApplicationSREMetrics, useConfigureApplicationSLO } from "@/hooks/use-sre";
import type { MetricValue } from "@/types/sre-api";

/**
 * SLO configuration + computed SRE metrics for one application. Every
 * metric is either a real value (with the exact formula the backend used to
 * derive it) or "Insufficient data" - never fabricated. See
 * docs/monitoring/SRE_METRICS.md for the full methodology.
 */
export function ApplicationSLO({ applicationId }: { applicationId: string }) {
  const canManageSLO = useHasPermission("slo:manage");

  const {
    data: sloConfig,
    error: sloError,
    isError: isSLOError,
    isLoading: isSLOLoading,
  } = useApplicationSLO(applicationId);
  const sloNotConfigured = axios.isAxiosError(sloError) && sloError.response?.status === 404;

  const {
    data: metrics,
    error: metricsError,
    isError: isMetricsError,
    isLoading: isMetricsLoading,
    refetch: refetchMetrics,
  } = useApplicationSREMetrics(applicationId);

  const configureSLO = useConfigureApplicationSLO();
  const [targetPercentage, setTargetPercentage] = useState("99.9");
  const [windowDays, setWindowDays] = useState("30");

  useEffect(() => {
    if (sloConfig) {
      setTargetPercentage(String(sloConfig.targetPercentage));
      setWindowDays(String(sloConfig.windowDays));
    }
  }, [sloConfig]);

  const handleSave = () => {
    const target = Number(targetPercentage);
    const window = Number(windowDays);
    if (!Number.isFinite(target) || target < 1 || target > 100) return;
    if (!Number.isFinite(window) || window < 1 || window > 365) return;

    void configureSLO.mutateAsync({
      applicationId,
      payload: { target_percentage: target, window_days: window },
    });
  };

  return (
    <div className="space-y-6">
      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          SLO Configuration
        </h3>
        <div className="mt-3">
          {isSLOLoading ? (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
              Loading...
            </p>
          ) : isSLOError && !sloNotConfigured ? (
            <p className="text-sm" style={{ color: "var(--danger)" }}>
              {metricsError instanceof Error ? metricsError.message : "Unable to load SLO configuration."}
            </p>
          ) : (
            <div className="space-y-3">
              {sloNotConfigured && (
                <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
                  No SLO configured yet for this application.
                </p>
              )}
              <div className="flex flex-wrap items-end gap-3">
                <label className="flex flex-col gap-1 text-sm">
                  <span style={{ color: "var(--muted-foreground)" }}>Target %</span>
                  <input
                    type="number"
                    min={1}
                    max={100}
                    step={0.01}
                    value={targetPercentage}
                    onChange={(event) => setTargetPercentage(event.target.value)}
                    disabled={!canManageSLO}
                    className="h-10 w-32 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)] disabled:opacity-60"
                    style={{ borderColor: "var(--border)" }}
                  />
                </label>
                <label className="flex flex-col gap-1 text-sm">
                  <span style={{ color: "var(--muted-foreground)" }}>Window (days)</span>
                  <input
                    type="number"
                    min={1}
                    max={365}
                    value={windowDays}
                    onChange={(event) => setWindowDays(event.target.value)}
                    disabled={!canManageSLO}
                    className="h-10 w-32 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)] disabled:opacity-60"
                    style={{ borderColor: "var(--border)" }}
                  />
                </label>
                {canManageSLO && (
                  <Button type="button" loading={configureSLO.isPending} onClick={handleSave}>
                    Save SLO
                  </Button>
                )}
              </div>
            </div>
          )}
        </div>
      </section>

      <section>
        <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          SRE Metrics
        </h3>
        <div className="mt-3">
          {isMetricsLoading ? (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
              Loading...
            </p>
          ) : isMetricsError ? (
            <div className="space-y-2">
              <p className="text-sm" style={{ color: "var(--danger)" }}>
                {metricsError instanceof Error ? metricsError.message : "Unable to load SRE metrics."}
              </p>
              <Button type="button" variant="secondary" onClick={() => void refetchMetrics()}>
                Retry
              </Button>
            </div>
          ) : metrics ? (
            <div className="space-y-3">
              <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                Observed {metrics.observedDays.toFixed(1)} of {metrics.windowDays} window days
                {metrics.incidentCount > 0 ? ` · ${metrics.incidentCount} incident${metrics.incidentCount === 1 ? "" : "s"} in window` : ""}
              </p>
              <dl className="grid grid-cols-2 gap-3 lg:grid-cols-3">
                <SREMetricTile label="Availability" metric={metrics.availability} />
                <SREMetricTile label="Uptime" metric={metrics.uptime} />
                <SREMetricTile
                  label="Current SLO"
                  metric={
                    metrics.currentSlo !== undefined
                      ? { available: true, formatted: `${metrics.currentSlo}%`, formula: "The configured SLO target." }
                      : { available: false, formatted: "Insufficient data", reason: "SLO not configured", formula: "" }
                  }
                />
                <SREMetricTile label="SLO Compliance" metric={metrics.sloCompliance} isCompliance />
                <SREMetricTile label="Error Budget Remaining" metric={metrics.errorBudget} />
                <SREMetricTile label="MTTR" metric={metrics.mttr} />
                <SREMetricTile label="MTTA" metric={metrics.mtta} />
              </dl>
            </div>
          ) : null}
        </div>
      </section>
    </div>
  );
}

function SREMetricTile({
  label,
  metric,
  isCompliance = false,
}: {
  label: string;
  metric: MetricValue;
  isCompliance?: boolean;
}) {
  return (
    <div
      className="rounded-2xl border p-4"
      style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}
      title={metric.formula || undefined}
    >
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        {label}
      </dt>
      <dd className="mt-1">
        {metric.available ? (
          isCompliance ? (
            <StatusBadge variant={metric.formatted === "Compliant" ? "success" : "critical"}>{metric.formatted}</StatusBadge>
          ) : (
            <span className="font-semibold">{metric.formatted}</span>
          )
        ) : (
          <div>
            <span className="text-sm font-medium" style={{ color: "var(--muted-foreground)" }}>
              Insufficient data
            </span>
            {metric.reason && (
              <p className="mt-0.5 text-xs" style={{ color: "var(--muted-foreground)" }}>
                {metric.reason}
              </p>
            )}
          </div>
        )}
      </dd>
    </div>
  );
}
