import { CheckCircle2, CircleAlert, type LucideIcon } from "lucide-react";

type HealthStatus = "healthy" | "degraded";

type HealthItemProps = {
  name: string;
  status: HealthStatus;
  responseTime: string;
  health: number;
  icon: LucideIcon;
};

/** A single measured dependency status used by the system-health list. */
export function HealthItem({ name, status, responseTime, health, icon: Icon }: HealthItemProps) {
  const isHealthy = status === "healthy";
  const statusColor = isHealthy ? "var(--success)" : "var(--warning)";
  const StatusIcon = isHealthy ? CheckCircle2 : CircleAlert;

  return (
    <li className="grid gap-3 rounded-2xl border p-4 transition-colors sm:grid-cols-[minmax(150px,1fr)_minmax(160px,1.25fr)_auto_auto] sm:items-center" style={{ borderColor: "var(--border)", backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)" }}>
      <div className="flex items-center gap-3"><span className="rounded-xl p-2" style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 10%, transparent)" }}><Icon aria-hidden="true" className="h-4 w-4" /></span><span className="font-medium">{name}</span></div>
      <div className="flex items-center gap-3"><div className="h-1.5 min-w-20 flex-1 overflow-hidden rounded-full" style={{ backgroundColor: "var(--muted)" }}><div className="h-full rounded-full" style={{ width: `${health}%`, backgroundColor: statusColor }} /></div><span className="w-9 text-right text-xs tabular-nums" style={{ color: "var(--muted-foreground)" }}>{health}%</span></div>
      <span className="text-sm tabular-nums" style={{ color: "var(--muted-foreground)" }}>{responseTime}</span>
      <span className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold capitalize" style={{ color: statusColor, backgroundColor: `color-mix(in srgb, ${statusColor} 12%, transparent)` }}><StatusIcon aria-hidden="true" className="h-3.5 w-3.5" />{status}</span>
    </li>
  );
}
