import { Boxes } from "lucide-react";
import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";
export function ResourceEmpty({ hasFilters, onClear, onCreate }: { hasFilters: boolean; onClear: () => void; onCreate: () => void }) { return <EmptyState icon={Boxes} title={hasFilters ? "No resources match these filters" : "No resources yet"} description={hasFilters ? "Try broadening your search or clearing active filters." : "Create a resource or synchronize a connected cluster to discover resources."} action={<Button type="button" variant={hasFilters ? "secondary" : "primary"} onClick={hasFilters ? onClear : onCreate}>{hasFilters ? "Clear filters" : "Create resource"}</Button>} />; }
