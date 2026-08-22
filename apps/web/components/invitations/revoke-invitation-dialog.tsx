"use client";

import { useEffect, useRef, useState } from "react";
import axios from "axios";
import { AlertTriangle, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useRevokeInvitation } from "@/hooks/use-invitations";
import type { InvitationResponse } from "@/types/invitation-api";

type RevokeInvitationDialogProps = {
  open: boolean;
  invitation: InvitationResponse | null;
  onClose: () => void;
};

export function RevokeInvitationDialog({ open, invitation, onClose }: RevokeInvitationDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const revokeInvitation = useRevokeInvitation();
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (revokeInvitation.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const handleRevoke = async () => {
    if (!invitation) return;
    setSubmitError(null);
    try {
      await revokeInvitation.mutateAsync(invitation.id);
      onClose();
    } catch (error) {
      setSubmitError(getRevokeErrorMessage(error));
    }
  };

  if (!invitation) return null;

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      onMouseDown={(event) => event.stopPropagation()}
      aria-labelledby="revoke-invitation-title"
      aria-describedby={submitError ? "revoke-invitation-description revoke-invitation-submit-error" : "revoke-invitation-description"}
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
              <h2 id="revoke-invitation-title" className="text-lg font-semibold">
                Revoke invitation
              </h2>
              <p id="revoke-invitation-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                This action cannot be undone.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={revokeInvitation.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close revoke invitation dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          {submitError ? (
            <div
              id="revoke-invitation-submit-error"
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
            Revoke the invitation sent to <strong>&quot;{invitation.email}&quot;</strong>? They will no longer be
            able to join using this invitation.
          </p>
        </div>

        <footer className="flex gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button
            type="button"
            variant="secondary"
            onClick={close}
            disabled={revokeInvitation.isPending}
            className="flex-1"
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="danger"
            onClick={handleRevoke}
            loading={revokeInvitation.isPending}
            disabled={revokeInvitation.isPending}
            className="flex-1"
          >
            Revoke
          </Button>
        </footer>
      </div>
    </dialog>
  );
}

function getRevokeErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to revoke invitation. Please try again.";
  }
  return "Unable to revoke invitation. Please try again.";
}
