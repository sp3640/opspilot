"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { AlertCircle, X } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useCreateIncident } from "@/hooks/use-incidents";
import type { CreateIncidentRequest } from "@/types/incident-api";
import { ProjectSelect } from "./project-select";

const incidentSchema = z.object({
  title: z
    .string()
    .trim()
    .min(3, "Title must contain at least 3 characters.")
    .max(255, "Title must be 255 characters or fewer."),
  description: z.string().trim().max(1000, "Description must be 1000 characters or fewer."),
  severity: z.enum(["P0", "P1", "P2", "P3", "P4"], { message: "Please select a valid severity level." }),
  project_id: z.string().trim().min(1, "Please select a project."),
  status: z.literal("OPEN"),
});

type CreateIncidentFormData = z.infer<typeof incidentSchema>;

type CreateIncidentModalProps = {
  open: boolean;
  onClose: () => void;
};

/** Validated incident-creation workflow backed by the incidents API. */
export function CreateIncidentModal({ open, onClose }: CreateIncidentModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const createIncident = useCreateIncident();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateIncidentFormData>({
    resolver: zodResolver(incidentSchema),
    defaultValues: {
      title: "",
      description: "",
      severity: "P0",
      project_id: "",
      status: "OPEN",
    },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (createIncident.isPending) return;
    reset();
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: CreateIncidentFormData) => {
    setSubmitError(null);
    try {
      await createIncident.mutateAsync(input as CreateIncidentRequest);
      reset();
      onClose();
    } catch (error) {
      setSubmitError(getIncidentErrorMessage(error));
    }
  };

  const inputClass =
    "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="create-incident-title"
      className="m-auto w-[calc(100%-2rem)] max-w-lg rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <form onSubmit={handleSubmit(submit)}>
        <header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-center gap-3">
            <div
              className="rounded-2xl p-3"
              style={{
                color: "var(--danger)",
                backgroundColor: "color-mix(in srgb, var(--danger) 12%, transparent)",
              }}
            >
              <AlertCircle aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="create-incident-title" className="text-lg font-semibold">
                Create incident
              </h2>
              <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Report a new issue affecting your platform.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={createIncident.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close create incident dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          {submitError && (
            <div
              className="rounded-2xl border p-3 text-sm"
              style={{
                backgroundColor: "color-mix(in srgb, var(--danger) 12%, transparent)",
                borderColor: "var(--danger)",
                color: "var(--danger)",
              }}
            >
              {submitError}
            </div>
          )}

          <div>
            <label htmlFor="incident-title" className="text-sm font-medium">
              Title
            </label>
            <input
              id="incident-title"
              {...register("title")}
              placeholder="e.g. Database connection timeout"
              className={inputClass}
              style={{ borderColor: errors.title ? "var(--danger)" : "var(--border)" }}
              disabled={createIncident.isPending}
            />
            {errors.title && (
              <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.title.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="incident-description" className="text-sm font-medium">
              Description
            </label>
            <textarea
              id="incident-description"
              {...register("description")}
              placeholder="Describe the issue and its impact on users."
              rows={4}
              className={inputClass}
              style={{ borderColor: errors.description ? "var(--danger)" : "var(--border)" }}
              disabled={createIncident.isPending}
            />
            {errors.description && (
              <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.description.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="incident-severity" className="text-sm font-medium">
              Severity
            </label>
            <select
              id="incident-severity"
              {...register("severity")}
              className={inputClass}
              style={{ borderColor: errors.severity ? "var(--danger)" : "var(--border)" }}
              disabled={createIncident.isPending}
            >
              <option value="P0">P0 - Critical</option>
              <option value="P1">P1 - High</option>
              <option value="P2">P2 - Medium</option>
              <option value="P3">P3 - Low</option>
              <option value="P4">P4 - Minimal</option>
            </select>
            {errors.severity && (
              <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.severity.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="incident-project" className="text-sm font-medium">
              Project
            </label>
            <ProjectSelect
              id="incident-project"
              queryEnabled={open}
              {...register("project_id")}
              className={inputClass}
              style={{ borderColor: errors.project_id ? "var(--danger)" : "var(--border)" }}
              disabled={createIncident.isPending}
            />
            {errors.project_id && (
              <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.project_id.message}
              </p>
            )}
          </div>
        </div>

        <footer className="flex gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={createIncident.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={createIncident.isPending} disabled={createIncident.isPending}>
            Create incident
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function getIncidentErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to create incident. Please try again.";
  }
  return "Unable to create incident. Please try again.";
}
