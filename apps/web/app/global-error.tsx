"use client";

import { useEffect } from "react";

import "./globals.css";

/**
 * Last-resort error boundary for a throw in the root layout itself (where
 * app/error.tsx cannot help, since it renders inside that same layout).
 * Next.js requires this file to render its own <html>/<body> - it replaces
 * the entire tree, not just a segment, when it's the one that catches.
 */
export default function GlobalError({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  useEffect(() => {
    console.error("Unhandled application error (root layout):", error);
  }, [error]);

  return (
    <html lang="en">
      <body>
        <div
          style={{
            display: "flex",
            minHeight: "100vh",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            padding: "3rem 1rem",
            textAlign: "center",
            backgroundColor: "var(--background, #0b0b0d)",
            color: "var(--foreground, #f4f4f5)",
          }}
        >
          <h1 style={{ fontSize: "1.25rem", fontWeight: 600 }}>Something went wrong</h1>
          <p style={{ marginTop: "0.5rem", maxWidth: "28rem", fontSize: "0.875rem", opacity: 0.75 }}>
            An unexpected error occurred and the page couldn&apos;t be displayed. Reloading usually resolves this.
          </p>
          <button
            type="button"
            onClick={() => reset()}
            style={{
              marginTop: "1.5rem",
              borderRadius: "0.75rem",
              border: "1px solid currentColor",
              padding: "0.5rem 1rem",
              fontSize: "0.875rem",
              cursor: "pointer",
              background: "transparent",
              color: "inherit",
            }}
          >
            Try again
          </button>
        </div>
      </body>
    </html>
  );
}
