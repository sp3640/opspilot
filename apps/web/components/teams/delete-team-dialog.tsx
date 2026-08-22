"use client";

import { useEffect, useRef, useState } from "react";
import axios from "axios";
import { AlertTriangle, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useDeleteTeam } from "@/hooks/use-teams";
import type { TeamResponse } from "@/types/team-api";

type DeleteTeamDialogProps = {
  open: boolean;
  team: TeamResponse | null;
  onClose: () => void;
  onSuccess: () => void;
};

export function DeleteTeamDialog({ open, team, onClose, onSuccess }: DeleteTeamDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const deleteTeam = useDeleteTeam();
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (deleteTeam.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const handleDelete = async () => {
    if (!team) return;
    setSubmitError(null);
    try {
      await deleteTeam.mutateAsync(team.id);
      onSuccess();
      onClose();
    } catch (error) {
      setSubmitError(getDeleteErrorMessage(error));
    }
  };

  if (!team) return null;

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      onMouseDown={(event) => event.stopPropagation()}
      aria-labelledby="delete-team-title"
      aria-describedby={submitError ? "delete-team-description delete-team-submit-error" : "delete-team-description"}
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
              <h2 id="delete-team-title" className="text-lg font-semibold">
                Delete team
              </h2>
              <p id="delete-team-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                This action cannot be undone.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={deleteTeam.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close delete team dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          {submitError ? (
            <div
              id="delete-team-submit-error"
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
            Delete team <strong>&quot;{team.name}&quot;</strong>? This permanently removes this team and its
            member assignments.
          </p>
        </div>

        <footer className="flex gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button
            type="button"
            variant="secondary"
            onClick={close}
            disabled={deleteTeam.isPending}
            className="flex-1"
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="danger"
            onClick={handleDelete}
            loading={deleteTeam.isPending}
            disabled={deleteTeam.isPending}
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
    return error.response?.data?.message ?? "Unable to delete team. Please try again.";
  }
  return "Unable to delete team. Please try again.";
}
