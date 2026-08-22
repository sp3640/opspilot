"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { Send, X } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useInviteUser } from "@/hooks/use-invitations";
import type { InviteRequest, InvitationResponse } from "@/types/invitation-api";

const inviteFormSchema = z.object({
  email: z.string().trim().email("Enter a valid email address."),
  role: z.enum(["Platform Admin", "DevOps Engineer", "Developer", "Viewer"], { message: "Select a valid role." }),
});

type InviteFormInput = z.infer<typeof inviteFormSchema>;

type InviteUserModalProps = {
  open: boolean;
  onClose: () => void;
  onInvited?: (invitation: InvitationResponse) => void;
};

/** Validated member-invitation workflow backed by the invitations API. */
export function InviteUserModal({ open, onClose, onInvited }: InviteUserModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const inviteUser = useInviteUser();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<InviteFormInput>({
    resolver: zodResolver(inviteFormSchema),
    defaultValues: { email: "", role: "Viewer" },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (inviteUser.isPending) return;
    reset();
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: InviteFormInput) => {
    setSubmitError(null);
    try {
      const invitation = await inviteUser.mutateAsync(input as InviteRequest);
      reset();
      onInvited?.(invitation);
      onClose();
    } catch (error) {
      setSubmitError(getInvitationErrorMessage(error));
    }
  };

  const inputClass =
    "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="invite-user-title"
      aria-describedby={submitError ? "invite-user-description invite-user-submit-error" : "invite-user-description"}
      className="m-auto w-[calc(100%-2rem)] max-w-lg rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <form onSubmit={handleSubmit(submit)}>
        <header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-center gap-3">
            <div
              className="rounded-2xl p-3"
              style={{
                color: "var(--primary)",
                backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)",
              }}
            >
              <Send aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="invite-user-title" className="text-lg font-semibold">
                Invite member
              </h2>
              <p id="invite-user-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Send an invitation to join this organization.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={inviteUser.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close invite member dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          <div>
            <label htmlFor="invite-email" className="text-sm font-medium">
              Email
            </label>
            <input
              id="invite-email"
              type="email"
              {...register("email")}
              placeholder="teammate@example.com"
              className={inputClass}
              style={{ borderColor: errors.email ? "var(--danger)" : "var(--border)" }}
              disabled={inviteUser.isPending}
              aria-invalid={Boolean(errors.email)}
              aria-describedby={errors.email ? "invite-email-error" : undefined}
            />
            {errors.email && (
              <p id="invite-email-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.email.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="invite-role" className="text-sm font-medium">
              Role
            </label>
            <select
              id="invite-role"
              {...register("role")}
              className={inputClass}
              style={{ borderColor: errors.role ? "var(--danger)" : "var(--border)" }}
              disabled={inviteUser.isPending}
              aria-invalid={Boolean(errors.role)}
              aria-describedby={errors.role ? "invite-role-error" : undefined}
            >
              <option value="Viewer">Viewer</option>
              <option value="Developer">Developer</option>
              <option value="DevOps Engineer">DevOps Engineer</option>
              <option value="Platform Admin">Platform Admin</option>
            </select>
            {errors.role && (
              <p id="invite-role-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.role.message}
              </p>
            )}
          </div>

          {submitError && (
            <p
              id="invite-user-submit-error"
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
          <Button type="button" variant="secondary" onClick={close} disabled={inviteUser.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={inviteUser.isPending} disabled={inviteUser.isPending}>
            Send invitation
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function getInvitationErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to send invitation. Please try again.";
  }
  return "Unable to send invitation. Please try again.";
}
