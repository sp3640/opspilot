import { ChevronDown } from "lucide-react";
import { useMemo } from "react";

import { useOrganizations } from "@/hooks/use-organizations";
import { useAuthStore } from "@/store/auth-store";

const organizationListParams = {
  page: 1,
  limit: 1,
  sort: "updated_at" as const,
  order: "desc" as const,
};

export function UserMenu() {
  const currentUser = useAuthStore((state) => state.currentUser);
  const { data } = useOrganizations(organizationListParams);

  const name = currentUser?.name?.trim() || "User";
  const email = currentUser?.email?.trim() || "";
  const role = currentUser?.role?.trim() || "Viewer";
  const organizationName = useMemo(() => {
    return data?.items?.[0]?.name?.trim() || "No organization";
  }, [data]);
  const initials = name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((token) => token[0]?.toUpperCase() ?? "")
    .join("") || "U";

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
        {initials}
      </div>

      <div className="text-left">
        <p className="text-sm font-semibold">{name}</p>

        <p className="text-xs text-zinc-400">{email}</p>

        <p className="text-xs text-zinc-500">{organizationName}</p>

        <p className="text-xs text-zinc-500">{role}</p>
      </div>

      <ChevronDown className="h-4 w-4 text-zinc-500" />
    </button>
  );
}