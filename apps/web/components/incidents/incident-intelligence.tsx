"use client";

import type { ReactNode } from "react";
import {
  AlertCircle,
  AlertOctagon,
  AlertTriangle,
  Bell,
  Container,
  Filter,
  HelpCircle,
  History,
  Layers,
  Rocket,
  ScrollText,
  Settings2,
  ShieldAlert,
  Sparkles,
  type LucideIcon,
} from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton, StatusBadge } from "@/components/common";
import { useIncidentRCA } from "@/hooks/use-incident-rca";
import type {
  IncidentRCAResponse,
  RCAEvidenceLevel,
  RCAPossibleCause,
  RCATimelineEvent,
} from "@/types/incident-rca-api";
import type { IncidentResponse } from "@/types/incident-api";

const EVIDENCE_LEVEL_META: Record<RCAEvidenceLevel, { label: string; color: string; icon: LucideIcon }> = {
  INSUFFICIENT: { label: "Insufficient evidence", color: "var(--muted-foreground)", icon: HelpCircle },
  WEAK: { label: "Weak signal", color: "var(--warning)", icon: AlertCircle },
  MODERATE: { label: "Moderate signal", color: "var(--warning)", icon: AlertTriangle },
  STRONG: { label: "Strong signal", color: "var(--danger)", icon: AlertOctagon },
};

const TIMELINE_ICONS: Record<RCATimelineEvent["type"], LucideIcon> = {
  incident: ShieldAlert,
  alert: Bell,
  deployment: Rocket,
  config_change: Settings2,
  kubernetes_event: Container,
};

/**
 * Incident Intelligence: the deterministic evidence/correlation layer for
 * Root Cause Analysis (Phase 25). Every possible cause below is a
 * correlation over already-real evidence computed server-side (see
 * apps/backend/internal/rca) - nothing here is AI-generated. Wording is
 * deliberately hedged throughout ("possible cause", "may be related",
 * "evidence suggests") and must stay that way: never render this data as
 * a confirmed root cause. A human engineer remains responsible for
 * reviewing and approving any remediation.
 */
export function IncidentIntelligence({ incident }: { incident: IncidentResponse }) {
  const { data: rca, error, isError, isLoading, refetch } = useIncidentRCA(incident.id);

  if (isLoading) {
    return <ListSkeleton rows={4} />;
  }

  if (isError || !rca) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to generate incident intelligence for this incident."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  return (
    <div className="space-y-8">
      <EvidenceLevelBanner rca={rca} />
      <Summary rca={rca} />
      <PossibleCauses causes={rca.possibleCauses} />
      <RecommendedActions rca={rca} />
      <Timeline events={rca.timeline} />
      <Evidence rca={rca} />
    </div>
  );
}

function EvidenceLevelBanner({ rca }: { rca: IncidentRCAResponse }) {
  const meta = EVIDENCE_LEVEL_META[rca.evidenceLevel];

  return (
    <div
      className="flex items-start gap-3 rounded-2xl border p-4"
      style={{ borderColor: meta.color, backgroundColor: `color-mix(in srgb, ${meta.color} 8%, transparent)` }}
    >
      <meta.icon aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0" style={{ color: meta.color }} />
      <div>
        <p className="font-semibold" style={{ color: meta.color }}>
          {meta.label}
        </p>
        <p className="mt-1 text-sm leading-6" style={{ color: "var(--foreground)" }}>
          {rca.evidenceLevelReason}
        </p>
      </div>
    </div>
  );
}

function Summary({ rca }: { rca: IncidentRCAResponse }) {
  return (
    <section>
      <SubHeading icon={Sparkles} title="Summary" />
      <p className="text-sm leading-6" style={{ color: "var(--foreground)" }}>
        {rca.summary}
      </p>
      <p className="mt-2 text-xs" style={{ color: "var(--muted-foreground)" }}>
        Evidence window: {formatDateTime(rca.windowStart)} &ndash; {formatDateTime(rca.windowEnd)}
        {rca.applicationKnown && rca.affectedApplicationName ? ` · Affected application: ${rca.affectedApplicationName}` : ""}
      </p>
    </section>
  );
}

function PossibleCauses({ causes }: { causes: RCAPossibleCause[] }) {
  return (
    <section>
      <SubHeading icon={Filter} title="Possible Causes" />
      {causes.length === 0 ? (
        <EmptyState
          icon={HelpCircle}
          title="No possible cause identified"
          description="Nothing in the available evidence correlates strongly enough to suggest a possible cause yet. Continue investigating using the evidence below."
        />
      ) : (
        <ul className="space-y-3">
          {causes.map((cause) => {
            const meta = EVIDENCE_LEVEL_META[cause.confidence];
            return (
              <li
                key={cause.title}
                className="rounded-2xl border p-4"
                style={{ borderColor: "var(--border)", backgroundColor: "var(--card)" }}
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <p className="text-sm font-semibold" style={{ color: "var(--foreground)" }}>
                    {cause.title}
                  </p>
                  <StatusBadge
                    variant={cause.confidence === "STRONG" ? "critical" : cause.confidence === "MODERATE" ? "warning" : "info"}
                  >
                    {meta.label}
                  </StatusBadge>
                </div>
                <ul className="mt-3 list-disc space-y-1 pl-5 text-sm" style={{ color: "var(--foreground)" }}>
                  {cause.evidence.map((line) => (
                    <li key={line}>{line}</li>
                  ))}
                </ul>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}

function RecommendedActions({ rca }: { rca: IncidentRCAResponse }) {
  return (
    <section>
      <SubHeading icon={Layers} title="Recommended Actions" />
      <div className="grid gap-4 sm:grid-cols-2">
        <div>
          <h4 className="text-xs font-semibold uppercase tracking-wide" style={{ color: "var(--muted-foreground)" }}>
            Investigation
          </h4>
          <ol className="mt-2 list-decimal space-y-1 pl-5 text-sm" style={{ color: "var(--foreground)" }}>
            {rca.recommendedInvestigationSteps.map((step) => (
              <li key={step}>{step}</li>
            ))}
          </ol>
        </div>
        <div>
          <h4 className="text-xs font-semibold uppercase tracking-wide" style={{ color: "var(--muted-foreground)" }}>
            Remediation
          </h4>
          <ol className="mt-2 list-decimal space-y-1 pl-5 text-sm" style={{ color: "var(--foreground)" }}>
            {rca.recommendedRemediation.map((step) => (
              <li key={step}>{step}</li>
            ))}
          </ol>
        </div>
      </div>
      <p className="mt-4 text-xs italic" style={{ color: "var(--muted-foreground)" }}>
        This analysis does not authorize any action. A human engineer remains responsible for reviewing and approving
        remediation before it is applied.
      </p>
    </section>
  );
}

function Timeline({ events }: { events: RCATimelineEvent[] }) {
  return (
    <section>
      <SubHeading icon={History} title="Timeline" />
      {events.length === 0 ? (
        <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
          No timestamped evidence was available to build a timeline.
        </p>
      ) : (
        <ol className="space-y-2">
          {events.map((event, index) => {
            const Icon = TIMELINE_ICONS[event.type];
            return (
              <li
                key={`${event.type}-${event.timestamp}-${index}`}
                className="rounded-2xl border p-3"
                style={{ borderColor: "var(--border)" }}
              >
                <div className="flex items-start gap-3">
                  <div
                    className="mt-0.5 rounded-xl p-1.5"
                    style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}
                  >
                    <Icon aria-hidden="true" className="h-3.5 w-3.5" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <span className="text-sm font-medium">{event.title}</span>
                      <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                        {formatDateTime(event.timestamp)}
                      </span>
                    </div>
                    {event.description ? (
                      <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                        {event.description}
                      </p>
                    ) : null}
                  </div>
                </div>
              </li>
            );
          })}
        </ol>
      )}
    </section>
  );
}

function Evidence({ rca }: { rca: IncidentRCAResponse }) {
  return (
    <section>
      <SubHeading icon={ScrollText} title="Evidence" />
      <div className="grid gap-4 lg:grid-cols-2">
        <EvidenceGroup title="Correlated Alerts" empty="No alerts are attached to this incident.">
          {rca.correlatedAlerts.map((alert) => (
            <EvidenceRow key={alert.id} primary={alert.title} secondary={`${alert.severity} · ${alert.status}`} timestamp={alert.firstSeenAt} />
          ))}
        </EvidenceGroup>

        <EvidenceGroup title="Relevant Metrics" empty="No metric shift was detected within the evidence window.">
          {rca.relevantMetrics.map((metric) => (
            <EvidenceRow
              key={`${metric.metricType}-${metric.metricName}`}
              primary={metric.metricName || metric.metricType}
              secondary={`${metric.direction} · ${metric.before.toFixed(2)} → ${metric.after.toFixed(2)} ${metric.unit ?? ""} (${metric.changePercent > 0 ? "+" : ""}${metric.changePercent.toFixed(1)}%)`}
            />
          ))}
        </EvidenceGroup>

        <EvidenceGroup title="Relevant Logs" empty="No log lines matched known error/crash keywords in the sampled pods.">
          {rca.relevantLogs.map((log, index) => (
            <EvidenceRow key={`${log.podName}-${index}`} primary={log.podName} secondary={log.line} badge={log.keyword} />
          ))}
        </EvidenceGroup>

        <EvidenceGroup title="Kubernetes Evidence" empty="No relevant Kubernetes events were found in the affected namespace.">
          {rca.kubernetesEvidence.map((event, index) => (
            <EvidenceRow
              key={`${event.reason}-${index}`}
              primary={`${event.reason}: ${event.involvedObject}`}
              secondary={event.message}
              timestamp={event.lastTimestamp}
            />
          ))}
        </EvidenceGroup>

        <EvidenceGroup title="Pod Health" empty="No pod status was available for the affected application.">
          {rca.podEvidence.map((pod) => (
            <EvidenceRow
              key={pod.name}
              primary={pod.name}
              secondary={`${pod.ready ? "Ready" : "Not ready"} · ${pod.restartCount} restart(s)${pod.reason ? ` · ${pod.reason}` : ""}`}
            />
          ))}
        </EvidenceGroup>

        <EvidenceGroup title="Recent Changes" empty="No configuration changes were recorded in the evidence window.">
          {rca.recentChanges.map((change, index) => (
            <EvidenceRow
              key={`${change.entityType}-${change.entityId}-${index}`}
              primary={`${change.action} ${change.entityType} ${change.entityId}`}
              secondary={change.fieldName ? `${change.fieldName}: ${change.oldValue ?? ""} → ${change.newValue ?? ""}` : undefined}
              timestamp={change.changedAt}
            />
          ))}
        </EvidenceGroup>
      </div>
    </section>
  );
}

function EvidenceGroup({ title, empty, children }: { title: string; empty: string; children: ReactNode }) {
  const items = Array.isArray(children) ? children : [children];
  const hasItems = items.filter(Boolean).length > 0;

  return (
    <div className="rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
      <h4 className="text-xs font-semibold uppercase tracking-wide" style={{ color: "var(--muted-foreground)" }}>
        {title}
      </h4>
      {hasItems ? (
        <ul className="mt-3 space-y-2">{children}</ul>
      ) : (
        <p className="mt-3 text-sm" style={{ color: "var(--muted-foreground)" }}>
          {empty}
        </p>
      )}
    </div>
  );
}

function EvidenceRow({
  primary,
  secondary,
  timestamp,
  badge,
}: {
  primary: string;
  secondary?: string;
  timestamp?: string;
  badge?: string;
}) {
  return (
    <li className="text-sm">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="font-medium" style={{ color: "var(--foreground)" }}>
          {primary}
        </span>
        {timestamp ? (
          <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>
            {formatDateTime(timestamp)}
          </span>
        ) : null}
      </div>
      {secondary ? (
        <p className="mt-0.5 truncate text-xs" style={{ color: "var(--muted-foreground)" }} title={secondary}>
          {secondary}
        </p>
      ) : null}
      {badge ? (
        <span
          className="mt-1 inline-block rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase"
          style={{ color: "var(--warning)", backgroundColor: "color-mix(in srgb, var(--warning) 12%, transparent)" }}
        >
          {badge}
        </span>
      ) : null}
    </li>
  );
}

function SubHeading({ icon: Icon, title }: { icon: LucideIcon; title: string }) {
  return (
    <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide" style={{ color: "var(--muted-foreground)" }}>
      <Icon aria-hidden="true" className="h-4 w-4" />
      {title}
    </h3>
  );
}

function formatDateTime(value?: string) {
  if (!value) return "";
  return new Date(value).toLocaleString();
}
