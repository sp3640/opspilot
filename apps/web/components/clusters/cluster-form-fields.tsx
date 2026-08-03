"use client";

import type { FieldErrors, UseFormRegister } from "react-hook-form";

import {
  CLUSTER_CONNECTION_TYPE_VALUES,
  CLUSTER_PROVIDER_LABELS,
  CLUSTER_PROVIDER_VALUES,
  CLUSTER_STATUS_LABELS,
  CLUSTER_STATUS_VALUES,
} from "@/lib/constants/cluster";
import type { ClusterFormInput } from "@/lib/validation/cluster";
import { ProjectSelect } from "@/components/incidents/project-select";

type ClusterFormFieldsProps = {
  register: UseFormRegister<ClusterFormInput>;
  errors: FieldErrors<ClusterFormInput>;
  disabled: boolean;
  queryEnabled?: boolean;
  currentProjectId?: string;
};

const inputClass =
  "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

export function ClusterFormFields({
  register,
  errors,
  disabled,
  queryEnabled = true,
  currentProjectId,
}: ClusterFormFieldsProps) {
  return (
    <>
      <div>
        <label htmlFor="cluster-name" className="text-sm font-medium">
          Cluster name
        </label>
        <input
          id="cluster-name"
          {...register("name")}
          placeholder="e.g. Production East Cluster"
          className={inputClass}
          style={{ borderColor: errors.name ? "var(--danger)" : "var(--border)" }}
          disabled={disabled}
          aria-invalid={Boolean(errors.name)}
        />
        {errors.name ? <FieldError message={errors.name.message} /> : null}
      </div>

      <div>
        <label htmlFor="cluster-project" className="text-sm font-medium">
          Project
        </label>
        <ProjectSelect
          id="cluster-project"
          queryEnabled={queryEnabled}
          currentProjectId={currentProjectId}
          {...register("project_id")}
          className={inputClass}
          style={{ borderColor: errors.project_id ? "var(--danger)" : "var(--border)" }}
          disabled={disabled}
          aria-invalid={Boolean(errors.project_id)}
        />
        {errors.project_id ? <FieldError message={errors.project_id.message} /> : null}
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <SelectField
          id="cluster-provider"
          label="Provider"
          disabled={disabled}
          errorMessage={errors.provider?.message}
          register={register("provider")}
          options={CLUSTER_PROVIDER_VALUES.map((value) => ({ value, label: CLUSTER_PROVIDER_LABELS[value] }))}
        />

        <SelectField
          id="cluster-status"
          label="Status"
          disabled={disabled}
          errorMessage={errors.status?.message}
          register={register("status")}
          options={CLUSTER_STATUS_VALUES.map((value) => ({ value, label: CLUSTER_STATUS_LABELS[value] }))}
        />
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <SelectField
          id="cluster-connection-type"
          label="Connection type"
          disabled={disabled}
          errorMessage={errors.connection_type?.message}
          register={register("connection_type")}
          options={CLUSTER_CONNECTION_TYPE_VALUES.map((value) => ({ value, label: value }))}
        />

        <div>
          <label htmlFor="cluster-region" className="text-sm font-medium">
            Region
          </label>
          <input
            id="cluster-region"
            {...register("region")}
            placeholder="e.g. us-east-1"
            className={inputClass}
            style={{ borderColor: errors.region ? "var(--danger)" : "var(--border)" }}
            disabled={disabled}
            aria-invalid={Boolean(errors.region)}
          />
          {errors.region ? <FieldError message={errors.region.message} /> : null}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label htmlFor="cluster-api-endpoint" className="text-sm font-medium">
            API endpoint
          </label>
          <input
            id="cluster-api-endpoint"
            {...register("api_endpoint")}
            placeholder="https://api.cluster.example"
            className={inputClass}
            style={{ borderColor: errors.api_endpoint ? "var(--danger)" : "var(--border)" }}
            disabled={disabled}
            aria-invalid={Boolean(errors.api_endpoint)}
          />
          {errors.api_endpoint ? <FieldError message={errors.api_endpoint.message} /> : null}
        </div>

        <div>
          <label htmlFor="cluster-version" className="text-sm font-medium">
            Kubernetes version
          </label>
          <input
            id="cluster-version"
            {...register("version")}
            placeholder="e.g. v1.30.2"
            className={inputClass}
            style={{ borderColor: errors.version ? "var(--danger)" : "var(--border)" }}
            disabled={disabled}
            aria-invalid={Boolean(errors.version)}
          />
          {errors.version ? <FieldError message={errors.version.message} /> : null}
        </div>
      </div>

      <div>
        <label htmlFor="cluster-kubeconfig" className="text-sm font-medium">
          Kubeconfig (encrypted string)
        </label>
        <textarea
          id="cluster-kubeconfig"
          {...register("kubeconfig_encrypted")}
          rows={4}
          placeholder="Paste encrypted kubeconfig"
          className={inputClass}
          style={{ borderColor: errors.kubeconfig_encrypted ? "var(--danger)" : "var(--border)" }}
          disabled={disabled}
          aria-invalid={Boolean(errors.kubeconfig_encrypted)}
        />
        {errors.kubeconfig_encrypted ? <FieldError message={errors.kubeconfig_encrypted.message} /> : null}
      </div>

      <div>
        <label htmlFor="cluster-metadata" className="text-sm font-medium">
          Metadata (JSON)
        </label>
        <textarea
          id="cluster-metadata"
          {...register("metadataText")}
          rows={4}
          placeholder="{}"
          className={inputClass}
          style={{ borderColor: errors.metadataText ? "var(--danger)" : "var(--border)" }}
          disabled={disabled}
          aria-invalid={Boolean(errors.metadataText)}
        />
        {errors.metadataText ? <FieldError message={errors.metadataText.message} /> : null}
      </div>
    </>
  );
}

function SelectField({
  id,
  label,
  register,
  options,
  disabled,
  errorMessage,
}: {
  id: string;
  label: string;
  register: ReturnType<UseFormRegister<ClusterFormInput>>;
  options: Array<{ value: string; label: string }>;
  disabled: boolean;
  errorMessage?: string;
}) {
  return (
    <div>
      <label htmlFor={id} className="text-sm font-medium">
        {label}
      </label>
      <select
        id={id}
        {...register}
        className={inputClass}
        style={{ borderColor: errorMessage ? "var(--danger)" : "var(--border)" }}
        disabled={disabled}
        aria-invalid={Boolean(errorMessage)}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {errorMessage ? <FieldError message={errorMessage} /> : null}
    </div>
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
