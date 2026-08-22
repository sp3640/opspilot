"use client";

import { useEffect, useRef, useState } from "react";
import axios from "axios";
import { AlertTriangle, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useDeleteDeployment } from "@/hooks/use-deployments";
import type { DeploymentResponse } from "@/types/deployment-api";

type DeleteDeploymentDialogProps = {
  open: boolean;
  deployment: DeploymentResponse | null;
  onClose: () => void;
  onSuccess: () => void;
};

/** Confirmation dialog for deleting a deployment record. Requires explicit user confirmation. */
export function DeleteDeploymentDialog({ open, deployment, onClose, onSuccess }: DeleteDeploymentDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const deleteDeployment = useDeleteDeployment();
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !deleteDeployment.isPending) onClose();
    };
    if (open) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [open, onClose, deleteDeployment.isPending]);

  const close = () => {
    if (deleteDeployment.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const handleDelete = async () => {
    if (!deployment) return;
    setSubmitError(null);
    try {
      await deleteDeployment.mutateAsync(deployment.id);
      onSuccess();
      onClose();
    } catch (error) {
      setSubmitError(getDeleteErrorMessage(error));
    }
  };

  if (!deployment) return null;

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      onMouseDown={(event) => event.stopPropagation()}
      aria-labelledby="delete-deployment-title"
      aria-describedby={submitError ? "delete-deployment-description delete-deployment-submit-error" : "delete-deployment-description"}
      className="m-auto w-[calc(100%-2rem)] max-w-md rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <div className="flex flex-col">
        <header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-center gap-3">
            <div
              className="rounded-2xl p-3"
              style={{ color: "var(--danger)", backgroundColor: "color-mix(in srgb, var(--danger) 12%, transparent)" }}
            >
              <AlertTriangle aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="delete-deployment-title" className="text-lg font-semibold">
                Delete deployment
              </h2>
              <p id="delete-deployment-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                This action cannot be undone.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={deleteDeployment.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close delete confirmation dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          {submitError && (
            <div
              id="delete-deployment-submit-error"
              role="alert"
              aria-live="assertive"
              className="rounded-2xl border p-3 text-sm"
              style={{ backgroundColor: "color-mix(in srgb, var(--danger) 12%, transparent)", borderColor: "var(--danger)", color: "var(--danger)" }}
            >
              {submitError}
            </div>
          )}

          <p className="text-sm leading-6">
            Delete the deployment record <strong>&quot;{deployment.imageTag ? `${deployment.image}:${deployment.imageTag}` : deployment.image}&quot;</strong>?
            This will remove it permanently along with its revision history.
          </p>
        </div>

        <footer className="flex gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={deleteDeployment.isPending} className="flex-1">
            Cancel
          </Button>
          <Button
            type="button"
            onClick={handleDelete}
            loading={deleteDeployment.isPending}
            disabled={deleteDeployment.isPending}
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
    return error.response?.data?.message ?? "Unable to delete deployment. Please try again.";
  }
  return "Unable to delete deployment. Please try again.";
}
