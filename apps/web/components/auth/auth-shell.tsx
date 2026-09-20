import type { ReactNode } from "react";
import { Activity, GitBranch, ShieldCheck } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { Logo } from "@/components/layout/logo";

type Capability = {
  icon: LucideIcon;
  title: string;
  description: string;
};

const capabilities: Capability[] = [
  {
    icon: Activity,
    title: "Real-time incident response",
    description: "Track, assign, and resolve incidents with full operational context.",
  },
  {
    icon: ShieldCheck,
    title: "Infrastructure visibility",
    description: "Monitor clusters, resources, and deployments from one control plane.",
  },
  {
    icon: GitBranch,
    title: "Deployment tracking",
    description: "See what shipped, where it's running, and its current health.",
  },
];

/**
 * Shared shell for every authentication page (login, register, accept-
 * invitation) - previously each page defined its own near-identical copy.
 * A restrained branding panel (no fake stats/testimonials/logos - just a
 * plain, accurate description of what the product does) sits beside the
 * form on wide screens and collapses to just the logo above the form on
 * narrower ones, so the actual form composition is unaffected.
 */
export function AuthShell({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen" style={{ backgroundColor: "var(--background)", color: "var(--foreground)" }}>
      <div
        className="relative hidden w-[42%] max-w-xl shrink-0 flex-col justify-between overflow-hidden border-r p-10 lg:flex xl:p-14"
        style={{ backgroundColor: "var(--surface)", color: "var(--surface-foreground)", borderColor: "var(--border)" }}
      >
        <Logo />

        <div className="space-y-8">
          <div>
            <h2 className="text-2xl font-semibold tracking-tight">Operations, unified.</h2>
            <p className="mt-3 max-w-sm text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>
              OpsPilot brings incidents, deployments, infrastructure, and your team into a single operations control
              plane.
            </p>
          </div>

          <ul className="space-y-5">
            {capabilities.map(({ icon: Icon, title, description }) => (
              <li key={title} className="flex gap-3">
                <span
                  className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl"
                  style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 15%, transparent)" }}
                >
                  <Icon aria-hidden="true" className="h-4 w-4" />
                </span>
                <div className="min-w-0">
                  <p className="text-sm font-medium">{title}</p>
                  <p className="mt-0.5 text-xs leading-5" style={{ color: "var(--muted-foreground)" }}>
                    {description}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        </div>

        <p className="text-xs uppercase tracking-[0.18em]" style={{ color: "var(--subtle-foreground)" }}>
          OpsPilot Enterprise Platform
        </p>
      </div>

      <div className="flex flex-1 flex-col items-center justify-center px-4 py-12 sm:px-6">
        <div className="mb-8 lg:hidden">
          <Logo />
        </div>

        <div
          className="w-full max-w-md rounded-[var(--radius)] border p-6 sm:p-8"
          style={{ backgroundColor: "var(--card)", borderColor: "var(--border)", boxShadow: "var(--shadow-lg)" }}
        >
          {children}
        </div>
      </div>
    </div>
  );
}
