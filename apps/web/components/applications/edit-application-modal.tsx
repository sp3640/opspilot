"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { Pencil, X } from "lucide-react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { useUpdateApplication } from "@/hooks/use-applications";
import { applicationSchema, type ApplicationFormValues } from "@/lib/validation/application";
import type { ApplicationResponse, UpdateApplicationRequest } from "@/types/application-api";

type EditApplicationModalProps = {
  open: boolean;
  application: ApplicationResponse | null;
  onClose: () => void;
};

const RUNTIME_OPTIONS = ["NodeJS", "Go", "Java", "Python", "DotNet", "Static", "Docker"] as const;
const STATUS_OPTIONS = ["Draft", "Ready", "Archived"] as const;

const inputClass =
  "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

export function EditApplicationModal({ open, application, onClose }: EditApplicationModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateApplication = useUpdateApplication();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ApplicationFormValues>({
    resolver: zodResolver(applicationSchema),
    defaultValues: {
      name: application?.name ?? "",
      slug: application?.slug ?? "",
      description: application?.description ?? "",
      repository_url: application?.repositoryUrl ?? "",
      default_branch: application?.defaultBranch ?? "",
      runtime: (application?.runtime as ApplicationFormValues["runtime"]) ?? "NodeJS",
      build_command: application?.buildCommand ?? "",
      start_command: application?.startCommand ?? "",
      port: application?.port ?? 0,
      environment: application?.environment ?? "",
      status: (application?.status as ApplicationFormValues["status"]) ?? "Draft",
    },
    values: {
      name: application?.name ?? "",
      slug: application?.slug ?? "",
      description: application?.description ?? "",
      repository_url: application?.repositoryUrl ?? "",
      default_branch: application?.defaultBranch ?? "",
      runtime: (application?.runtime as ApplicationFormValues["runtime"]) ?? "NodeJS",
      build_command: application?.buildCommand ?? "",
      start_command: application?.startCommand ?? "",
      port: application?.port ?? 0,
      environment: application?.environment ?? "",
      status: (application?.status as ApplicationFormValues["status"]) ?? "Draft",
    },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (updateApplication.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: ApplicationFormValues) => {
    if (!application) return;
    setSubmitError(null);

    const payload: UpdateApplicationRequest = {
      name: input.name,
      slug: input.slug,
      description: input.description,
      repository_url: input.repository_url,
      default_branch: input.default_branch,
      runtime: input.runtime,
      build_command: input.build_command,
      start_command: input.start_command,
      port: input.port,
      environment: input.environment,
      status: input.status,
    };

    try {
      await updateApplication.mutateAsync({ id: application.id, payload });
      reset();
      onClose();
    } catch (error) {
      setSubmitError(getApplicationErrorMessage(error));
    }
  };

  if (!application) return null;

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="edit-application-title"
      aria-describedby={submitError ? "edit-application-description edit-application-submit-error" : "edit-application-description"}
      className="m-auto w-[calc(100%-2rem)] max-w-2xl rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]"
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
              <h2 id="edit-application-title" className="text-lg font-semibold">
                Edit application
              </h2>
              <p id="edit-application-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Update application configuration fields.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={updateApplication.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close edit application dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          <div>
            <label htmlFor="edit-application-name" className="text-sm font-medium">
              Name
            </label>
            <input
              id="edit-application-name"
              {...register("name")}
              placeholder="e.g. Customer API"
              className={inputClass}
              style={{ borderColor: errors.name ? "var(--danger)" : "var(--border)" }}
              disabled={updateApplication.isPending}
              aria-invalid={Boolean(errors.name)}
            />
            {errors.name && (
              <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.name.message}
              </p>
            )}
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="edit-application-runtime" className="text-sm font-medium">
                Runtime
              </label>
              <select
                id="edit-application-runtime"
                {...register("runtime")}
                className={inputClass}
                style={{ borderColor: errors.runtime ? "var(--danger)" : "var(--border)" }}
                disabled={updateApplication.isPending}
                aria-invalid={Boolean(errors.runtime)}
              >
                {RUNTIME_OPTIONS.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
              {errors.runtime && (
                <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                  {errors.runtime.message}
                </p>
              )}
            </div>

            <div>
              <label htmlFor="edit-application-port" className="text-sm font-medium">
                Port
              </label>
              <input
                id="edit-application-port"
                type="number"
                min={1}
                max={65535}
                {...register("port", { valueAsNumber: true })}
                placeholder="e.g. 3000"
                className={inputClass}
                style={{ borderColor: errors.port ? "var(--danger)" : "var(--border)" }}
                disabled={updateApplication.isPending}
                aria-invalid={Boolean(errors.port)}
              />
              {errors.port && (
                <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                  {errors.port.message}
                </p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="edit-application-slug" className="text-sm font-medium">
                Slug
              </label>
              <input
                id="edit-application-slug"
                {...register("slug")}
                placeholder="e.g. customer-api"
                className={inputClass}
                style={{ borderColor: errors.slug ? "var(--danger)" : "var(--border)" }}
                disabled={updateApplication.isPending}
                aria-invalid={Boolean(errors.slug)}
              />
              {errors.slug && (
                <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                  {errors.slug.message}
                </p>
              )}
            </div>

            <div>
              <label htmlFor="edit-application-default-branch" className="text-sm font-medium">
                Default branch
              </label>
              <input
                id="edit-application-default-branch"
                {...register("default_branch")}
                placeholder="e.g. main"
                className={inputClass}
                style={{ borderColor: errors.default_branch ? "var(--danger)" : "var(--border)" }}
                disabled={updateApplication.isPending}
                aria-invalid={Boolean(errors.default_branch)}
              />
              {errors.default_branch && (
                <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                  {errors.default_branch.message}
                </p>
              )}
            </div>
          </div>

          <div>
            <label htmlFor="edit-application-repository-url" className="text-sm font-medium">
              Repository URL
            </label>
            <input
              id="edit-application-repository-url"
              {...register("repository_url")}
              placeholder="https://github.com/org/repo"
              className={inputClass}
              style={{ borderColor: errors.repository_url ? "var(--danger)" : "var(--border)" }}
              disabled={updateApplication.isPending}
              aria-invalid={Boolean(errors.repository_url)}
            />
            {errors.repository_url && (
              <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.repository_url.message}
              </p>
            )}
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="edit-application-build-command" className="text-sm font-medium">
                Build command
              </label>
              <input
                id="edit-application-build-command"
                {...register("build_command")}
                placeholder="e.g. npm run build"
                className={inputClass}
                style={{ borderColor: errors.build_command ? "var(--danger)" : "var(--border)" }}
                disabled={updateApplication.isPending}
                aria-invalid={Boolean(errors.build_command)}
              />
              {errors.build_command && (
                <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                  {errors.build_command.message}
                </p>
              )}
            </div>

            <div>
              <label htmlFor="edit-application-start-command" className="text-sm font-medium">
                Start command
              </label>
              <input
                id="edit-application-start-command"
                {...register("start_command")}
                placeholder="e.g. npm run start"
                className={inputClass}
                style={{ borderColor: errors.start_command ? "var(--danger)" : "var(--border)" }}
                disabled={updateApplication.isPending}
                aria-invalid={Boolean(errors.start_command)}
              />
              {errors.start_command && (
                <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                  {errors.start_command.message}
                </p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="edit-application-environment" className="text-sm font-medium">
                Environment
              </label>
              <input
                id="edit-application-environment"
                {...register("environment")}
                placeholder="e.g. production"
                className={inputClass}
                style={{ borderColor: errors.environment ? "var(--danger)" : "var(--border)" }}
                disabled={updateApplication.isPending}
                aria-invalid={Boolean(errors.environment)}
              />
              {errors.environment && (
                <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                  {errors.environment.message}
                </p>
              )}
            </div>

            <div>
              <label htmlFor="edit-application-status" className="text-sm font-medium">
                Status
              </label>
              <select
                id="edit-application-status"
                {...register("status")}
                className={inputClass}
                style={{ borderColor: errors.status ? "var(--danger)" : "var(--border)" }}
                disabled={updateApplication.isPending}
                aria-invalid={Boolean(errors.status)}
              >
                {STATUS_OPTIONS.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
              {errors.status && (
                <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                  {errors.status.message}
                </p>
              )}
            </div>
          </div>

          <div>
            <label htmlFor="edit-application-description" className="text-sm font-medium">
              Description
            </label>
            <textarea
              id="edit-application-description"
              {...register("description")}
              placeholder="Describe the responsibility of this application."
              rows={4}
              className={inputClass}
              style={{ borderColor: errors.description ? "var(--danger)" : "var(--border)" }}
              disabled={updateApplication.isPending}
              aria-invalid={Boolean(errors.description)}
            />
            {errors.description && (
              <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
                {errors.description.message}
              </p>
            )}
          </div>

          {submitError ? (
            <p id="edit-application-submit-error" role="alert" aria-live="assertive" className="text-sm" style={{ color: "var(--danger)" }}>
              {submitError}
            </p>
          ) : null}
        </div>

        <footer className="flex justify-end gap-3 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={updateApplication.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={updateApplication.isPending} disabled={updateApplication.isPending}>
            Save changes
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function getApplicationErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to update application. Please try again.";
  }
  return "Unable to update application. Please try again.";
}
