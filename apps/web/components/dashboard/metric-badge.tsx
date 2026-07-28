import { ArrowDownRight, ArrowUpRight, Minus } from "lucide-react";

type MetricBadgeProps = {
  value: string;
  direction?: "up" | "down" | "neutral";
  tone?: "success" | "warning" | "neutral";
};

const toneStyles = {
  success: { color: "var(--success)", backgroundColor: "color-mix(in srgb, var(--success) 12%, transparent)" },
  warning: { color: "var(--warning)", backgroundColor: "color-mix(in srgb, var(--warning) 12%, transparent)" },
  neutral: { color: "var(--muted-foreground)", backgroundColor: "var(--muted)" },
};

/** A compact, semantic trend indicator used in metric summaries. */
export function MetricBadge({ value, direction = "up", tone = "success" }: MetricBadgeProps) {
  const Icon = direction === "up" ? ArrowUpRight : direction === "down" ? ArrowDownRight : Minus;

  return (
    <span
      className="inline-flex items-center gap-1 rounded-full px-2 py-1 text-xs font-semibold"
      style={toneStyles[tone]}
    >
      <Icon aria-hidden="true" className="h-3.5 w-3.5" />
      {value}
    </span>
  );
}
