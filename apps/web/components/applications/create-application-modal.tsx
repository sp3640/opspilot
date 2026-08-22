"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { Rocket, X } from "lucide-react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { useCreateApplication } from "@/hooks/use-applications";
import { APPLICATION_ENVIRONMENT_VALUES, applicationSchema, type ApplicationFormValues } from "@/lib/validation/application";
import type { CreateApplicationRequest } from "@/types/application-api";

type CreateApplicationModalProps = {
  open: boolean;
  projectId: string;
  onClose: () => void;
};

const RUNTIME_OPTIONS = ["NodeJS", "Go", "Java", "Python", "DotNet", "Static", "Docker"] as const;
const STATUS_OPTIONS = ["Draft", "Ready", "Archived"] as const;

const inputClass =
  "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

export function CreateApplicationModal({ open, projectId, onClose }: CreateApplicationModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const createApplication = useCreateApplication();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ApplicationFormValues>({
    resolver: zodResolver(applicationSchema),
    defaultValues: {
      name: "",
      slug: "",
      description: "",
      repository_url: "",
      default_branch: "",
      runtime: "NodeJS",
      build_command: "",
      start_command: "",
      environment: "",
      status: "Draft",
    },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (createApplication.isPending) return;
    reset();
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: ApplicationFormValues) => {
    setSubmitError(null);

    const payload: CreateApplicationRequest = {
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
      await createApplication.mutateAsync({ projectId, payload });
      reset();
      onClose();
    } catch (error) {
      setSubmitError(getApplicationErrorMessage(error));
    }
  };

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="create-application-title"
      aria-describedby={submitError ? "create-application-description create-application-submit-error" : "create-application-description"}
      className="m-auto w-[calc(100%-2rem)] max-w-2xl rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]"
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
              <Rocket aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="create-application-title" className="text-lg font-semibold">
                Create application
              </h2>
              <p id="create-application-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Define a service and its runtime configuration.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={createApplication.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close create application dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          <div>
            <label htmlFor="application-name" className="text-sm font-medium">
              Name
            </label>
            <input
              id="application-name"
              {...register("name")}
              placeholder="e.g. Customer API"
              className={inputClass}
              style={{ borderColor: errors.name ? "var(--danger)" : "var(--border)" }}
              disabled={createApplication.isPending}
              aria-invalid={Boolean(errors.name)}
            />
            {errors.name ? <FieldError message={errors.name.message} /> : null}
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="application-runtime" className="text-sm font-medium">
                Runtime
              </label>
              <select
                id="application-runtime"
                {...register("runtime")}
                className={inputClass}
                style={{ borderColor: errors.runtime ? "var(--danger)" : "var(--border)" }}
                disabled={createApplication.isPending}
                aria-invalid={Boolean(errors.runtime)}
              >
                {RUNTIME_OPTIONS.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
              {errors.runtime ? <FieldError message={errors.runtime.message} /> : null}
            </div>

            <div>
              <label htmlFor="application-port" className="text-sm font-medium">
                Port
              </label>
              <input
                id="application-port"
                type="number"
                min={1}
                max={65535}
                {...register("port", { valueAsNumber: true })}
                placeholder="e.g. 3000"
                className={inputClass}
                style={{ borderColor: errors.port ? "var(--danger)" : "var(--border)" }}
                disabled={createApplication.isPending}
                aria-invalid={Boolean(errors.port)}
              />
              {errors.port ? <FieldError message={errors.port.message} /> : null}
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="application-slug" className="text-sm font-medium">
                Slug
              </label>
              <input
                id="application-slug"
                {...register("slug")}
                placeholder="e.g. customer-api"
                className={inputClass}
                style={{ borderColor: errors.slug ? "var(--danger)" : "var(--border)" }}
                disabled={createApplication.isPending}
                aria-invalid={Boolean(errors.slug)}
              />
              {errors.slug ? <FieldError message={errors.slug.message} /> : null}
            </div>

            <div>
              <label htmlFor="application-default-branch" className="text-sm font-medium">
                Default branch
              </label>
              <input
                id="application-default-branch"
                {...register("default_branch")}
                placeholder="e.g. main"
                className={inputClass}
                style={{ borderColor: errors.default_branch ? "var(--danger)" : "var(--border)" }}
                disabled={createApplication.isPending}
                aria-invalid={Boolean(errors.default_branch)}
              />
              {errors.default_branch ? <FieldError message={errors.default_branch.message} /> : null}
            </div>
          </div>

          <div>
            <label htmlFor="application-repository-url" className="text-sm font-medium">
              Repository URL
            </label>
            <input
              id="application-repository-url"
              {...register("repository_url")}
              placeholder="https://github.com/org/repo"
              className={inputClass}
              style={{ borderColor: errors.repository_url ? "var(--danger)" : "var(--border)" }}
              disabled={createApplication.isPending}
              aria-invalid={Boolean(errors.repository_url)}
            />
            {errors.repository_url ? <FieldError message={errors.repository_url.message} /> : null}
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="application-build-command" className="text-sm font-medium">
                Build command
              </label>
              <input
                id="application-build-command"
                {...register("build_command")}
                placeholder="e.g. npm run build"
                className={inputClass}
                style={{ borderColor: errors.build_command ? "var(--danger)" : "var(--border)" }}
                disabled={createApplication.isPending}
                aria-invalid={Boolean(errors.build_command)}
              />
              {errors.build_command ? <FieldError message={errors.build_command.message} /> : null}
            </div>

            <div>
              <label htmlFor="application-start-command" className="text-sm font-medium">
                Start command
              </label>
              <input
                id="application-start-command"
                {...register("start_command")}
                placeholder="e.g. npm run start"
                className={inputClass}
                style={{ borderColor: errors.start_command ? "var(--danger)" : "var(--border)" }}
                disabled={createApplication.isPending}
                aria-invalid={Boolean(errors.start_command)}
              />
              {errors.start_command ? <FieldError message={errors.start_command.message} /> : null}
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="application-environment" className="text-sm font-medium">
                Environment
              </label>
              <select
                id="application-environment"
                {...register("environment")}
                className={inputClass}
                style={{ borderColor: errors.environment ? "var(--danger)" : "var(--border)" }}
                disabled={createApplication.isPending}
                aria-invalid={Boolean(errors.environment)}
              >
                {APPLICATION_ENVIRONMENT_VALUES.map((option) => (
                  <option key={option} value={option}>
                    {option || "Not set"}
                  </option>
                ))}
              </select>
              {errors.environment ? <FieldError message={errors.environment.message} /> : null}
            </div>

            <div>
              <label htmlFor="application-status" className="text-sm font-medium">
                Status
              </label>
              <select
                id="application-status"
                {...register("status")}
                className={inputClass}
                style={{ borderColor: errors.status ? "var(--danger)" : "var(--border)" }}
                disabled={createApplication.isPending}
                aria-invalid={Boolean(errors.status)}
              >
                {STATUS_OPTIONS.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
              {errors.status ? <FieldError message={errors.status.message} /> : null}
            </div>
          </div>

          <div>
            <label htmlFor="application-description" className="text-sm font-medium">
              Description
            </label>
            <textarea
              id="application-description"
              {...register("description")}
              placeholder="Describe the responsibility of this application."
              rows={4}
              className={inputClass}
              style={{ borderColor: errors.description ? "var(--danger)" : "var(--border)" }}
              disabled={createApplication.isPending}
              aria-invalid={Boolean(errors.description)}
            />
            {errors.description ? <FieldError message={errors.description.message} /> : null}
          </div>

          {submitError ? (
            <p id="create-application-submit-error" role="alert" aria-live="assertive" className="text-sm" style={{ color: "var(--danger)" }}>
              {submitError}
            </p>
          ) : null}
        </div>

        <footer className="flex justify-end gap-3 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={createApplication.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={createApplication.isPending} disabled={createApplication.isPending}>
            Create
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function FieldError({ message }: { message?: string }) {
  if (!message) return null;

  return (
    <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>
      {message}
    </p>
  );
}

function getApplicationErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to create application. Please try again.";
  }
  return "Unable to create application. Please try again.";
}
