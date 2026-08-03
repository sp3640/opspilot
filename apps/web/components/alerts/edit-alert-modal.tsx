"use client";
import { useUpdateAlert } from "@/hooks/use-alerts";
import type { ProjectResponse } from "@/types/project-api";
import type { AlertResponse } from "@/types/alert-api";
import { AlertModal } from "./alert-modal";
export function EditAlertModal({ open, alert, projects, onClose }: { open: boolean; alert: AlertResponse | null; projects: ProjectResponse[]; onClose: () => void }) { const mutation = useUpdateAlert(); return <AlertModal open={open} alert={alert} projects={projects} onClose={onClose} pending={mutation.isPending} onSubmit={(payload) => mutation.mutateAsync({ id: alert?.id ?? 0, payload })} />; }
