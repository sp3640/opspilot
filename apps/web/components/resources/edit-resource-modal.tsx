"use client";
import { useUpdateResource } from "@/hooks/use-resources";
import type { ProjectResponse } from "@/types/project-api";
import type { ResourceResponse } from "@/types/resource-api";
import { ResourceModal } from "./resource-modal";
export function EditResourceModal({ open, resource, projects, onClose }: { open: boolean; resource: ResourceResponse | null; projects: ProjectResponse[]; onClose: () => void }) { const mutation = useUpdateResource(); return <ResourceModal open={open} resource={resource} projects={projects} onClose={onClose} pending={mutation.isPending} onSubmit={(payload) => mutation.mutateAsync({ id: resource?.id ?? "", payload })} />; }
