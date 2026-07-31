"use client";

import { forwardRef } from "react";
import type { SelectHTMLAttributes } from "react";

import { useProjects } from "@/hooks/use-projects";
import { PROJECT_LOOKUP_QUERY } from "@/lib/constants";

type ProjectSelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  currentProjectId?: string;
  queryEnabled?: boolean;
};

/**
 * Project selector backed by the projects query.
 * It displays project names while preserving project IDs as option values.
 */
export const ProjectSelect = forwardRef<HTMLSelectElement, ProjectSelectProps>(function ProjectSelect(
  { currentProjectId, queryEnabled = true, disabled, ...props },
  ref
) {
  const { data, isLoading, isFetching, isError } = useProjects({
    page: PROJECT_LOOKUP_QUERY.page,
    limit: PROJECT_LOOKUP_QUERY.limit,
    sort: PROJECT_LOOKUP_QUERY.sort,
    order: PROJECT_LOOKUP_QUERY.order,
  }, queryEnabled);

  const projects = data?.items ?? [];
  const hasProjects = projects.length > 0;
  const isBusy = isLoading || isFetching;
  const showCurrentProjectFallback =
    Boolean(currentProjectId) && !projects.some((project) => project.id === currentProjectId);
  const helperMessageId = props.id ? `${props.id}-project-select-helper` : undefined;
  const errorMessageId = props.id ? `${props.id}-project-select-error` : undefined;

  const describedBy = [
    props["aria-describedby"],
    !isBusy && !isError && !hasProjects && !showCurrentProjectFallback ? helperMessageId : undefined,
    isError ? errorMessageId : undefined,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div>
      <select
        ref={ref}
        {...props}
        style={{
          backgroundColor: "var(--card)",
          color: "var(--foreground)",
          ...(props.style ?? {}),
        }}
        disabled={disabled || isBusy || (!hasProjects && !showCurrentProjectFallback)}
        aria-busy={isBusy}
        aria-describedby={describedBy || undefined}
      >
        {isBusy ? (
          <option value="" style={{ backgroundColor: "var(--card)", color: "var(--foreground)" }}>
            Loading projects...
          </option>
        ) : isError ? (
          <option value="" style={{ backgroundColor: "var(--card)", color: "var(--foreground)" }}>
            Unable to load projects
          </option>
        ) : (
          <>
            <option value="" style={{ backgroundColor: "var(--card)", color: "var(--foreground)" }}>
              Select a project
            </option>
            {showCurrentProjectFallback ? (
              <option
                value={currentProjectId}
                style={{ backgroundColor: "var(--card)", color: "var(--foreground)" }}
              >
                Current project (unavailable)
              </option>
            ) : null}
            {projects.map((project) => (
              <option
                key={project.id}
                value={project.id}
                style={{ backgroundColor: "var(--card)", color: "var(--foreground)" }}
              >
                {project.name}
              </option>
            ))}
          </>
        )}
      </select>

      {!isBusy && !isError && !hasProjects && !showCurrentProjectFallback ? (
        <p
          id={helperMessageId}
          aria-live="polite"
          className="mt-1.5 text-xs"
          style={{ color: "var(--muted-foreground)" }}
        >
          No projects found. Create a project first.
        </p>
      ) : null}

      {isError ? (
        <p
          id={errorMessageId}
          role="status"
          aria-live="polite"
          className="mt-1.5 text-xs"
          style={{ color: "var(--muted-foreground)" }}
        >
          Unable to load projects right now.
        </p>
      ) : null}
    </div>
  );
});
