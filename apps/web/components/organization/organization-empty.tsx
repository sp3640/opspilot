import { Building2 } from "lucide-react";

import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";

type OrganizationEmptyProps = {
  onCreate: () => void;
};

export function OrganizationEmpty({ onCreate }: OrganizationEmptyProps) {
  return (
    <EmptyState
      title="No organization found"
      description="Create your first organization to start managing shared workspace details."
      icon={Building2}
      action={
        <Button type="button" onClick={onCreate}>
          Create organization
        </Button>
      }
    />
  );
}
