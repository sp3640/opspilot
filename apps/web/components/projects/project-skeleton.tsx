import { CardSkeleton, TableSkeleton } from "@/components/common";

import type { ProjectView } from "./types";

/** Layout-aware loading state for future API-backed project data. */
export function ProjectSkeleton({ view }: { view: ProjectView }) {
  return view === "table" ? <TableSkeleton /> : <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3"><CardSkeleton /><CardSkeleton /><CardSkeleton /></div>;
}
