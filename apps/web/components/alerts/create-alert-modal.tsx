"use client";
import { useCreateAlert } from "@/hooks/use-alerts";
import type { ProjectResponse } from "@/types/project-api";
import { AlertModal } from "./alert-modal";
export function CreateAlertModal({ open, projects, onClose }: { open: boolean; projects: ProjectResponse[]; onClose: () => void }) { const mutation = useCreateAlert(); return <AlertModal open={open} projects={projects} onClose={onClose} pending={mutation.isPending} onSubmit={(payload) => mutation.mutateAsync(payload)} />; }
