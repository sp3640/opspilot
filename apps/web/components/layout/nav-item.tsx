"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

type NavItemProps = {
  href: string;
  title: string;
  icon: LucideIcon;
};

export function NavItem({
  href,
  title,
  icon: Icon,
}: NavItemProps) {
  const pathname = usePathname();

  const active = pathname === href;

  return (
    <Link
      href={href}
      className={cn(
        `
        group
        relative
        flex
        items-center
        gap-3

        rounded-2xl

        px-4
        py-2.5

        transition-all
        duration-200
        `,
        active ? "" : "hover:text-[var(--surface-foreground)]"
      )}
      style={{
        background: active ? "color-mix(in srgb, var(--primary) 15%, transparent)" : "transparent",
        color: active ? "var(--surface-foreground)" : "var(--muted-foreground)",
      }}
    >
      {active && (
        <div className="absolute left-0 top-2 bottom-2 w-1 rounded-full bg-[var(--primary)]" />
      )}

      <Icon
        className="h-5 w-5 shrink-0 transition-all"
        style={{ color: active ? "var(--primary)" : "currentColor" }}
      />

      <span className="truncate font-medium">
        {title}
      </span>
    </Link>
  );
}
