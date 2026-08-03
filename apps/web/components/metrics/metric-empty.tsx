import { Activity } from "lucide-react";
import { EmptyState } from "@/components/common";
export function MetricEmpty({ hasFilters, onClear }: { hasFilters: boolean; onClear: () => void }) { return <EmptyState icon={Activity} title={hasFilters ? "No metrics match these filters" : "No metrics yet"} description={hasFilters ? "Try broadening your filters." : "Metrics will appear when your connected infrastructure reports them."} action={hasFilters ? <button type="button" onClick={onClear} className="rounded-xl border px-3 py-2 text-sm">Clear filters</button> : undefined} />; }
