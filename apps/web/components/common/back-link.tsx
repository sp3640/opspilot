"use client";

import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";

type BackLinkProps = {
  /** Where to go when there's no usable in-app history to go back to. */
  fallbackHref: string;
  label?: string;
};

/**
 * Prefers router.back() when the user actually navigated here from within
 * the app (so it returns them to wherever they scrolled to/filtered by on
 * the parent list); falls back to a fixed route otherwise (direct load, a
 * bookmarked/shared link, or history exhausted) so it never navigates the
 * user out of the application or to a blank history entry.
 */
export function BackLink({ fallbackHref, label = "Back" }: BackLinkProps) {
  const router = useRouter();

  const handleClick = () => {
    const canGoBack =
      typeof window !== "undefined" &&
      window.history.length > 1 &&
      document.referrer.startsWith(window.location.origin);

    if (canGoBack) {
      router.back();
    } else {
      router.push(fallbackHref);
    }
  };

  return (
    <button
      type="button"
      onClick={handleClick}
      className="inline-flex items-center gap-1.5 text-sm font-medium transition-colors hover:text-[var(--foreground)]"
      style={{ color: "var(--muted-foreground)" }}
    >
      <ArrowLeft aria-hidden="true" className="h-4 w-4" />
      {label}
    </button>
  );
}
