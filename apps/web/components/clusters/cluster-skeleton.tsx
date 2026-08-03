import { CardSkeleton, TableSkeleton } from "@/components/common";

import type { ClusterView } from "./types";

export function ClusterSkeleton({ view }: { view: ClusterView }) {
  return view === "table" ? (
    <TableSkeleton />
  ) : (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <CardSkeleton />
      <CardSkeleton />
      <CardSkeleton />
    </div>
  );
}
