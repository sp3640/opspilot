"use client";

import { ChevronRight, Users } from "lucide-react";

import type { Team } from "./types";

type TeamTableProps = { teams: Team[]; onOpen: (teamID: string) => void };

/** Dense, keyboard-operable representation of API-backed teams. */
export function TeamTable({ teams, onOpen }: TeamTableProps) {
  return (
    <div className="overflow-x-auto rounded-2xl border" style={{ borderColor: "var(--border)" }}>
      <table className="w-full min-w-[800px] text-left text-sm">
        <caption className="sr-only">Teams list with description and activity</caption>
        <thead
          className="text-xs uppercase tracking-[0.1em]"
          style={{
            color: "var(--muted-foreground)",
            backgroundColor: "color-mix(in srgb, var(--muted) 55%, transparent)",
          }}
        >
          <tr>
            <th scope="col" className="px-5 py-3 font-semibold">Team</th>
            <th scope="col" className="px-5 py-3 font-semibold">Description</th>
            <th scope="col" className="px-5 py-3 font-semibold">Created</th>
            <th scope="col" className="px-5 py-3 font-semibold">Updated</th>
            <th scope="col" className="w-12 px-5 py-3"><span className="sr-only">Open</span></th>
          </tr>
        </thead>
        <tbody>
          {teams.map((team) => (
            <tr
              key={team.id}
              role="button"
              tabIndex={0}
              aria-label={`Open team details for ${team.name}`}
              onClick={() => onOpen(team.id)}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onOpen(team.id);
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
                    <Users aria-hidden="true" className="h-4 w-4" />
                  </span>
                  <p className="font-medium">{team.name}</p>
                </div>
              </td>
              <td className="max-w-72 truncate px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                {team.description || "-"}
              </td>
              <td className="px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                {formatDate(team.createdAt)}
              </td>
              <td className="px-5 py-4" style={{ color: "var(--muted-foreground)" }}>
                {formatDate(team.updatedAt)}
              </td>
              <td className="px-5 py-4">
                <ChevronRight aria-hidden="true" className="h-4 w-4" style={{ color: "var(--muted-foreground)" }} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
