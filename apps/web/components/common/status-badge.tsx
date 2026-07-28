import { AlertTriangle, Archive, CheckCircle2, CircleAlert, Info } from "lucide-react";

import { cn } from "@/lib/utils";

type StatusVariant = "healthy" | "warning" | "critical" | "archived" | "success" | "info";

type StatusBadgeProps = {
  variant: StatusVariant;
  children?: string;
  className?: string;
};

const statusConfig = {
  healthy: { label: "Healthy", color: "var(--success)", icon: CheckCircle2 },
  warning: { label: "Warning", color: "var(--warning)", icon: AlertTriangle },
  critical: { label: "Critical", color: "var(--danger)", icon: CircleAlert },
  archived: { label: "Archived", color: "var(--muted-foreground)", icon: Archive },
  success: { label: "Success", color: "var(--success)", icon: CheckCircle2 },
  info: { label: "Info", color: "var(--primary)", icon: Info },
} as const;

/** Semantic status marker that keeps status colors consistent across modules. */
export function StatusBadge({ variant, children, className }: StatusBadgeProps) {
  const { label, color, icon: Icon } = statusConfig[variant];

  return (
    <span className={cn("inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold", className)} style={{ color, backgroundColor: `color-mix(in srgb, ${color} 12%, transparent)` }}>
      <Icon aria-hidden="true" className="h-3.5 w-3.5" />
      {children ?? label}
    </span>
  );
}
