import { BellRing } from "lucide-react";
import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";
export function AlertEmpty({ hasFilters, onClear, onCreate }: { hasFilters: boolean; onClear: () => void; onCreate: () => void }) { return <EmptyState icon={BellRing} title={hasFilters ? "No alerts match these filters" : "No alerts yet"} description={hasFilters ? "Try broadening your search or clearing active filters." : "Create an alert to start tracking signals."} action={<Button onClick={hasFilters ? onClear : onCreate} variant={hasFilters ? "secondary" : "primary"}>{hasFilters ? "Clear filters" : "Create alert"}</Button>} />; }
