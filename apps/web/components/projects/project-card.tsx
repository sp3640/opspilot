"use client";

import { Ellipsis, ExternalLink, Layers3, Users } from "lucide-react";

import { Button } from "@/components/ui/button";

import { getProjectIcon } from "./project-icon";
import { ProjectEnvironmentBadge, ProjectHealthBadge } from "./project-status";
import type { Project } from "./types";

type ProjectCardProps = { project: Project; onOpen: (projectID: string) => void };

/** Scannable project summary using fields returned by the projects API. */
export function ProjectCard({ project, onOpen }: ProjectCardProps) {
  const Icon = getProjectIcon("folder-kanban");

  return (
    <article
      onClick={() => onOpen(project.id)}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onOpen(project.id);
        }
      }}
      role="button"
      tabIndex={0}
      aria-label={`Open project details for ${project.name}`}
      className="group cursor-pointer rounded-3xl border p-5 shadow-[var(--shadow-sm)] transition-all duration-300 hover:-translate-y-1 hover:shadow-[var(--shadow-md)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <div
            className="rounded-2xl p-3"
            style={{
              color: "var(--primary)",
              backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)",
            }}
          >
            <Icon aria-hidden="true" className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <h2 className="truncate font-semibold tracking-tight">{project.name}</h2>
            <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
              Owned by {project.owner.name}
            </p>
          </div>
        </div>
        <Button
          type="button"
          variant="ghost"
          onClick={(event) => {
            event.stopPropagation();
            onOpen(project.id);
          }}
          className="h-9 w-9 rounded-xl p-0"
          aria-label={`Open ${project.name} actions`}
        >
          <Ellipsis aria-hidden="true" className="h-4 w-4" />
        </Button>
      </div>

      <p className="mt-5 line-clamp-2 min-h-10 text-sm leading-5" style={{ color: "var(--muted-foreground)" }}>
        {project.description}
      </p>

      <div className="mt-5 flex flex-wrap gap-2">
        <ProjectHealthBadge health={project.health} />
        <ProjectEnvironmentBadge environment={project.environment} />
      </div>

      <div className="mt-5 grid grid-cols-2 gap-3 border-y py-4 text-sm" style={{ borderColor: "var(--border)" }}>
        <div>
          <span className="flex items-center gap-1.5 text-xs" style={{ color: "var(--muted-foreground)" }}>
            <Layers3 aria-hidden="true" className="h-3.5 w-3.5" />
            Services
          </span>
          <p className="mt-1 font-semibold">{project.services}</p>
        </div>
        <div>
          <span className="flex items-center gap-1.5 text-xs" style={{ color: "var(--muted-foreground)" }}>
            <Users aria-hidden="true" className="h-3.5 w-3.5" />
            Members
          </span>
          <p className="mt-1 font-semibold">{project.members}</p>
        </div>
      </div>

      <div className="mt-4 flex items-center justify-between gap-3">
        <span className="inline-flex items-center gap-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Slug {project.slug}
        </span>
        <span className="inline-flex items-center gap-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Updated {formatDate(project.updatedAt)}
          <ExternalLink aria-hidden="true" className="h-3 w-3 opacity-0 transition-opacity group-hover:opacity-100" />
        </span>
      </div>
    </article>
  );
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
