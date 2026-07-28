import { Database, HardDrive, KeyRound, Server, Workflow } from "lucide-react";

import { HealthItem } from "./health-item";
import { SectionCard } from "./section-card";

const services = [
  { name: "API", status: "healthy" as const, responseTime: "82 ms", health: 100, icon: Server },
  { name: "Database", status: "healthy" as const, responseTime: "14 ms", health: 99, icon: Database },
  { name: "Storage", status: "healthy" as const, responseTime: "39 ms", health: 100, icon: HardDrive },
  { name: "Queue", status: "degraded" as const, responseTime: "215 ms", health: 92, icon: Workflow },
  { name: "Authentication", status: "healthy" as const, responseTime: "61 ms", health: 99, icon: KeyRound },
];

/** Dependency availability and response-time overview for the current environment. */
export function SystemHealth() {
  return (
    <SectionCard title="System health" description="Live service availability for the production environment" action={<span className="inline-flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-semibold" style={{ color: "var(--success)", backgroundColor: "color-mix(in srgb, var(--success) 10%, transparent)" }}><span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: "var(--success)" }} />Live</span>}>
      <ul className="space-y-3" aria-label="Production service health">
        {services.map((service) => <HealthItem key={service.name} {...service} />)}
      </ul>
    </SectionCard>
  );
}
