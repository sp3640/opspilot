"use client";

import { TriangleAlert } from "lucide-react";

import { Button } from "@/components/ui/button";

type ErrorStateProps = {
  title?: string;
  description: string;
  onRetry?: () => void;
  retryLabel?: string;
};

/** A recoverable error surface with a token-aligned diagnostic illustration. */
export function ErrorState({ title = "Unable to load this content", description, onRetry, retryLabel = "Try again" }: ErrorStateProps) {
  return (
    <section role="alert" className="flex min-h-72 flex-col items-center justify-center rounded-3xl border px-6 py-12 text-center shadow-[var(--shadow-sm)]" style={{ backgroundColor: "var(--card)", borderColor: "color-mix(in srgb, var(--danger) 30%, var(--border))" }}>
      <div className="rounded-2xl border p-4" style={{ color: "var(--danger)", backgroundColor: "color-mix(in srgb, var(--danger) 10%, transparent)", borderColor: "color-mix(in srgb, var(--danger) 25%, var(--border))" }}><TriangleAlert aria-hidden="true" className="h-7 w-7" /></div>
      <h2 className="mt-5 text-lg font-semibold tracking-tight">{title}</h2>
      <p className="mt-2 max-w-md text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>{description}</p>
      <Button type="button" onClick={onRetry} className="mt-6">{retryLabel}</Button>
    </section>
  );
}
