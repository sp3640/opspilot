'use client';

import { Database, HardDrive, KeyRound, Server, Workflow } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useDashboardServices } from "@/hooks/use-dashboard-data";
import { HealthItem } from "./health-item";
import { SectionCard } from "./section-card";

const serviceIcons: Record<string, LucideIcon> = {
  API: Server,
  Database: Database,
  Storage: HardDrive,
  Queue: Workflow,
  Authentication: KeyRound,
};

/** Dependency availability and response-time overview for the current environment. */
export function SystemHealth() {
  const { data: services, isLoading } = useDashboardServices();

  if (isLoading) {
    return (
      <SectionCard
        title="System health"
        description="Live service availability for the production environment"
        action={
          <span
            className="inline-flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-semibold"
            style={{
              color: "var(--muted-foreground)",
              backgroundColor: "color-mix(in srgb, var(--muted) 20%, transparent)",
            }}
          >
            <span className="h-1.5 w-1.5 rounded-full animate-pulse" style={{ backgroundColor: "var(--muted-foreground)" }} />
            Loading...
          </span>
        }
      >
        <ul className="space-y-3" aria-label="Service health loading">
          {[...Array(3)].map((_, i) => (
            <li key={i} className="h-12 rounded-2xl bg-gray-700 animate-pulse" />
          ))}
        </ul>
      </SectionCard>
    );
  }

  if (!services || services.length === 0) {
    return (
      <SectionCard
        title="System health"
        description="Live service availability for the production environment"
        action={
          <span
            className="inline-flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-semibold"
            style={{
              color: "var(--warning)",
              backgroundColor: "color-mix(in srgb, var(--warning) 10%, transparent)",
            }}
          >
            <span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: "var(--warning)" }} />
            No data
          </span>
        }
      >
        <div
          className="rounded-2xl p-6 text-center"
          style={{ backgroundColor: "color-mix(in srgb, var(--muted) 50%, transparent)" }}
        >
          <p style={{ color: "var(--muted-foreground)" }}>No service health data available</p>
        </div>
      </SectionCard>
    );
  }

  return (
    <SectionCard
      title="System health"
      description="Live service availability for the production environment"
      action={
        <span
          className="inline-flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-semibold"
          style={{
            color: "var(--success)",
            backgroundColor: "color-mix(in srgb, var(--success) 10%, transparent)",
          }}
        >
          <span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: "var(--success)" }} />
          Live
        </span>
      }
    >
      <ul className="space-y-3" aria-label="Production service health">
        {services.map((service, index) => (
          <HealthItem
            key={`${service.name}-${index}`}
            name={service.name}
            status={service.status === "healthy" ? "healthy" : "degraded"}
            responseTime={`${service.responseTime} ms`}
            health={service.health}
            icon={serviceIcons[service.name] || Server}
          />
        ))}
      </ul>
    </SectionCard>
  );
}
