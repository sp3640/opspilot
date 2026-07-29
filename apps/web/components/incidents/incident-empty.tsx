import { AlertCircle } from "lucide-react";

import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";

export function IncidentEmpty({
  hasFilters,
  onClear,
}: {
  hasFilters: boolean;
  onClear: () => void;
}) {
  return (
    <EmptyState
      icon={AlertCircle}
      title={hasFilters ? "No incidents match these filters" : "No incidents yet"}
      description={
        hasFilters
          ? "Try broadening your search to find the incident you need."
          : "Incidents will appear here when created."
      }
      action={
        hasFilters && (
          <Button type="button" variant="secondary" onClick={onClear}>
            Clear filters
          </Button>
        )
      }
    />
  );
}
