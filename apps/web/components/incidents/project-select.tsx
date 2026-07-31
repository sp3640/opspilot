"use client";

import { forwardRef } from "react";
import type { SelectHTMLAttributes } from "react";

import { useProjects } from "@/hooks/use-projects";

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
    page: 1,
    limit: 100,
    sort: "name",
    order: "asc",
  }, queryEnabled);

  const projects = data?.items ?? [];
  const hasProjects = projects.length > 0;
  const isBusy = isLoading || isFetching;
  const showCurrentProjectFallback =
    Boolean(currentProjectId) && !projects.some((project) => project.id === currentProjectId);

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
        <p className="mt-1.5 text-xs" style={{ color: "var(--muted-foreground)" }}>
          No projects found. Create a project first.
        </p>
      ) : null}

      {isError ? (
        <p className="mt-1.5 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Unable to load projects right now.
        </p>
      ) : null}
    </div>
  );
});
