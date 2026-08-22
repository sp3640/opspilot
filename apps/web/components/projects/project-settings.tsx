"use client";

import { Pencil, Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { ProjectResponse } from "@/types/project-api";

/**
 * Only exposes settings the backend actually supports mutating today:
 * name + description (via PUT /projects/:id) and delete (via DELETE
 * /projects/:id). Environment/health/owner have no update endpoint, so
 * they're shown read-only rather than as editable "fake" settings.
 */
export function ProjectSettings({
  project,
  isAdmin,
  onEdit,
  onDelete,
}: {
  project: ProjectResponse;
  isAdmin: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="space-y-6">
      <section className="rounded-2xl border p-5" style={{ borderColor: "var(--border)" }}>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h3 className="text-sm font-semibold">Project identity</h3>
            <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
              Name and description shown across the app.
            </p>
          </div>
          {isAdmin ? (
            <Button type="button" variant="secondary" onClick={onEdit} className="px-3 shrink-0">
              <Pencil aria-hidden="true" className="h-4 w-4" />
              Edit
            </Button>
          ) : null}
        </div>

        <dl className="mt-5 space-y-4">
          <div>
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Name</dt>
            <dd className="mt-1 text-sm font-medium">{project.name}</dd>
          </div>
          <div>
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Description</dt>
            <dd className="mt-1 text-sm">{project.description || "No description provided"}</dd>
          </div>
          <div>
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Slug</dt>
            <dd className="mt-1 break-all text-sm" style={{ color: "var(--muted-foreground)" }}>{project.slug}</dd>
          </div>
        </dl>

        {!isAdmin ? (
          <p className="mt-5 text-xs" style={{ color: "var(--muted-foreground)" }}>
            Only Platform Admins can edit project settings.
          </p>
        ) : null}
      </section>

      <section className="rounded-2xl border p-5" style={{ borderColor: "var(--border)" }}>
        <h3 className="text-sm font-semibold">Read-only attributes</h3>
        <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
          Set automatically and not currently editable via the API.
        </p>
        <dl className="mt-5 grid grid-cols-2 gap-4">
          <div>
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Environment</dt>
            <dd className="mt-1 text-sm font-medium">{project.environment}</dd>
          </div>
          <div>
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Health</dt>
            <dd className="mt-1 text-sm font-medium">{project.health}</dd>
          </div>
          <div>
            <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>Owner</dt>
            <dd className="mt-1 text-sm font-medium">{project.owner.name}</dd>
          </div>
        </dl>
      </section>

      {isAdmin ? (
        <section
          className="rounded-2xl border p-5"
          style={{ borderColor: "var(--danger)", backgroundColor: "color-mix(in srgb, var(--danger) 6%, transparent)" }}
        >
          <h3 className="text-sm font-semibold" style={{ color: "var(--danger)" }}>Danger zone</h3>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
            Permanently delete this project. This action cannot be undone.
          </p>
          <Button type="button" variant="danger" onClick={onDelete} className="mt-4 px-3">
            <Trash2 aria-hidden="true" className="h-4 w-4" />
            Delete project
          </Button>
        </section>
      ) : null}
    </div>
  );
}
