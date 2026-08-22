"use client";

import { forwardRef, useMemo } from "react";
import type { SelectHTMLAttributes } from "react";

import { useApplicationsByProject } from "@/hooks/use-applications";

type ApplicationSelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  projectId: string;
  queryEnabled?: boolean;
};

const LOOKUP_PARAMS = { page: 1, limit: 100 };

/**
 * Affected-application selector, scoped to the incident's chosen project
 * (the backend rejects an application from a different project) - optional,
 * so "no application" is always a valid choice for incidents that aren't
 * scoped to one.
 */
export const ApplicationSelect = forwardRef<HTMLSelectElement, ApplicationSelectProps>(function ApplicationSelect(
  { projectId, queryEnabled = true, disabled, ...props },
  ref
) {
  const params = useMemo(() => LOOKUP_PARAMS, []);
  const { data, isLoading, isFetching, isError } = useApplicationsByProject(
    projectId || null,
    params,
    queryEnabled && Boolean(projectId)
  );

  const applications = data?.items ?? [];
  const isBusy = isLoading || isFetching;

  return (
    <select
      ref={ref}
      {...props}
      style={{ backgroundColor: "var(--card)", color: "var(--foreground)", ...(props.style ?? {}) }}
      disabled={disabled || !projectId || isBusy}
    >
      <option value="">No application</option>
      {isError ? (
        <option value="" disabled>Unable to load applications</option>
      ) : (
        applications.map((application) => (
          <option key={application.id} value={application.id}>{application.name}</option>
        ))
      )}
    </select>
  );
});
