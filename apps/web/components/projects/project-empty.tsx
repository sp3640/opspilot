import { FolderKanban } from "lucide-react";

import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useIsPlatformAdmin } from "@/store/auth-store";

/** Projects-specific empty copy composed from the global EmptyState primitive. */
export function ProjectEmpty({ hasFilters, onClear, onCreate }: { hasFilters: boolean; onClear: () => void; onCreate: () => void }) {
  const isAdmin = useIsPlatformAdmin();
  return <EmptyState icon={FolderKanban} title={hasFilters ? "No projects match these filters" : "No projects yet"} description={hasFilters ? "Try broadening your search or clearing active filters to find the project you need." : "Create your first project to begin organizing services, owners, and deployments."} action={hasFilters ? <Button type="button" variant="secondary" onClick={onClear}>Clear filters</Button> : isAdmin ? <Button type="button" onClick={onCreate}>Create project</Button> : undefined} />;
}
