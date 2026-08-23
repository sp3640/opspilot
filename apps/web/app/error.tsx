"use client";

import { useEffect } from "react";
import { TriangleAlert } from "lucide-react";

import { Button } from "@/components/ui/button";

/**
 * Route-level error boundary (Next.js App Router convention). Without this
 * file, an unhandled throw in any Server or Client Component under this
 * layout unmounts the React tree with no recovery UI - the user is left
 * looking at a blank screen. This catches that, offers a retry (Next.js
 * re-renders the segment on reset()) and a way back to a known-good page,
 * and logs the error rather than silently swallowing it.
 */
export default function Error({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  useEffect(() => {
    console.error("Unhandled application error:", error);
  }, [error]);

  return (
    <div
      className="flex min-h-screen flex-col items-center justify-center px-4 py-12 text-center"
      style={{ backgroundColor: "var(--background)", color: "var(--foreground)" }}
    >
      <div
        className="rounded-2xl border p-4"
        style={{
          color: "var(--danger)",
          backgroundColor: "color-mix(in srgb, var(--danger) 10%, transparent)",
          borderColor: "color-mix(in srgb, var(--danger) 25%, var(--border))",
        }}
      >
        <TriangleAlert aria-hidden="true" className="h-7 w-7" />
      </div>
      <h1 className="mt-5 text-xl font-semibold tracking-tight">Something went wrong</h1>
      <p className="mt-2 max-w-md text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>
        An unexpected error occurred while rendering this page. You can try again, or head back to the dashboard.
      </p>
      <div className="mt-6 flex items-center gap-3">
        <Button type="button" variant="secondary" onClick={() => window.location.assign("/")}>
          Back to dashboard
        </Button>
        <Button type="button" onClick={() => reset()}>
          Try again
        </Button>
      </div>
    </div>
  );
}
