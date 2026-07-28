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
        py-3

        transition-all
        duration-200
        `,
        active
          ? "text-white"
          : "text-zinc-400 hover:text-white"
      )}
      style={{
        background: active
          ? "rgba(79,140,255,.15)"
          : "transparent",
      }}
    >
      {active && (
        <div className="absolute left-0 top-2 bottom-2 w-1 rounded-full bg-[var(--primary)]" />
      )}

      <Icon
        className={cn(
          "h-5 w-5 transition-all",
          active
            ? "text-[var(--primary)]"
            : "group-hover:text-white"
        )}
      />

      <span className="font-medium">
        {title}
      </span>
    </Link>
  );
}
