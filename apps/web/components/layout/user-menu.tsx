import { ChevronDown } from "lucide-react";

export function UserMenu() {
  return (
    <button
      className="
        flex
        items-center
        gap-3
        rounded-2xl
        border
        px-3
        py-2
        transition-all
        duration-300
        hover:border-blue-500/30
      "
      style={{
        background: "rgba(255,255,255,0.03)",
        borderColor: "var(--border)",
      }}
    >
      <div
        className="
          flex
          h-10
          w-10
          items-center
          justify-center
          rounded-xl
          font-semibold
          text-white
        "
        style={{
          background: "var(--primary)",
        }}
      >
        SR
      </div>

      <div className="text-left">
        <p className="text-sm font-semibold">
          Siddharth
        </p>

        <p className="text-xs text-zinc-400">
          Platform Admin
        </p>
      </div>

      <ChevronDown className="h-4 w-4 text-zinc-500" />
    </button>
  );
}