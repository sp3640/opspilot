import { CardSkeleton, TableSkeleton } from "@/components/common";
import type { ResourceView } from "./types";
export function ResourceSkeleton({ view }: { view: ResourceView }) { return view === "table" ? <TableSkeleton /> : <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3"><CardSkeleton /><CardSkeleton /><CardSkeleton /></div>; }
