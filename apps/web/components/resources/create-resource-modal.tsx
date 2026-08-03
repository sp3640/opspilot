"use client";
import { useCreateResource } from "@/hooks/use-resources";
import type { ProjectResponse } from "@/types/project-api";
import type { ResourceResponse } from "@/types/resource-api";
import { ResourceModal } from "./resource-modal";
export function CreateResourceModal({ open, projects, onClose, onCreated }: { open: boolean; projects: ProjectResponse[]; onClose: () => void; onCreated: (resource: ResourceResponse) => void }) { const mutation = useCreateResource(); return <ResourceModal open={open} projects={projects} onClose={onClose} pending={mutation.isPending} onSubmit={async (payload) => { const resource = await mutation.mutateAsync(payload); onCreated(resource); }} />; }
