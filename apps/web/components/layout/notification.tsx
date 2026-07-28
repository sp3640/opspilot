import { Bell } from "lucide-react";

export function NotificationButton() {
  return (
    <button
      className="
        relative
        flex
        h-11
        w-11
        items-center
        justify-center
        rounded-2xl
        border
        transition-all
        duration-300
        hover:-translate-y-0.5
        hover:border-blue-500/30
      "
      style={{
        background: "rgba(255,255,255,0.03)",
        borderColor: "var(--border)",
      }}
    >
      <Bell className="h-5 w-5" />

      <span className="absolute right-3 top-3 h-2 w-2 rounded-full bg-red-500" />
    </button>
  );
}