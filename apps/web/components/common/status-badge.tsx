import { AlertTriangle, Archive, CheckCircle2, CircleAlert, Info } from "lucide-react";

import { cn } from "@/lib/utils";

type StatusVariant = "healthy" | "warning" | "critical" | "archived" | "success" | "info";

type StatusBadgeProps = {
  variant: StatusVariant;
  children?: string;
  className?: string;
};

const statusConfig = {
  healthy: { label: "Healthy", tone: "success", icon: CheckCircle2 },
  warning: { label: "Warning", tone: "warning", icon: AlertTriangle },
  critical: { label: "Critical", tone: "danger", icon: CircleAlert },
  archived: { label: "Archived", tone: "neutral", icon: Archive },
  success: { label: "Success", tone: "success", icon: CheckCircle2 },
  info: { label: "Info", tone: "info", icon: Info },
} as const;

/** Every tone resolves to the centralized semantic tokens in globals.css. */
const toneStyles = {
  success: { color: "var(--success-text)", backgroundColor: "var(--success-bg)", borderColor: "var(--success-border)" },
  warning: { color: "var(--warning-text)", backgroundColor: "var(--warning-bg)", borderColor: "var(--warning-border)" },
  danger: { color: "var(--danger-text)", backgroundColor: "var(--danger-bg)", borderColor: "var(--danger-border)" },
  info: { color: "var(--info-text)", backgroundColor: "var(--info-bg)", borderColor: "var(--info-border)" },
  neutral: { color: "var(--muted-foreground)", backgroundColor: "var(--card-hover)", borderColor: "var(--border)" },
} as const;

/** Semantic status marker that keeps status colors consistent across modules. */
export function StatusBadge({ variant, children, className }: StatusBadgeProps) {
  const { label, tone, icon: Icon } = statusConfig[variant];
  const { borderColor, ...toneColors } = toneStyles[tone];

  return (
    <span className={cn("inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold", className)} style={{ ...toneColors, boxShadow: `inset 0 0 0 1px ${borderColor}` }}>
      <Icon aria-hidden="true" className="h-3.5 w-3.5" />
      {children ?? label}
    </span>
  );
}
