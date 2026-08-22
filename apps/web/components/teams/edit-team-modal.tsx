"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { Pencil, X } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useUpdateTeam } from "@/hooks/use-teams";
import type { TeamResponse, UpdateTeamRequest } from "@/types/team-api";

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

type EditTeamModalProps = {
  open: boolean;
  team: TeamResponse | null;
  onClose: () => void;
};

export function EditTeamModal({ open, team, onClose }: EditTeamModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateTeam = useUpdateTeam();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TeamFormInput>({
    resolver: zodResolver(teamFormSchema),
    defaultValues: {
      name: team?.name ?? "",
      description: team?.description ?? "",
    },
    values: {
      name: team?.name ?? "",
      description: team?.description ?? "",
    },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (updateTeam.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: TeamFormInput) => {
    if (!team) return;
    setSubmitError(null);
    try {
      await updateTeam.mutateAsync({ id: team.id, payload: input as UpdateTeamRequest });
      onClose();
    } catch (error) {
      setSubmitError(getTeamErrorMessage(error));
    }
  };

  if (!team) return null;

  const inputClass =
    "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="edit-team-title"
      aria-describedby={submitError ? "edit-team-description edit-team-submit-error" : "edit-team-description"}
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
              <h2 id="edit-team-title" className="text-lg font-semibold">
                Edit team
              </h2>
              <p id="edit-team-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Update team name and description.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={updateTeam.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close edit team dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          <div>
            <label htmlFor="edit-team-name" className="text-sm font-medium">
              Team name
            </label>
            <input
              id="edit-team-name"
              {...register("name")}
              placeholder="e.g. Platform Engineering"
              className={inputClass}
              style={{ borderColor: errors.name ? "var(--danger)" : "var(--border)" }}
              disabled={updateTeam.isPending}
              aria-invalid={Boolean(errors.name)}
              aria-describedby={errors.name ? "edit-team-name-error" : undefined}
            />
            {errors.name && (
              <p id="edit-team-name-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.name.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="edit-team-description" className="text-sm font-medium">
              Description
            </label>
            <textarea
              id="edit-team-description"
              {...register("description")}
              placeholder="Describe this team's purpose and responsibilities."
              rows={4}
              className={inputClass}
              style={{ borderColor: errors.description ? "var(--danger)" : "var(--border)" }}
              disabled={updateTeam.isPending}
              aria-invalid={Boolean(errors.description)}
              aria-describedby={errors.description ? "edit-team-description-error" : undefined}
            />
            {errors.description && (
              <p id="edit-team-description-error" className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.description.message}
              </p>
            )}
          </div>

          {submitError && (
            <p
              id="edit-team-submit-error"
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
          <Button type="button" variant="secondary" onClick={close} disabled={updateTeam.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={updateTeam.isPending} disabled={updateTeam.isPending}>
            Save changes
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function getTeamErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to update team. Please try again.";
  }
  return "Unable to update team. Please try again.";
}
