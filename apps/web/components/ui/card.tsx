import { ReactNode } from "react";
import { cn } from "@/lib/utils";

interface CardProps {
  children: ReactNode;
  className?: string;
}

export function Card({ children, className }: CardProps) {
  return (
    <div
      className={cn(
        `
        rounded-2xl
        border
        p-6

        transition-all
        duration-300

        hover:-translate-y-1
        hover:-translate-y-1

        `,
        className
      )}
      style={{
        background: "var(--card)",
        color: "var(--card-foreground)",
        borderColor: "var(--border)",
        boxShadow: "var(--shadow-md)",
      }}
    >
      {children}
    </div>
  );
}
