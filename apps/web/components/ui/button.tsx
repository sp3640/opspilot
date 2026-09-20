import * as React from "react";
import { cn } from "@/lib/utils";

type ButtonVariant =
  | "primary"
  | "secondary"
  | "outline"
  | "ghost"
  | "danger"
  | "success";

type ButtonSize = "sm" | "md" | "icon";

interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
}

// One fixed height per size so buttons line up with inputs/selects (h-10)
// instead of growing with padding and overlapping neighbours.
const sizes: Record<ButtonSize, string> = {
  sm: "h-8 px-3 text-xs",
  md: "h-10 px-4 text-sm",
  icon: "h-10 w-10 p-0 text-sm",
};

// Variants are pure classes (no inline style) so hover/focus/disabled
// states and caller className overrides all work. Colors come only from the
// centralized tokens in globals.css.
const variants: Record<ButtonVariant, string> = {
  primary: `
    border border-transparent
    bg-[var(--primary)] text-[color:var(--primary-foreground)]
    shadow-[var(--shadow-sm)]
    hover:brightness-110
    active:scale-[0.98]
  `,

  secondary: `
    border border-[color:var(--border)]
    bg-[var(--card)] text-[color:var(--foreground)]
    hover:bg-[var(--card-hover)]
    active:scale-[0.98]
  `,

  outline: `
    border border-[color:var(--border)]
    bg-transparent text-[color:var(--foreground)]
    hover:bg-[var(--card-hover)]
    active:scale-[0.98]
  `,

  ghost: `
    border border-transparent
    bg-transparent text-[color:var(--foreground)]
    hover:bg-[var(--card-hover)]
    active:scale-[0.98]
  `,

  // Semantic actions use the same dark tint + subtle border + readable text
  // pattern as badges and cards instead of a loud solid fill.
  danger: `
    border border-[color:var(--danger-border)]
    bg-[var(--danger-bg)] text-[color:var(--danger-text)]
    hover:bg-[color-mix(in_srgb,var(--danger)_16%,var(--danger-bg))]
    active:scale-[0.98]
  `,

  success: `
    border border-[color:var(--success-border)]
    bg-[var(--success-bg)] text-[color:var(--success-text)]
    hover:bg-[color-mix(in_srgb,var(--success)_16%,var(--success-bg))]
    active:scale-[0.98]
  `,
};

export function Button({
  className,
  variant = "primary",
  size = "md",
  loading = false,
  children,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <button
      disabled={disabled || loading}
      aria-busy={loading}
      className={cn(
        `
        inline-flex
        items-center
        justify-center
        gap-2

        shrink-0
        whitespace-nowrap
        rounded-xl

        font-semibold

        transition-all
        duration-300

        focus:outline-none
        focus:ring-2
        focus:ring-[var(--focus-ring)]/40

        disabled:pointer-events-none
        disabled:opacity-50
        disabled:cursor-not-allowed
        `,
        sizes[size],
        variants[variant],
        className
      )}
      {...props}
    >
      {loading && (
        <span
          aria-hidden="true"
          className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent opacity-70"
        />
      )}

      {loading && <span className="sr-only">Loading</span>}

      {children}
    </button>
  );
}
