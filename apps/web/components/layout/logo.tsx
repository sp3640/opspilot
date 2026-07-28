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
        <ShieldCheck className="h-6 w-6 text-white" />
      </div>

      <div>
        <h1 className="text-xl font-bold tracking-tight">
          OpsPilot
        </h1>

        <p className="text-xs uppercase tracking-[0.18em] text-zinc-500">
          Enterprise Platform
        </p>
      </div>
    </div>
  );
}