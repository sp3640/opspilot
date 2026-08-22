"use client";

import { forwardRef, useMemo } from "react";
import type { SelectHTMLAttributes } from "react";

import { useClusters } from "@/hooks/use-clusters";

type ClusterSelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  projectId: string;
  queryEnabled?: boolean;
};

/** Target cluster selector, scoped to the deployment's chosen project. */
export const ClusterSelect = forwardRef<HTMLSelectElement, ClusterSelectProps>(function ClusterSelect(
  { projectId, queryEnabled = true, disabled, ...props },
  ref
) {
  const params = useMemo(() => ({ page: 1, limit: 100, projectId: projectId || undefined }), [projectId]);
  const { data, isLoading, isFetching, isError } = useClusters(params, queryEnabled && Boolean(projectId));

  const clusters = data?.items ?? [];
  const isBusy = isLoading || isFetching;

  return (
    <select
      ref={ref}
      {...props}
      style={{ backgroundColor: "var(--card)", color: "var(--foreground)", ...(props.style ?? {}) }}
      disabled={disabled || !projectId || isBusy}
    >
      <option value="">Select a cluster</option>
      {isError ? (
        <option value="" disabled>Unable to load clusters</option>
      ) : (
        clusters.map((cluster) => (
          <option key={cluster.id} value={cluster.id}>{cluster.name}</option>
        ))
      )}
    </select>
  );
});
