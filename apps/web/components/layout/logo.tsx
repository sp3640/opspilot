import { ShieldCheck } from "lucide-react";

export function Logo() {
  return (
    <div className="flex items-center gap-4 px-2">
      <div
        className="
        flex
        h-12
        w-12
        items-center
        justify-center

        rounded-2xl
        "
        style={{
          background: "var(--primary)",
        }}
      >
        <ShieldCheck className="h-6 w-6 text-[var(--primary-foreground)]" />
      </div>

      <div className="min-w-0">
        <h1 className="truncate text-xl font-bold tracking-tight" style={{ color: "var(--surface-foreground)" }}>
          OpsPilot
        </h1>

        <p className="truncate text-xs uppercase tracking-[0.18em]" style={{ color: "var(--subtle-foreground)" }}>
          Enterprise Platform
        </p>
      </div>
    </div>
  );
}