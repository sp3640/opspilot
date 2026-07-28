"use client";

import { navigationGroups } from "@/constants/navigation";
import { Logo } from "./logo";
import { NavItem } from "./nav-item";
import { SidebarGroup } from "./sidebar-group";

export function Sidebar() {
  return (
    <aside
      className="
      flex
      h-screen
      w-72
      flex-col

      border-r

      px-4
      py-6
      "
      style={{
        background: "#0d1117",
        borderColor: "var(--border)",
      }}
    >
      <Logo />

      <nav className="mt-8 flex-1 space-y-8">
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

        <p className="mt-2 text-xs text-zinc-400">
          Workspace: Enterprise
        </p>
      </div>
    </aside>
  );
}