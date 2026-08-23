"use client";

import { useMemo } from "react";
import { X } from "lucide-react";

import { cn } from "@/lib/utils";
import { useOrganizations } from "@/hooks/use-organizations";
import { navigationGroups } from "@/constants/navigation";
import { Logo } from "./logo";
import { NavItem } from "./nav-item";
import { SidebarGroup } from "./sidebar-group";

const organizationListParams = {
  page: 1,
  limit: 1,
  sort: "updated_at" as const,
  order: "desc" as const,
};

type SidebarProps = {
  /** Whether the mobile off-canvas drawer is open. Ignored at the `lg:`
   * breakpoint and above, where the sidebar is always shown inline. */
  mobileOpen?: boolean;
  onMobileClose?: () => void;
};

export function Sidebar({ mobileOpen = false, onMobileClose }: SidebarProps) {
  const { data } = useOrganizations(organizationListParams);

  const organizationName = useMemo(() => {
    return data?.items?.[0]?.name ?? "Workspace";
  }, [data]);

  return (
    <>
      {mobileOpen ? (
        <div
          className="fixed inset-0 z-40 bg-black/60 lg:hidden"
          aria-hidden="true"
          onClick={onMobileClose}
        />
      ) : null}

      <aside
        className={cn(
          `
          fixed
          inset-y-0
          left-0
          z-50
          flex
          h-screen
          w-72
          flex-col

          border-r

          px-4
          py-6

          transition-transform
          duration-200

          lg:sticky
          lg:top-0
          lg:translate-x-0
          `,
          mobileOpen ? "translate-x-0" : "-translate-x-full"
        )}
        style={{
          background: "#0d1117",
          borderColor: "var(--border)",
        }}
      >
        <div className="flex items-center justify-between">
          <Logo />

          <button
            type="button"
            onClick={onMobileClose}
            aria-label="Close navigation menu"
            className="rounded-lg p-1.5 text-zinc-400 hover:text-white lg:hidden"
          >
            <X aria-hidden="true" className="h-5 w-5" />
          </button>
        </div>

        <nav className="mt-8 flex-1 space-y-8 overflow-y-auto">
        {navigationGroups.map((group) => (
          <SidebarGroup
            key={group.title}
            title={group.title}
          >
            {group.items.map((item) => (
              <NavItem
                key={item.href}
                {...item}
              />
            ))}
          </SidebarGroup>
        ))}
      </nav>

      <div
        className="
        rounded-2xl
        border
        p-4
        "
        style={{
          borderColor: "var(--border)",
          background: "rgba(255,255,255,0.03)",
        }}
      >
        <div className="flex items-center gap-2">
          <div className="h-2.5 w-2.5 rounded-full bg-emerald-500" />

          <span className="text-sm font-medium">
            Connected
          </span>
        </div>

        <p className="mt-2 text-xs text-zinc-400">Workspace: {organizationName}</p>
      </div>
      </aside>
    </>
  );
}