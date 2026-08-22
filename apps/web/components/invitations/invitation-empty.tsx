import { Mail } from "lucide-react";

import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";

export function InvitationEmpty({
  hasFilters,
  onClear,
  onInvite,
}: {
  hasFilters: boolean;
  onClear: () => void;
  onInvite: () => void;
}) {
  return (
    <EmptyState
      icon={Mail}
      title={hasFilters ? "No invitations match these filters" : "No invitations yet"}
      description={
        hasFilters
          ? "Try broadening your search or clearing active filters."
          : "Invite your first member to join this organization."
      }
      action={
        hasFilters ? (
          <Button type="button" variant="secondary" onClick={onClear}>
            Clear filters
          </Button>
        ) : (
          <Button type="button" onClick={onInvite}>
            Invite member
          </Button>
        )
      }
    />
  );
}
