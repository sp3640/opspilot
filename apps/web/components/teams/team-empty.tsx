import { Users } from "lucide-react";

import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useIsPlatformAdmin } from "@/store/auth-store";

export function TeamEmpty({
  hasFilters,
  onClear,
  onCreate,
}: {
  hasFilters: boolean;
  onClear: () => void;
  onCreate: () => void;
}) {
  const isAdmin = useIsPlatformAdmin();
  return (
    <EmptyState
      icon={Users}
      title={hasFilters ? "No teams match these filters" : "No teams yet"}
      description={
        hasFilters
          ? "Try broadening your search or clearing active filters."
          : "Create your first team to organize members across your organization."
      }
      action={
        hasFilters ? (
          <Button type="button" variant="secondary" onClick={onClear}>
            Clear filters
          </Button>
        ) : isAdmin ? (
          <Button type="button" onClick={onCreate}>
            Create team
          </Button>
        ) : undefined
      }
    />
  );
}
