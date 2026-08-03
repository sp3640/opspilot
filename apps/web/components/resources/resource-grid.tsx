import { ResourceCard } from "./resource-card";
import type { Resource } from "./types";
export function ResourceGrid({ resources, projectNameById, onOpen }: { resources: Resource[]; projectNameById: Map<string, string>; onOpen: (id: string) => void }) { return <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{resources.map((resource) => <ResourceCard key={resource.id} resource={resource} projectName={projectNameById.get(resource.projectId)} onOpen={onOpen} />)}</div>; }
