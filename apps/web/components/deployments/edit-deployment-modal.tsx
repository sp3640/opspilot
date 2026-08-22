"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { Pencil, X } from "lucide-react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { Button } from "@/components/ui/button";
import { useUpdateDeployment } from "@/hooks/use-deployments";
import { DEPLOYMENT_ENVIRONMENT_OPTIONS, DEPLOYMENT_STRATEGY_OPTIONS } from "@/lib/constants";
import { editDeploymentSchema } from "@/lib/validation/deployment";
import type { DeploymentEnvironment, DeploymentStrategy } from "@/lib/constants";
import type { DeploymentResponse } from "@/types/deployment-api";
import { ClusterSelect } from "./cluster-select";

type EditDeploymentFormData = z.infer<typeof editDeploymentSchema>;

/** Edits an existing deployment record's configuration and provenance (PATCH /deployments/:id). */
export function EditDeploymentModal({
  open,
  deployment,
  onClose,
}: {
  open: boolean;
  deployment: DeploymentResponse | null;
  onClose: () => void;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateDeployment = useUpdateDeployment();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<EditDeploymentFormData>({
    resolver: zodResolver(editDeploymentSchema) as never,
    values: deployment
      ? {
          image: deployment.image,
          imageTag: deployment.imageTag ?? "",
          environment: deployment.environment as DeploymentEnvironment,
          namespace: deployment.namespace,
          replicaCount: deployment.replicaCount,
          deploymentStrategy: deployment.deploymentStrategy as DeploymentStrategy,
          targetClusterId: deployment.targetClusterId,
          commitSha: deployment.commitSha ?? "",
          author: deployment.author ?? "",
        }
      : undefined,
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (updateDeployment.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: EditDeploymentFormData) => {
    if (!deployment) return;
    setSubmitError(null);
    try {
      await updateDeployment.mutateAsync({
        id: deployment.id,
        payload: {
          targetClusterId: input.targetClusterId,
          image: input.image,
          imageTag: input.imageTag ?? "",
          environment: input.environment,
          namespace: input.namespace,
          replicaCount: input.replicaCount,
          deploymentStrategy: input.deploymentStrategy,
          commitSha: input.commitSha ?? "",
          author: input.author ?? "",
        },
      });
      reset();
      onClose();
    } catch (error) {
      setSubmitError(getDeploymentErrorMessage(error));
    }
  };

  if (!deployment) return null;

  const inputClass =
    "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="edit-deployment-title"
      aria-describedby={submitError ? "edit-deployment-description edit-deployment-submit-error" : "edit-deployment-description"}
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
              <Pencil aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="edit-deployment-title" className="text-lg font-semibold">
                Edit deployment
              </h2>
              <p id="edit-deployment-description" className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Update this deployment&apos;s configuration and provenance.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={updateDeployment.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close edit deployment dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="max-h-[70vh] space-y-5 overflow-y-auto p-5">
          {submitError && (
            <div
              id="edit-deployment-submit-error"
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
              <label htmlFor="edit-deployment-image" className="text-sm font-medium">Image</label>
              <input
                id="edit-deployment-image"
                {...register("image")}
                className={inputClass}
                style={{ borderColor: errors.image ? "var(--danger)" : "var(--border)" }}
                disabled={updateDeployment.isPending}
              />
              {errors.image && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.image.message}</p>}
            </div>

            <div>
              <label htmlFor="edit-deployment-image-tag" className="text-sm font-medium">Image tag</label>
              <input
                id="edit-deployment-image-tag"
                {...register("imageTag")}
                className={inputClass}
                style={{ borderColor: errors.imageTag ? "var(--danger)" : "var(--border)" }}
                disabled={updateDeployment.isPending}
              />
              {errors.imageTag && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.imageTag.message}</p>}
            </div>

            <div>
              <label htmlFor="edit-deployment-replica-count" className="text-sm font-medium">Replica count</label>
              <input
                id="edit-deployment-replica-count"
                type="number"
                min={1}
                {...register("replicaCount")}
                className={inputClass}
                style={{ borderColor: errors.replicaCount ? "var(--danger)" : "var(--border)" }}
                disabled={updateDeployment.isPending}
              />
              {errors.replicaCount && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.replicaCount.message}</p>}
            </div>

            <div>
              <label htmlFor="edit-deployment-environment" className="text-sm font-medium">Environment</label>
              <select
                id="edit-deployment-environment"
                {...register("environment")}
                className={inputClass}
                style={{ borderColor: errors.environment ? "var(--danger)" : "var(--border)" }}
                disabled={updateDeployment.isPending}
              >
                {DEPLOYMENT_ENVIRONMENT_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>{option.label}</option>
                ))}
              </select>
            </div>

            <div>
              <label htmlFor="edit-deployment-strategy" className="text-sm font-medium">Strategy</label>
              <select
                id="edit-deployment-strategy"
                {...register("deploymentStrategy")}
                className={inputClass}
                style={{ borderColor: errors.deploymentStrategy ? "var(--danger)" : "var(--border)" }}
                disabled={updateDeployment.isPending}
              >
                {DEPLOYMENT_STRATEGY_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>{option.label}</option>
                ))}
              </select>
            </div>

            <div className="col-span-2">
              <label htmlFor="edit-deployment-namespace" className="text-sm font-medium">Namespace</label>
              <input
                id="edit-deployment-namespace"
                {...register("namespace")}
                className={inputClass}
                style={{ borderColor: errors.namespace ? "var(--danger)" : "var(--border)" }}
                disabled={updateDeployment.isPending}
              />
              {errors.namespace && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.namespace.message}</p>}
            </div>

            <div className="col-span-2">
              <label htmlFor="edit-deployment-cluster" className="text-sm font-medium">Target cluster</label>
              <ClusterSelect
                id="edit-deployment-cluster"
                projectId={deployment.projectId}
                queryEnabled={open}
                {...register("targetClusterId")}
                className={inputClass}
                style={{ borderColor: errors.targetClusterId ? "var(--danger)" : "var(--border)" }}
                disabled={updateDeployment.isPending}
              />
              {errors.targetClusterId && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.targetClusterId.message}</p>}
            </div>

            <div>
              <label htmlFor="edit-deployment-commit-sha" className="text-sm font-medium">Commit SHA (optional)</label>
              <input
                id="edit-deployment-commit-sha"
                {...register("commitSha")}
                placeholder="a1b2c3d"
                className={inputClass}
                style={{ borderColor: errors.commitSha ? "var(--danger)" : "var(--border)" }}
                disabled={updateDeployment.isPending}
              />
              {errors.commitSha && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.commitSha.message}</p>}
            </div>

            <div>
              <label htmlFor="edit-deployment-author" className="text-sm font-medium">Author (optional)</label>
              <input
                id="edit-deployment-author"
                {...register("author")}
                placeholder="Jane Doe"
                className={inputClass}
                style={{ borderColor: errors.author ? "var(--danger)" : "var(--border)" }}
                disabled={updateDeployment.isPending}
              />
              {errors.author && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.author.message}</p>}
            </div>
          </div>
        </div>

        <footer className="flex gap-2 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={updateDeployment.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={updateDeployment.isPending} disabled={updateDeployment.isPending}>
            Save changes
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function getDeploymentErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to update deployment. Please try again.";
  }
  return "Unable to update deployment. Please try again.";
}
