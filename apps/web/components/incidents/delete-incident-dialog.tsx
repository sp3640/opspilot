"use client";

import { useEffect, useRef, useState } from "react";
import axios from "axios";
import { AlertTriangle, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useDeleteIncident } from "@/hooks/use-incidents";
import type { IncidentResponse } from "@/types/incident-api";

type DeleteIncidentDialogProps = {
  open: boolean;
  incident: IncidentResponse | null;
  onClose: () => void;
  onSuccess: () => void;
};

/** Confirmation dialog for deleting an incident. Requires explicit user confirmation. */
export function DeleteIncidentDialog({ open, incident, onClose, onSuccess }: DeleteIncidentDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const deleteIncident = useDeleteIncident();
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !deleteIncident.isPending) onClose();
    };
    if (open) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [open, onClose, deleteIncident.isPending]);

  const close = () => {
    if (deleteIncident.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const handleDelete = async () => {
    if (!incident) return;
    setSubmitError(null);
    try {
      await deleteIncident.mutateAsync(incident.id);
      onSuccess();
      onClose();
    } catch (error) {
      setSubmitError(getDeleteErrorMessage(error));
    }
  };

  if (!incident) return null;

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="delete-incident-title"
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
              <h2 id="delete-incident-title" className="text-lg font-semibold">
                Delete incident
              </h2>
              <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                This action cannot be undone.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={deleteIncident.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close delete confirmation dialog"
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
            <p className="text-sm leading-6">
              Delete incident <strong>"{incident.title}"</strong>? This will remove it permanently along with all associated data.
            </p>
          </div>
        </div>

        <footer className="flex gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button
            type="button"
            variant="secondary"
            onClick={close}
            disabled={deleteIncident.isPending}
            className="flex-1"
          >
            Cancel
          </Button>
          <Button
            type="button"
            onClick={handleDelete}
            loading={deleteIncident.isPending}
            disabled={deleteIncident.isPending}
            className="flex-1 bg-[var(--danger)] hover:bg-[color-mix(in_srgb,var(--danger)_85%,black)]"
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
    return error.response?.data?.message ?? "Unable to delete incident. Please try again.";
  }
  return "Unable to delete incident. Please try again.";
}
