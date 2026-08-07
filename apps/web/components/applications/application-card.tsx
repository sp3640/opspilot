"use client";

import { ExternalLink, Pencil, Rocket, Trash2 } from "lucide-react";

import { StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";
import type { ApplicationResponse } from "@/types/application-api";

type ApplicationCardProps = {
  application: ApplicationResponse;
  onClick?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
};

/** Scannable application summary using fields returned by the applications API. */
export function ApplicationCard({ application, onClick, onEdit, onDelete }: ApplicationCardProps) {
  return (
    <article
      onClick={() => onClick?.()}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onClick?.();
        }
      }}
      role="button"
      tabIndex={0}
      aria-label={`Open application details for ${application.name}`}
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
            <Rocket aria-hidden="true" className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <h2 className="truncate font-semibold tracking-tight">{application.name}</h2>
            <p className="mt-1 truncate text-xs" style={{ color: "var(--muted-foreground)" }}>
              {application.repositoryUrl || "No repository configured"}
            </p>
          </div>
        </div>
        {(onEdit || onDelete) && (
          <div className="flex items-center gap-1">
            {onEdit && (
              <Button
                type="button"
                variant="ghost"
                onClick={(event) => {
                  event.stopPropagation();
                  onEdit();
                }}
                className="h-9 w-9 rounded-xl p-0"
                aria-label={`Edit ${application.name}`}
              >
                <Pencil aria-hidden="true" className="h-4 w-4" />
              </Button>
            )}
            {onDelete && (
              <Button
                type="button"
                variant="ghost"
                onClick={(event) => {
                  event.stopPropagation();
                  onDelete();
                }}
                className="h-9 w-9 rounded-xl p-0"
                aria-label={`Delete ${application.name}`}
              >
                <Trash2 aria-hidden="true" className="h-4 w-4" />
              </Button>
            )}
          </div>
        )}
      </div>

      <div className="mt-5 flex flex-wrap gap-2">
        <StatusBadge variant={getStatusVariant(application.status)}>{application.status}</StatusBadge>
        <span
          className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
          style={{
            color: "var(--muted-foreground)",
            backgroundColor: "color-mix(in srgb, var(--muted-foreground) 12%, transparent)",
          }}
        >
          {application.runtime}
        </span>
      </div>

      <dl className="mt-5 grid grid-cols-2 gap-3 border-y py-4 text-sm" style={{ borderColor: "var(--border)" }}>
        <Meta label="Default branch" value={application.defaultBranch || "-"} />
        <Meta label="Port" value={String(application.port)} />
        {application.environment ? <Meta label="Environment" value={application.environment} /> : null}
      </dl>

      <div className="mt-4 flex items-center justify-between gap-3">
        <span className="inline-flex items-center gap-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Slug {application.slug}
        </span>
        <span className="inline-flex items-center gap-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Updated {formatDate(application.updatedAt)}
          <ExternalLink aria-hidden="true" className="h-3 w-3 opacity-0 transition-opacity group-hover:opacity-100" />
        </span>
      </div>
    </article>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        {label}
      </dt>
      <dd className="mt-1 truncate font-semibold">{value}</dd>
    </div>
  );
}

function getStatusVariant(status: string): "success" | "info" | "archived" {
  switch (status) {
    case "Ready":
      return "success";
    case "Archived":
      return "archived";
    default:
      return "info";
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
