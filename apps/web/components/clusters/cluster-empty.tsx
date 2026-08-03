import { Server } from "lucide-react";

import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";

export function ClusterEmpty({
  hasFilters,
  onClear,
  onCreate,
}: {
  hasFilters: boolean;
  onClear: () => void;
  onCreate: () => void;
}) {
  return (
    <EmptyState
      icon={Server}
      title={hasFilters ? "No clusters match these filters" : "No clusters yet"}
      description={
        hasFilters
          ? "Try broadening your search or clearing active filters."
          : "Create your first cluster to connect infrastructure to OpsPilot."
      }
      action={
        hasFilters ? (
          <Button type="button" variant="secondary" onClick={onClear}>
            Clear filters
          </Button>
        ) : (
          <Button type="button" onClick={onCreate}>
            Create cluster
          </Button>
        )
      }
    />
  );
}
