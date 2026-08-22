"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { Rocket, X } from "lucide-react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { Button } from "@/components/ui/button";
import { useCreateDeployment } from "@/hooks/use-deployments";
import { DEPLOYMENT_DEFAULT_ENVIRONMENT, DEPLOYMENT_DEFAULT_STRATEGY, DEPLOYMENT_ENVIRONMENT_OPTIONS, DEPLOYMENT_STRATEGY_OPTIONS } from "@/lib/constants";
import { createDeploymentSchema } from "@/lib/validation/deployment";
import { ClusterSelect } from "./cluster-select";

type CreateDeploymentFormData = z.infer<typeof createDeploymentSchema>;

/** Creates a new deployment record for a fixed application/project (POST /deployments). */
export function CreateDeploymentModal({
  open,
  applicationId,
  projectId,
  onClose,
}: {
  open: boolean;
  applicationId: string;
  projectId: string;
  onClose: () => void;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const createDeployment = useCreateDeployment();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateDeploymentFormData>({
    resolver: zodResolver(createDeploymentSchema) as never,
    defaultValues: {
      image: "",
      imageTag: "",
      environment: DEPLOYMENT_DEFAULT_ENVIRONMENT,
      namespace: "",
      replicaCount: 1,
      deploymentStrategy: DEPLOYMENT_DEFAULT_STRATEGY,
      targetClusterId: "",
      commitSha: "",
      author: "",
    },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (createDeployment.isPending) return;
    reset();
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: CreateDeploymentFormData) => {
    setSubmitError(null);
    try {
      await createDeployment.mutateAsync({
        applicationId,
        projectId,
        targetClusterId: input.targetClusterId,
        image: input.image,
        imageTag: input.imageTag || undefined,
        environment: input.environment,
        namespace: input.namespace,
        replicaCount: input.replicaCount,
        deploymentStrategy: input.deploymentStrategy,
        commitSha: input.commitSha || undefined,
        author: input.author || undefined,
      });
      reset();
      onClose();
    } catch (error) {
      setSubmitError(getDeploymentErrorMessage(error));
    }
  };

  const inputClass =
    "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="create-deployment-title"
      aria-describedby={submitError ? "create-deployment-description create-deployment-submit-error" : "create-deployment-description"}
      className="m-auto w-[calc(100%-2rem)] max-w-lg rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <form onSubmit={handleSubmit(submit)}>
        <header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-center gap-3">
            <div
              className="rounded-2xl p-3"
              style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}
            >
              <Rocket aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="create-deployment-title" className="text-lg font-semibold">
                New deployment
              </h2>
              <p id="create-deployment-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Deploy a new version of this application.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={createDeployment.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close new deployment dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="max-h-[70vh] space-y-5 overflow-y-auto p-5">
          {submitError && (
            <div
              id="create-deployment-submit-error"
              role="alert"
              aria-live="assertive"
              className="rounded-2xl border p-3 text-sm"
              style={{ backgroundColor: "color-mix(in srgb, var(--danger) 12%, transparent)", borderColor: "var(--danger)", color: "var(--danger)" }}
            >
              {submitError}
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="col-span-2">
              <label htmlFor="deployment-image" className="text-sm font-medium">Image</label>
              <input
                id="deployment-image"
                {...register("image")}
                placeholder="ghcr.io/opspilot/api"
                className={inputClass}
                style={{ borderColor: errors.image ? "var(--danger)" : "var(--border)" }}
                disabled={createDeployment.isPending}
                aria-invalid={Boolean(errors.image)}
              />
              {errors.image && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.image.message}</p>}
            </div>

            <div>
              <label htmlFor="deployment-image-tag" className="text-sm font-medium">Image tag</label>
              <input
                id="deployment-image-tag"
                {...register("imageTag")}
                placeholder="v1.2.3"
                className={inputClass}
                style={{ borderColor: errors.imageTag ? "var(--danger)" : "var(--border)" }}
                disabled={createDeployment.isPending}
              />
              {errors.imageTag && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.imageTag.message}</p>}
            </div>

            <div>
              <label htmlFor="deployment-replica-count" className="text-sm font-medium">Replica count</label>
              <input
                id="deployment-replica-count"
                type="number"
                min={1}
                {...register("replicaCount")}
                className={inputClass}
                style={{ borderColor: errors.replicaCount ? "var(--danger)" : "var(--border)" }}
                disabled={createDeployment.isPending}
              />
              {errors.replicaCount && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.replicaCount.message}</p>}
            </div>

            <div>
              <label htmlFor="deployment-environment" className="text-sm font-medium">Environment</label>
              <select
                id="deployment-environment"
                {...register("environment")}
                className={inputClass}
                style={{ borderColor: errors.environment ? "var(--danger)" : "var(--border)" }}
                disabled={createDeployment.isPending}
              >
                {DEPLOYMENT_ENVIRONMENT_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>{option.label}</option>
                ))}
              </select>
            </div>

            <div>
              <label htmlFor="deployment-strategy" className="text-sm font-medium">Strategy</label>
              <select
                id="deployment-strategy"
                {...register("deploymentStrategy")}
                className={inputClass}
                style={{ borderColor: errors.deploymentStrategy ? "var(--danger)" : "var(--border)" }}
                disabled={createDeployment.isPending}
              >
                {DEPLOYMENT_STRATEGY_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>{option.label}</option>
                ))}
              </select>
            </div>

            <div className="col-span-2">
              <label htmlFor="deployment-namespace" className="text-sm font-medium">Namespace</label>
              <input
                id="deployment-namespace"
                {...register("namespace")}
                placeholder="production"
                className={inputClass}
                style={{ borderColor: errors.namespace ? "var(--danger)" : "var(--border)" }}
                disabled={createDeployment.isPending}
              />
              {errors.namespace && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.namespace.message}</p>}
            </div>

            <div className="col-span-2">
              <label htmlFor="deployment-cluster" className="text-sm font-medium">Target cluster</label>
              <ClusterSelect
                id="deployment-cluster"
                projectId={projectId}
                queryEnabled={open}
                {...register("targetClusterId")}
                className={inputClass}
                style={{ borderColor: errors.targetClusterId ? "var(--danger)" : "var(--border)" }}
                disabled={createDeployment.isPending}
              />
              {errors.targetClusterId && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.targetClusterId.message}</p>}
            </div>

            <div>
              <label htmlFor="deployment-commit-sha" className="text-sm font-medium">Commit SHA (optional)</label>
              <input
                id="deployment-commit-sha"
                {...register("commitSha")}
                placeholder="a1b2c3d"
                className={inputClass}
                style={{ borderColor: errors.commitSha ? "var(--danger)" : "var(--border)" }}
                disabled={createDeployment.isPending}
              />
              {errors.commitSha && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.commitSha.message}</p>}
            </div>

            <div>
              <label htmlFor="deployment-author" className="text-sm font-medium">Author (optional)</label>
              <input
                id="deployment-author"
                {...register("author")}
                placeholder="Jane Doe"
                className={inputClass}
                style={{ borderColor: errors.author ? "var(--danger)" : "var(--border)" }}
                disabled={createDeployment.isPending}
              />
              {errors.author && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.author.message}</p>}
            </div>
          </div>
        </div>

        <footer className="flex gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={createDeployment.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={createDeployment.isPending} disabled={createDeployment.isPending}>
            Create deployment
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function getDeploymentErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to create deployment. Please try again.";
  }
  return "Unable to create deployment. Please try again.";
}
