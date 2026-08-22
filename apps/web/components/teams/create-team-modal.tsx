"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { UsersRound, X } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useCreateTeam } from "@/hooks/use-teams";
import type { CreateTeamRequest, TeamResponse } from "@/types/team-api";

const teamFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(3, "Team name must contain at least 3 characters.")
    .max(100, "Team name must be 100 characters or fewer."),
  description: z
    .string()
    .trim()
    .max(500, "Description must be 500 characters or fewer."),
});

type TeamFormInput = z.infer<typeof teamFormSchema>;

type CreateTeamModalProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (team: TeamResponse) => void;
};

/** Validated team-creation workflow backed by the teams API. */
export function CreateTeamModal({ open, onClose, onCreated }: CreateTeamModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const createTeam = useCreateTeam();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<TeamFormInput>({
    resolver: zodResolver(teamFormSchema),
    defaultValues: { name: "", description: "" },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (createTeam.isPending) return;
    reset();
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: TeamFormInput) => {
    setSubmitError(null);
    try {
      const team = await createTeam.mutateAsync(input as CreateTeamRequest);
      reset();
      onCreated(team);
      onClose();
    } catch (error) {
      setSubmitError(getTeamErrorMessage(error));
    }
  };

  const inputClass =
    "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="create-team-title"
      aria-describedby={submitError ? "create-team-description create-team-submit-error" : "create-team-description"}
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
              <UsersRound aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="create-team-title" className="text-lg font-semibold">
                Create team
              </h2>
              <p id="create-team-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Organize members and manage access together.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={createTeam.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close create team dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          <div>
            <label htmlFor="team-name" className="text-sm font-medium">
              Team name
            </label>
            <input
              id="team-name"
              {...register("name")}
              placeholder="e.g. Platform Engineering"
              className={inputClass}
              style={{ borderColor: errors.name ? "var(--danger)" : "var(--border)" }}
              disabled={createTeam.isPending}
              aria-invalid={Boolean(errors.name)}
              aria-describedby={errors.name ? "team-name-error" : undefined}
            />
            {errors.name && (
              <p id="team-name-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.name.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="team-description" className="text-sm font-medium">
              Description
            </label>
            <textarea
              id="team-description"
              {...register("description")}
              placeholder="Describe this team's purpose and responsibilities."
              rows={4}
              className={inputClass}
              style={{ borderColor: errors.description ? "var(--danger)" : "var(--border)" }}
              disabled={createTeam.isPending}
              aria-invalid={Boolean(errors.description)}
              aria-describedby={errors.description ? "team-description-error" : undefined}
            />
            {errors.description && (
              <p id="team-description-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.description.message}
              </p>
            )}
          </div>

          {submitError && (
            <p
              id="create-team-submit-error"
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
          <Button type="button" variant="secondary" onClick={close} disabled={createTeam.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={createTeam.isPending} disabled={createTeam.isPending}>
            Create
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function getTeamErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to create team. Please try again.";
  }
  return "Unable to create team. Please try again.";
}
