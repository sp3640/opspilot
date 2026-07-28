import { cn } from "@/lib/utils";

type SkeletonProps = { className?: string };

/** Base shimmer primitive for predictable content-loading layouts. */
export function Skeleton({ className }: SkeletonProps) {
  return <div aria-hidden="true" className={cn("animate-pulse rounded-xl", className)} style={{ backgroundColor: "var(--muted)" }} />;
}

export function MetricSkeleton() {
  return (
    <div className="rounded-2xl border p-5 shadow-[var(--shadow-sm)]" style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}>
      <Skeleton className="h-3 w-20" />
      <Skeleton className="mt-4 h-9 w-16" />
      <Skeleton className="mt-5 h-4 w-32" />
    </div>
  );
}

export function CardSkeleton() {
  return (
    <div className="rounded-3xl border p-6 shadow-[var(--shadow-sm)]" style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}>
      <Skeleton className="h-5 w-36" />
      <Skeleton className="mt-3 h-4 w-full" />
      <Skeleton className="mt-2 h-4 w-4/5" />
      <Skeleton className="mt-7 h-28 w-full" />
    </div>
  );
}

export function ListSkeleton({ rows = 4 }: { rows?: number }) {
  return (
    <div className="space-y-3" aria-label="Loading content" role="status">
      {Array.from({ length: rows }, (_, index) => (
        <div key={index} className="flex items-center gap-3 rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
          <Skeleton className="h-9 w-9 shrink-0 rounded-xl" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-1/3" />
            <Skeleton className="h-3 w-2/3" />
          </div>
          <Skeleton className="h-4 w-12" />
        </div>
      ))}
      <span className="sr-only">Loading content</span>
    </div>
  );
}

export function TableSkeleton({ columns = 4, rows = 5 }: { columns?: number; rows?: number }) {
  return (
    <div className="overflow-hidden rounded-3xl border" role="status" aria-label="Loading table" style={{ borderColor: "var(--border)" }}>
      <div className="grid gap-4 border-b p-4" style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`, borderColor: "var(--border)" }}>
        {Array.from({ length: columns }, (_, index) => <Skeleton key={index} className="h-3 w-3/4" />)}
      </div>
      {Array.from({ length: rows }, (_, row) => (
        <div key={row} className="grid gap-4 border-b p-4 last:border-b-0" style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`, borderColor: "var(--border)" }}>
          {Array.from({ length: columns }, (_, column) => <Skeleton key={column} className="h-4 w-full" />)}
        </div>
      ))}
      <span className="sr-only">Loading table</span>
    </div>
  );
}
