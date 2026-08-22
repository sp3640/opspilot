"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { Pencil, X } from "lucide-react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { useUpdateCluster } from "@/hooks/use-clusters";
import {
  CLUSTER_DEFAULT_CONNECTION_TYPE,
  CLUSTER_DEFAULT_PROVIDER,
} from "@/lib/constants/cluster";
import { editClusterFormSchema, type ClusterFormInput } from "@/lib/validation/cluster";
import type { ClusterResponse, UpdateClusterRequest } from "@/types/cluster-api";

import { ClusterFormFields } from "./cluster-form-fields";

type EditClusterModalProps = {
  open: boolean;
  cluster: ClusterResponse | null;
  onClose: () => void;
};

export function EditClusterModal({ open, cluster, onClose }: EditClusterModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const updateCluster = useUpdateCluster();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors },
    setError,
  } = useForm<ClusterFormInput>({
    resolver: zodResolver(editClusterFormSchema),
    defaultValues: {
      project_id: cluster?.projectId ?? "",
      name: cluster?.name ?? "",
      provider: (cluster?.provider as ClusterFormInput["provider"]) ?? CLUSTER_DEFAULT_PROVIDER,
      connection_type: (cluster?.connectionType as ClusterFormInput["connection_type"]) ?? CLUSTER_DEFAULT_CONNECTION_TYPE,
      kubeconfig_encrypted: "",
      api_endpoint: cluster?.apiEndpoint ?? "",
      region: cluster?.region ?? "",
      metadataText: cluster?.metadata ? JSON.stringify(cluster.metadata, null, 2) : "{}",
    },
    values: {
      project_id: cluster?.projectId ?? "",
      name: cluster?.name ?? "",
      provider: (cluster?.provider as ClusterFormInput["provider"]) ?? CLUSTER_DEFAULT_PROVIDER,
      connection_type: (cluster?.connectionType as ClusterFormInput["connection_type"]) ?? CLUSTER_DEFAULT_CONNECTION_TYPE,
      kubeconfig_encrypted: "",
      api_endpoint: cluster?.apiEndpoint ?? "",
      region: cluster?.region ?? "",
      metadataText: cluster?.metadata ? JSON.stringify(cluster.metadata, null, 2) : "{}",
    },
  });

  useEffect(() => {
    if (open && !dialogRef.current?.open) dialogRef.current?.showModal();
    if (!open && dialogRef.current?.open) dialogRef.current.close();
  }, [open]);

  const close = () => {
    if (updateCluster.isPending) return;
    setSubmitError(null);
    onClose();
  };

  const submit = async (input: ClusterFormInput) => {
    if (!cluster) return;
    setSubmitError(null);

    const metadata = parseMetadata(input.metadataText);
    if (!metadata.ok) {
      setError("metadataText", { message: "Metadata must be valid JSON." });
      return;
    }

    // kubeconfig_encrypted is only included when the user actually typed a
    // replacement — omitting the key (rather than sending "") tells the
    // backend to leave the stored credential untouched, since it can never
    // be pre-filled here to round-trip in the first place.
    const trimmedKubeconfig = input.kubeconfig_encrypted.trim();
    const payload: UpdateClusterRequest = {
      project_id: input.project_id,
      name: input.name,
      provider: input.provider,
      connection_type: input.connection_type,
      api_endpoint: input.api_endpoint,
      region: input.region,
      metadata: metadata.value,
      ...(trimmedKubeconfig ? { kubeconfig_encrypted: trimmedKubeconfig } : {}),
    };

    try {
      await updateCluster.mutateAsync({ id: cluster.id, payload });
      onClose();
    } catch (error) {
      setSubmitError(getClusterErrorMessage(error));
    }
  };

  if (!cluster) return null;

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      aria-labelledby="edit-cluster-title"
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
              <h2 id="edit-cluster-title" className="text-lg font-semibold">
                Edit cluster
              </h2>
              <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                Update cluster configuration fields.
              </p>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            onClick={close}
            disabled={updateCluster.isPending}
            className="h-9 w-9 rounded-xl p-0"
            aria-label="Close edit cluster dialog"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>

        <div className="space-y-5 p-5">
          <ClusterFormFields
            register={register}
            errors={errors}
            disabled={updateCluster.isPending}
            queryEnabled={open}
            currentProjectId={cluster.projectId}
            mode="edit"
          />

          {submitError ? (
            <p role="alert" aria-live="assertive" className="text-sm" style={{ color: "var(--danger)" }}>
              {submitError}
            </p>
          ) : null}
        </div>

        <footer className="flex justify-end gap-3 border-t p-5" style={{ borderColor: "var(--border)" }}>
          <Button type="button" variant="secondary" onClick={close} disabled={updateCluster.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={updateCluster.isPending} disabled={updateCluster.isPending}>
            Save changes
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
    return error.response?.data?.message ?? "Unable to update cluster. Please try again.";
  }
  return "Unable to update cluster. Please try again.";
}
