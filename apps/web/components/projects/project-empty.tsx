import { FolderKanban } from "lucide-react";

import { EmptyState } from "@/components/common";
import { Button } from "@/components/ui/button";

/** Projects-specific empty copy composed from the global EmptyState primitive. */
export function ProjectEmpty({ hasFilters, onClear, onCreate }: { hasFilters: boolean; onClear: () => void; onCreate: () => void }) {
  return <EmptyState icon={FolderKanban} title={hasFilters ? "No projects match these filters" : "No projects yet"} description={hasFilters ? "Try broadening your search or clearing active filters to find the project you need." : "Create your first project to begin organizing services, owners, and deployments."} action={hasFilters ? <Button type="button" variant="secondary" onClick={onClear}>Clear filters</Button> : <Button type="button" onClick={onCreate}>Create project</Button>} />;
}
