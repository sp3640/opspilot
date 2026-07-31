import * as React from "react";
import { cn } from "@/lib/utils";

type ButtonVariant =
  | "primary"
  | "secondary"
  | "outline"
  | "ghost"
  | "danger";

interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  loading?: boolean;
}

const variants: Record<ButtonVariant, string> = {
  primary: `
    text-white
    border
    hover:brightness-110
    active:scale-[0.98]
  `,

  secondary: `
    border
    hover:bg-white/5
    active:scale-[0.98]
  `,

  outline: `
    border
    bg-transparent
    hover:bg-white/5
    active:scale-[0.98]
  `,

  ghost: `
    border border-transparent
    bg-transparent
    hover:bg-white/5
    active:scale-[0.98]
  `,

  danger: `
    text-white
    border border-red-500/30
    bg-red-600
    hover:bg-red-500
    active:scale-[0.98]
  `,
};

export function Button({
  className,
  variant = "primary",
  loading = false,
  children,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <button
      disabled={disabled || loading}
      aria-busy={loading}
      style={
        variant === "primary"
          ? {
              background: "var(--primary)",
              borderColor: "transparent",
              boxShadow: "var(--shadow-sm)",
            }
          : variant === "secondary"
          ? {
              background: "var(--card)",
              borderColor: "var(--border)",
              color: "var(--foreground)",
            }
          : variant === "outline"
          ? {
              borderColor: "var(--border)",
              color: "var(--foreground)",
            }
          : variant === "ghost"
          ? {
              color: "var(--foreground)",
            }
          : undefined
      }
      className={cn(
        `
        inline-flex
        items-center
        justify-center
        gap-2

        rounded-2xl

        px-5
        py-2.5

        text-sm
        font-semibold

        transition-all
        duration-300

        focus:outline-none
        focus:ring-2
        focus:ring-[var(--primary)]/40

        disabled:pointer-events-none
        disabled:opacity-50
        disabled:cursor-not-allowed
        `,
        variants[variant],
        className
      )}
      {...props}
    >
      {loading && (
        <span
          aria-hidden="true"
          className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"
        />
      )}

      {loading && <span className="sr-only">Loading</span>}

      {children}
    </button>
  );
}
