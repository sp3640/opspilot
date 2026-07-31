"use client";

import { ChevronRight, Layers3, Users } from "lucide-react";

import { getProjectIcon } from "./project-icon";
import { ProjectEnvironmentBadge, ProjectHealthBadge } from "./project-status";
import type { Project } from "./types";

type ProjectTableProps = { projects: Project[]; onOpen: (projectID: string) => void };

/** Dense, keyboard-operable representation for API-backed project portfolios. */
export function ProjectTable({ projects, onOpen }: ProjectTableProps) {
  return (
    <div className="overflow-x-auto rounded-2xl border" style={{ borderColor: "var(--border)" }}>
      <table className="w-full min-w-[800px] text-left text-sm">
        <caption className="sr-only">Projects list with health, environment, ownership, and activity</caption>
        <thead
          className="text-xs uppercase tracking-[0.1em]"
          style={{
            color: "var(--muted-foreground)",
            backgroundColor: "color-mix(in srgb, var(--muted) 55%, transparent)",
          }}
        >
          <tr>
            <th scope="col" className="px-5 py-3 font-semibold">Project</th>
            <th scope="col" className="px-5 py-3 font-semibold">Health</th>
            <th scope="col" className="px-5 py-3 font-semibold">Environment</th>
            <th scope="col" className="px-5 py-3 font-semibold">Owner</th>
            <th scope="col" className="px-5 py-3 font-semibold">Services</th>
            <th scope="col" className="px-5 py-3 font-semibold">Updated</th>
            <th scope="col" className="w-12 px-5 py-3"><span className="sr-only">Open</span></th>
          </tr>
        </thead>
        <tbody>
          {projects.map((project) => {
            const Icon = getProjectIcon("folder-kanban");

            return (
              <tr
                key={project.id}
                role="button"
                tabIndex={0}
                aria-label={`Open project details for ${project.name}`}
                onClick={() => onOpen(project.id)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    onOpen(project.id);
                  }
                }}
                className="cursor-pointer border-t transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
                style={{ borderColor: "var(--border)" }}
              >
                <td className="px-5 py-4">
                  <div className="flex items-center gap-3">
                    <span
                      className="rounded-xl p-2"
                      style={{
                        color: "var(--primary)",
                        backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)",
                      }}
                    >
                      <Icon aria-hidden="true" className="h-4 w-4" />
                    </span>
                    <div>
                      <p className="font-medium">{project.name}</p>
                      <p className="mt-0.5 max-w-60 truncate text-xs" style={{ color: "var(--muted-foreground)" }}>
                        {project.description}
                      </p>
                    </div>
                  </div>
                </td>
                <td className="px-5 py-4"><ProjectHealthBadge health={project.health} /></td>
                <td className="px-5 py-4"><ProjectEnvironmentBadge environment={project.environment} /></td>
                <td className="px-5 py-4">{project.owner.name}</td>
                <td className="px-5 py-4">
                  <span className="inline-flex items-center gap-1.5">
                    <Layers3 aria-hidden="true" className="h-3.5 w-3.5" style={{ color: "var(--muted-foreground)" }} />
                    {project.services}
                    <Users aria-hidden="true" className="ml-2 h-3.5 w-3.5" style={{ color: "var(--muted-foreground)" }} />
                    {project.members}
                  </span>
                </td>
                <td className="px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                  {formatDate(project.updatedAt)}
                </td>
                <td className="px-5 py-4">
                  <ChevronRight aria-hidden="true" className="h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
