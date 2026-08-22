import { CardSkeleton, TableSkeleton } from "@/components/common";

import type { TeamView } from "./types";

export function TeamSkeleton({ view }: { view: TeamView }) {
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
