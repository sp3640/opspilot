import { BellRing } from "lucide-react";
import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useHasPermission } from "@/store/auth-store";
export function AlertEmpty({ hasFilters, onClear, onCreate }: { hasFilters: boolean; onClear: () => void; onCreate: () => void }) { const canManageAlerts = useHasPermission("alert:manage"); return <EmptyState icon={BellRing} title={hasFilters ? "No alerts match these filters" : "No alerts yet"} description={hasFilters ? "Try broadening your search or clearing active filters." : "Create an alert to start tracking signals."} action={hasFilters ? <Button onClick={onClear} variant="secondary">Clear filters</Button> : canManageAlerts ? <Button onClick={onCreate} variant="primary">Create alert</Button> : undefined} />; }
