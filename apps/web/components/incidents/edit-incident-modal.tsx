"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { AlertCircle, X } from "lucide-react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { Button } from "@/components/ui/button";
import { useUpdateIncident } from "@/hooks/use-incidents";
import {
  INCIDENT_DEFAULT_SEVERITY,
  INCIDENT_DEFAULT_STATUS,
  INCIDENT_SEVERITY_OPTIONS,
  INCIDENT_STATUS_OPTIONS,
  type IncidentSeverity,
  type IncidentStatus,
} from "@/lib/constants";
import { editIncidentSchema } from "@/lib/validation/incident";
import type { IncidentResponse, UpdateIncidentRequest } from "@/types/incident-api";
import { ProjectSelect } from "./project-select";

type EditIncidentFormData = z.infer<typeof editIncidentSchema>;

type EditIncidentModalProps = {
  open: boolean;
  incident: IncidentResponse | null;
  onClose: () => void;
};

/** Validated incident-editing workflow backed by the incidents API. */
export function EditIncidentModal({ open, incident, onClose }: EditIncidentModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateIncident = useUpdateIncident();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<EditIncidentFormData>({
    resolver: zodResolver(editIncidentSchema),
    defaultValues: {
      title: incident?.title ?? "",
      description: incident?.description ?? "",
      severity: (incident?.severity as IncidentSeverity) ?? INCIDENT_DEFAULT_SEVERITY,
      project_id: incident?.projectId ?? "",
      status: (incident?.status as IncidentStatus) ?? INCIDENT_DEFAULT_STATUS,
    },
    values: {
      title: incident?.title ?? "",
      description: incident?.description ?? "",
      severity: (incident?.severity as IncidentSeverity) ?? INCIDENT_DEFAULT_SEVERITY,
      project_id: incident?.projectId ?? "",
      status: (incident?.status as IncidentStatus) ?? INCIDENT_DEFAULT_STATUS,
    },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (updateIncident.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: EditIncidentFormData) => {
    if (!incident) return;
    setSubmitError(null);
    try {
      await updateIncident.mutateAsync({
        id: incident.id,
        payload: {
          ...input,
        } as UpdateIncidentRequest,
      });
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
      onMouseDown={(event) => event.stopPropagation()}
      aria-labelledby="edit-incident-title"
      aria-describedby={submitError ? "edit-incident-summary edit-incident-submit-error" : "edit-incident-summary"}
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
              <h2 id="edit-incident-title" className="text-lg font-semibold">
                Edit incident
              </h2>
              <p id="edit-incident-summary" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Update incident details.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={updateIncident.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close edit incident dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          {submitError && (
            <div
              id="edit-incident-submit-error"
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
          )}

          <div>
            <label htmlFor="edit-incident-title-field" className="text-sm font-medium">
              Title
            </label>
            <input
              id="edit-incident-title-field"
              {...register("title")}
              placeholder="e.g. Database connection timeout"
              className={inputClass}
              style={{ borderColor: errors.title ? "var(--danger)" : "var(--border)" }}
              disabled={updateIncident.isPending}
              aria-invalid={Boolean(errors.title)}
              aria-describedby={errors.title ? "edit-incident-title-error" : undefined}
            />
            {errors.title && (
              <p id="edit-incident-title-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.title.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="edit-incident-description" className="text-sm font-medium">
              Description
            </label>
            <textarea
              id="edit-incident-description"
              {...register("description")}
              placeholder="Describe the issue and its impact on users."
              rows={4}
              className={inputClass}
              style={{ borderColor: errors.description ? "var(--danger)" : "var(--border)" }}
              disabled={updateIncident.isPending}
              aria-invalid={Boolean(errors.description)}
              aria-describedby={errors.description ? "edit-incident-description-error" : undefined}
            />
            {errors.description && (
              <p id="edit-incident-description-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.description.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="edit-incident-severity" className="text-sm font-medium">
              Severity
            </label>
            <select
              id="edit-incident-severity"
              {...register("severity")}
              className={inputClass}
              style={{ borderColor: errors.severity ? "var(--danger)" : "var(--border)" }}
              disabled={updateIncident.isPending}
              aria-invalid={Boolean(errors.severity)}
              aria-describedby={errors.severity ? "edit-incident-severity-error" : undefined}
            >
              {INCIDENT_SEVERITY_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            {errors.severity && (
              <p id="edit-incident-severity-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.severity.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="edit-incident-project" className="text-sm font-medium">
              Project
            </label>
            <ProjectSelect
              id="edit-incident-project"
              queryEnabled={open}
              {...register("project_id")}
              currentProjectId={incident?.projectId}
              className={inputClass}
              style={{ borderColor: errors.project_id ? "var(--danger)" : "var(--border)" }}
              disabled={updateIncident.isPending}
              aria-invalid={Boolean(errors.project_id)}
              aria-describedby={errors.project_id ? "edit-incident-project-error" : undefined}
            />
            {errors.project_id && (
              <p id="edit-incident-project-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.project_id.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="edit-incident-status" className="text-sm font-medium">
              Status
            </label>
            <select
              id="edit-incident-status"
              {...register("status")}
              className={inputClass}
              style={{ borderColor: errors.status ? "var(--danger)" : "var(--border)" }}
              disabled={updateIncident.isPending}
              aria-invalid={Boolean(errors.status)}
              aria-describedby={errors.status ? "edit-incident-status-error" : undefined}
            >
              {INCIDENT_STATUS_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            {errors.status && (
              <p id="edit-incident-status-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.status.message}
              </p>
            )}
          </div>
        </div>

        <footer className="flex gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={updateIncident.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={updateIncident.isPending} disabled={updateIncident.isPending}>
            Update incident
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function getIncidentErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to update incident. Please try again.";
  }
  return "Unable to update incident. Please try again.";
}
