import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";

type EmptyStateProps = {
  icon: LucideIcon;
  title: string;
  description: string;
  action?: ReactNode;
};

/** A calm zero-data state for lists, search, and content areas. */
export function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps) {
  return (
    <section className="flex min-h-72 flex-col items-center justify-center rounded-3xl border px-6 py-12 text-center shadow-[var(--shadow-sm)]" style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}>
      <div className="rounded-2xl p-4" style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}><Icon aria-hidden="true" className="h-7 w-7" /></div>
      <h2 className="mt-5 text-lg font-semibold tracking-tight">{title}</h2>
      <p className="mt-2 max-w-md text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>{description}</p>
      {action && <div className="mt-6">{action}</div>}
    </section>
  );
}
