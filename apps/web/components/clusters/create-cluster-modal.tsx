"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { Server, X } from "lucide-react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { useCreateCluster } from "@/hooks/use-clusters";
import {
  CLUSTER_DEFAULT_CONNECTION_TYPE,
  CLUSTER_DEFAULT_PROVIDER,
  CLUSTER_DEFAULT_STATUS,
} from "@/lib/constants/cluster";
import { clusterFormSchema, type ClusterFormInput } from "@/lib/validation/cluster";
import type { ClusterResponse, CreateClusterRequest } from "@/types/cluster-api";

import { ClusterFormFields } from "./cluster-form-fields";

type CreateClusterModalProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (cluster: ClusterResponse) => void;
};

export function CreateClusterModal({ open, onClose, onCreated }: CreateClusterModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const createCluster = useCreateCluster();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
    setError,
  } = useForm<ClusterFormInput>({
    resolver: zodResolver(clusterFormSchema),
    defaultValues: {
      project_id: "",
      name: "",
      provider: CLUSTER_DEFAULT_PROVIDER,
      status: CLUSTER_DEFAULT_STATUS,
      connection_type: CLUSTER_DEFAULT_CONNECTION_TYPE,
      kubeconfig_encrypted: "",
      api_endpoint: "",
      region: "",
      version: "",
      validation_error: "",
      metadataText: "{}",
    },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (createCluster.isPending) return;
    reset();
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: ClusterFormInput) => {
    setSubmitError(null);

    const metadata = parseMetadata(input.metadataText);
    if (!metadata.ok) {
      setError("metadataText", { message: "Metadata must be valid JSON." });
      return;
    }

    const payload: CreateClusterRequest = {
      project_id: input.project_id,
      name: input.name,
      provider: input.provider,
      status: input.status,
      connection_type: input.connection_type,
      kubeconfig_encrypted: input.kubeconfig_encrypted,
      api_endpoint: input.api_endpoint,
      region: input.region,
      version: input.version,
      validation_error: input.validation_error,
      metadata: metadata.value,
    };

    try {
      const cluster = await createCluster.mutateAsync(payload);
      reset();
      onCreated(cluster);
      onClose();
    } catch (error) {
      setSubmitError(getClusterErrorMessage(error));
    }
  };

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="create-cluster-title"
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
              <Server aria-hidden="true" className="h-5 w-5" />
            </div>
            <div>
              <h2 id="create-cluster-title" className="text-lg font-semibold">
                Create cluster
              </h2>
              <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Register a cluster and attach it to a project.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={createCluster.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close create cluster dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          <ClusterFormFields register={register} errors={errors} disabled={createCluster.isPending} queryEnabled={open} />

          {submitError ? (
            <p role="alert" aria-live="assertive" className="text-sm" style={{ color: "var(--danger)" }}>
              {submitError}
            </p>
          ) : null}
        </div>

        <footer className="flex justify-end gap-3 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={createCluster.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={createCluster.isPending} disabled={createCluster.isPending}>
            Create
          </Button>
        </footer>
      </form>
    </dialog>
  );
}

function parseMetadata(text: string | undefined):
  | { ok: true; value: Record<string, unknown> }
  | { ok: false } {
  const trimmed = (text ?? "{}").trim();
  if (!trimmed) return { ok: true, value: {} };

  try {
    const parsed = JSON.parse(trimmed) as unknown;
    if (typeof parsed === "object" && parsed !== null && !Array.isArray(parsed)) {
      return { ok: true, value: parsed as Record<string, unknown> };
    }
    return { ok: false };
  } catch {
    return { ok: false };
  }
}

function getClusterErrorMessage(error: unknown) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? "Unable to create cluster. Please try again.";
  }
  return "Unable to create cluster. Please try again.";
}
