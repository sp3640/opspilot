import { Search } from "lucide-react";

export function SearchBar() {
  return (
    <div
      className="
        flex
        h-11
        w-full
        max-w-md
        items-center
        gap-3
        rounded-2xl
        border
        px-4
        transition-all
        duration-300
        hover:border-blue-500/30
      "
      style={{
        background: "rgba(255,255,255,0.03)",
        borderColor: "var(--border)",
      }}
    >
      <Search className="h-4 w-4 text-zinc-500" />

      <input
        placeholder="Search projects, incidents..."
        className="
          flex-1
          bg-transparent
          text-sm
          outline-none
          placeholder:text-zinc-500
        "
      />

      <kbd
        className="
          rounded-lg
          border
          px-2
          py-1
          text-[11px]
          text-zinc-400
        "
        style={{ borderColor: "var(--border)" }}
      >
        ⌘K
      </kbd>
    </div>
  );
}