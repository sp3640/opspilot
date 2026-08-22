"use client";

import { ExternalLink, Users } from "lucide-react";

import type { Team } from "./types";

type TeamCardProps = { team: Team; onOpen: (teamID: string) => void };

/** Scannable team summary using fields returned by the teams API. */
export function TeamCard({ team, onOpen }: TeamCardProps) {
  return (
    <article
      onClick={() => onOpen(team.id)}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onOpen(team.id);
        }
      }}
      role="button"
      tabIndex={0}
      aria-label={`Open team details for ${team.name}`}
      className="group cursor-pointer rounded-3xl border p-5 shadow-[var(--shadow-sm)] transition-all duration-300 hover:-translate-y-1 hover:shadow-[var(--shadow-md)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <div className="flex items-start gap-3">
        <div
          className="rounded-2xl p-3"
          style={{
            color: "var(--primary)",
            backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)",
          }}
        >
          <Users aria-hidden="true" className="h-5 w-5" />
        </div>
        <div className="min-w-0">
          <h2 className="truncate font-semibold tracking-tight">{team.name}</h2>
          <p className="mt-1 line-clamp-2 min-h-10 text-sm leading-5" style={{ color: "var(--muted-foreground)" }}>
            {team.description || "No description provided"}
          </p>
        </div>
      </div>

      <div className="mt-5 flex items-center justify-between gap-3 border-t pt-4" style={{ borderColor: "var(--border)" }}>
        <span className="inline-flex items-center gap-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Created {formatDate(team.createdAt)}
        </span>
        <span className="inline-flex items-center gap-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
          Updated {formatDate(team.updatedAt)}
          <ExternalLink aria-hidden="true" className="h-3 w-3 opacity-0 transition-opacity group-hover:opacity-100" />
        </span>
      </div>
    </article>
  );
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
