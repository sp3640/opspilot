"use client";

import { useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ChevronDown, LogOut, Settings } from "lucide-react";

import { useOrganizations } from "@/hooks/use-organizations";
import { useAuthStore } from "@/store/auth-store";
import { useDismissable } from "@/hooks/use-dismissable";

const organizationListParams = {
  page: 1,
  limit: 1,
  sort: "updated_at" as const,
  order: "desc" as const,
};

export function UserMenu() {
  const router = useRouter();
  const containerRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const currentUser = useAuthStore((state) => state.currentUser);
  const logout = useAuthStore((state) => state.logout);
  const { data } = useOrganizations(organizationListParams);

  const name = currentUser?.name?.trim() || "User";
  const email = currentUser?.email?.trim() || "";
  const role = currentUser?.role?.trim() || "Viewer";
  const organizationName = useMemo(() => {
    return data?.items?.[0]?.name?.trim() || "No organization";
  }, [data]);
  const initials =
    name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((token) => token[0]?.toUpperCase() ?? "")
      .join("") || "U";

  useDismissable(containerRef, open, () => setOpen(false));

  const handleLogout = () => {
    setOpen(false);
    // Clears the persisted access token (localStorage) and in-memory auth
    // state (accessToken + currentUser). AuthGuard reacts to that
    // immediately (it returns null the moment isAuthenticated goes false),
    // and this explicit redirect guarantees the login page is reached even
    // for anything not wrapped in AuthGuard - no server-side session to
    // revoke, since auth is stateless JWT (see internal/middleware/auth.go).
    logout();
    router.push("/login");
  };

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        aria-haspopup="true"
        aria-expanded={open}
        aria-label={`Account menu for ${name}`}
        className="
          flex
          max-w-64
          items-center
          gap-3
          rounded-2xl
          border
          px-3
          py-2
          transition-all
          duration-300
          hover:border-[var(--primary-hover)]/40
        "
        style={{
          background: "var(--surface-hover)",
          borderColor: "var(--border)",
          color: "var(--surface-foreground)",
        }}
      >
        <div
          className="
            flex
            h-10
            w-10
            shrink-0
            items-center
            justify-center
            rounded-xl
            font-semibold
            text-[var(--primary-foreground)]
          "
          style={{
            background: "var(--primary)",
          }}
        >
          {initials}
        </div>

        <div className="min-w-0 flex-1 text-left">
          <p className="truncate text-sm font-semibold">{name}</p>

          <p className="truncate text-xs" style={{ color: "var(--muted-foreground)" }}>
            {email}
          </p>

          <p className="truncate text-xs" style={{ color: "var(--subtle-foreground)" }}>
            {organizationName}
          </p>

          <p className="truncate text-xs" style={{ color: "var(--subtle-foreground)" }}>
            {role}
          </p>
        </div>

        <ChevronDown aria-hidden="true" className="h-4 w-4 shrink-0 transition-transform" style={{ color: "var(--subtle-foreground)", transform: open ? "rotate(180deg)" : undefined }} />
      </button>

      {open && (
        <div
          role="menu"
          aria-label="Account menu"
          className="absolute right-0 z-40 mt-2 w-64 overflow-hidden rounded-2xl border"
          style={{ backgroundColor: "var(--card)", borderColor: "var(--border)", boxShadow: "var(--shadow-lg)", color: "var(--card-foreground)" }}
        >
          <div className="border-b px-4 py-3" style={{ borderColor: "var(--border)" }}>
            <p className="truncate text-sm font-semibold">{name}</p>
            {email && (
              <p className="mt-0.5 truncate text-xs" style={{ color: "var(--muted-foreground)" }}>
                {email}
              </p>
            )}
            <p className="mt-0.5 truncate text-xs" style={{ color: "var(--subtle-foreground)" }}>
              {role}
            </p>
          </div>

          <div className="p-1.5">
            <Link
              href="/settings"
              role="menuitem"
              onClick={() => setOpen(false)}
              className="flex items-center gap-2.5 rounded-xl px-2.5 py-2 text-sm transition-colors hover:bg-[var(--muted)]"
            >
              <Settings aria-hidden="true" className="h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
              Settings
            </Link>
          </div>

          <div className="border-t p-1.5" style={{ borderColor: "var(--border)" }}>
            <button
              type="button"
              role="menuitem"
              onClick={handleLogout}
              className="flex w-full items-center gap-2.5 rounded-xl px-2.5 py-2 text-left text-sm transition-colors hover:bg-[var(--muted)]"
              style={{ color: "var(--danger-text)" }}
            >
              <LogOut aria-hidden="true" className="h-4 w-4" />
              Log out
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
