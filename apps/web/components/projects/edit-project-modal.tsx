"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { Pencil, X } from "lucide-react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { useUpdateProject } from "@/hooks/use-projects";
import { projectFormSchema } from "@/lib/validation/project";
import type { ProjectResponse, UpdateProjectRequest } from "@/types/project-api";

type EditProjectModalProps = {
  open: boolean;
  project: ProjectResponse | null;
  onClose: () => void;
};

/** Validated project-editing workflow backed by the projects API. */
export function EditProjectModal({ open, project, onClose }: EditProjectModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateProject = useUpdateProject();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<UpdateProjectRequest>({
    resolver: zodResolver(projectFormSchema),
    defaultValues: {
      name: project?.name ?? "",
      description: project?.description ?? "",
    },
    values: {
      name: project?.name ?? "",
      description: project?.description ?? "",
    },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (updateProject.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: UpdateProjectRequest) => {
    if (!project) return;
    setSubmitError(null);
    try {
      await updateProject.mutateAsync({ id: project.id, payload: input });
      onClose();
    } catch (error) {
      setSubmitError(getProjectErrorMessage(error));
    }
  };

  if (!project) return null;

  const inputClass =
    "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="edit-project-title"
      aria-describedby={submitError ? "edit-project-description edit-project-submit-error" : "edit-project-description"}
      className="m-auto w-[calc(100%-2rem)] max-w-lg rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <form onSubmit={handleSubmit(submit)}>
        <header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-center gap-3">
            <div
              className="rounded-2xl p-3"
              style={{
                color: "var(--warning)",
                backgroundColor: "color-mix(in srgb, var(--warning) 12%, transparent)",
              }}
            >
              <Pencil aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="edit-project-title" className="text-lg font-semibold">
                Edit project
              </h2>
              <p id="edit-project-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Update the project name and description.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={updateProject.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close edit project dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          <div>
            <label htmlFor="edit-project-name" className="text-sm font-medium">
              Project name
            </label>
            <input
              id="edit-project-name"
              {...register("name")}
              placeholder="e.g. Customer API"
              className={inputClass}
              style={{ borderColor: errors.name ? "var(--danger)" : "var(--border)" }}
              disabled={updateProject.isPending}
              aria-invalid={Boolean(errors.name)}
              aria-describedby={errors.name ? "edit-project-name-error" : undefined}
            />
            {errors.name && (
              <p id="edit-project-name-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.name.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="edit-project-description" className="text-sm font-medium">
              Description
            </label>
            <textarea
              id="edit-project-description"
              {...register("description")}
              placeholder="Describe the services and responsibility of this project."
              rows={4}
              className={inputClass}
              style={{ borderColor: errors.description ? "var(--danger)" : "var(--border)" }}
              disabled={updateProject.isPending}
              aria-invalid={Boolean(errors.description)}
              aria-describedby={errors.description ? "edit-project-description-error" : undefined}
            />
            {errors.description && (
              <p id="edit-project-description-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.description.message}
              </p>
            )}
          </div>

          {submitError && (
            <p
              id="edit-project-submit-error"
              role="alert"
              aria-live="assertive"
              className="text-sm"
              style={{ color: "var(--danger)" }}
            >
              {submitError}
            </p>
          )}
        </div>

        <footer className="flex justify-end gap-3 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={updateProject.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={updateProject.isPending} disabled={updateProject.isPending}>
            Save changes
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function getProjectErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to update project. Please try again.";
  }
  return "Unable to update project. Please try again.";
}
