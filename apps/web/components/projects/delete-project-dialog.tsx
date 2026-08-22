"use client";

import { useEffect, useRef, useState } from "react";
import axios from "axios";
import { AlertTriangle, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useDeleteProject } from "@/hooks/use-projects";
import type { ProjectResponse } from "@/types/project-api";

type DeleteProjectDialogProps = {
  open: boolean;
  project: ProjectResponse | null;
  onClose: () => void;
  onSuccess: () => void;
};

export function DeleteProjectDialog({ open, project, onClose, onSuccess }: DeleteProjectDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const deleteProject = useDeleteProject();
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (deleteProject.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const handleDelete = async () => {
    if (!project) return;
    setSubmitError(null);
    try {
      await deleteProject.mutateAsync(project.id);
      onSuccess();
      onClose();
    } catch (error) {
      setSubmitError(getDeleteErrorMessage(error));
    }
  };

  if (!project) return null;

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      onMouseDown={(event) => event.stopPropagation()}
      aria-labelledby="delete-project-title"
      aria-describedby={submitError ? "delete-project-description delete-project-submit-error" : "delete-project-description"}
      className="m-auto w-[calc(100%-2rem)] max-w-md rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <div className="flex flex-col">
        <header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-center gap-3">
            <div
              className="rounded-2xl p-3"
              style={{
                color: "var(--danger)",
                backgroundColor: "color-mix(in srgb, var(--danger) 12%, transparent)",
              }}
            >
              <AlertTriangle aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="delete-project-title" className="text-lg font-semibold">
                Delete project
              </h2>
              <p id="delete-project-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                This action cannot be undone.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={deleteProject.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close delete project dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          {submitError ? (
            <div
              id="delete-project-submit-error"
              role="alert"
              aria-live="assertive"
              className="rounded-2xl border p-3 text-sm"
              style={{
                backgroundColor: "color-mix(in srgb, var(--danger) 12%, transparent)",
                borderColor: "var(--danger)",
                color: "var(--danger)",
              }}
            >
              {submitError}
            </div>
          ) : null}

          <p className="text-sm leading-6">
            Delete project <strong>&quot;{project.name}&quot;</strong>? This permanently removes this project
            and cannot be undone.
          </p>
        </div>

        <footer className="flex gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button
            type="button"
            variant="secondary"
            onClick={close}
            disabled={deleteProject.isPending}
            className="flex-1"
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="danger"
            onClick={handleDelete}
            loading={deleteProject.isPending}
            disabled={deleteProject.isPending}
            className="flex-1"
          >
            Delete
          </Button>
        </footer>
      </div>
    </dialog>
  );
}

function getDeleteErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to delete project. Please try again.";
  }
  return "Unable to delete project. Please try again.";
}
