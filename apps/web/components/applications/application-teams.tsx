"use client";

import { useMemo, useState } from "react";
import axios from "axios";
import { Trash2, UserPlus, Users } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { Button } from "@/components/ui/button";
import {
  useApplicationTeams,
  useAssignApplicationTeam,
  useRemoveApplicationTeam,
} from "@/hooks/use-application-teams";
import { useTeams } from "@/hooks/use-teams";
import { useIsPlatformAdmin } from "@/store/auth-store";

const TEAM_LOOKUP_QUERY = { page: 1, limit: 100, sort: "name" as const, order: "asc" as const };

const selectClass =
  "mt-1 h-11 w-full rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]";

/** Application ↔ Team ownership. Mounted only when the Teams tab is active. */
export function ApplicationTeams({ applicationId }: { applicationId: string }) {
  const isAdmin = useIsPlatformAdmin();
  const [selectedTeamId, setSelectedTeamId] = useState<string | null>(null);
  const [teamToAssign, setTeamToAssign] = useState("");
  const [assignError, setAssignError] = useState<string | null>(null);

  const { data: teamsData } = useTeams(TEAM_LOOKUP_QUERY);
  const teams = teamsData?.items ?? [];
  const teamById = useMemo(() => new Map(teams.map((team) => [team.id, team])), [teams]);

  const { data, error, isError, isLoading, refetch } = useApplicationTeams(applicationId);
  const assignments = data?.items ?? [];

  const assignedTeamIds = useMemo(() => new Set(assignments.map((assignment) => assignment.teamId)), [assignments]);
  const availableTeams = useMemo(
    () => teams.filter((team) => !assignedTeamIds.has(team.id)),
    [teams, assignedTeamIds]
  );

  const assignTeam = useAssignApplicationTeam();
  const removeTeam = useRemoveApplicationTeam();

  const handleAssign = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setAssignError(null);

    if (!teamToAssign) {
      setAssignError("Select a team to assign.");
      return;
    }

    try {
      await assignTeam.mutateAsync({ applicationId, payload: { teamId: teamToAssign } });
      setTeamToAssign("");
    } catch (assignErrorValue) {
      setAssignError(getErrorMessage(assignErrorValue, "Unable to assign team. Please try again."));
    }
  };

  const handleRemove = async (teamId: string) => {
    setAssignError(null);
    try {
      await removeTeam.mutateAsync({ applicationId, teamId });
      setSelectedTeamId((current) => (current === teamId ? null : current));
    } catch (removeErrorValue) {
      setAssignError(getErrorMessage(removeErrorValue, "Unable to remove team. Please try again."));
    }
  };

  if (isError) {
    return (
      <ErrorState
        description={error instanceof Error ? error.message : "Unable to load application teams. Please try again."}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  if (isLoading) {
    return <ListSkeleton rows={3} />;
  }

  return (
    <div className="space-y-5">
      {isAdmin ? (
        <form onSubmit={handleAssign} className="flex items-end gap-2">
          <div className="flex-1">
            <label htmlFor="assign-application-team-select" className="text-xs" style={{ color: "var(--muted-foreground)" }}>
              Assign a team
            </label>
            <select
              id="assign-application-team-select"
              value={teamToAssign}
              onChange={(event) => setTeamToAssign(event.target.value)}
              disabled={assignTeam.isPending || availableTeams.length === 0}
              className={selectClass}
              style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
            >
              <option value="">{availableTeams.length === 0 ? "No teams available to assign" : "Select a team"}</option>
              {availableTeams.map((team) => (
                <option key={team.id} value={team.id}>
                  {team.name}
                </option>
              ))}
            </select>
          </div>
          <Button type="submit" loading={assignTeam.isPending} disabled={assignTeam.isPending || !teamToAssign}>
            <UserPlus aria-hidden="true" className="h-4 w-4" />
            Assign
          </Button>
        </form>
      ) : null}
      {assignError ? (
        <p className="text-xs" style={{ color: "var(--danger)" }}>
          {assignError}
        </p>
      ) : null}

      {assignments.length === 0 ? (
        <EmptyState
          icon={Users}
          title="No teams assigned"
          description="Assign a team above to grant it ownership of this application."
        />
      ) : (
        <div className="space-y-3">
          {assignments.map((assignment) => {
            const team = teamById.get(assignment.teamId);
            const isSelected = selectedTeamId === assignment.teamId;

            return (
              <button
                key={assignment.id}
                type="button"
                onClick={() => setSelectedTeamId(assignment.teamId)}
                aria-pressed={isSelected}
                className="w-full rounded-2xl border p-4 text-left transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
                style={{ borderColor: isSelected ? "var(--primary)" : "var(--border)", backgroundColor: "var(--card)" }}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate font-semibold">{team?.name ?? assignment.teamId}</p>
                    <p className="mt-1 truncate text-xs" style={{ color: "var(--muted-foreground)" }}>
                      {team?.description || "No description provided"}
                    </p>
                  </div>
                  {isAdmin ? (
                    <Button
                      type="button"
                      variant="ghost"
                      onClick={(event) => {
                        event.stopPropagation();
                        void handleRemove(assignment.teamId);
                      }}
                      loading={removeTeam.isPending && removeTeam.variables?.teamId === assignment.teamId}
                      disabled={removeTeam.isPending}
                      className="h-9 w-9 shrink-0 rounded-xl p-0 text-[var(--danger)]"
                      aria-label={`Remove ${team?.name ?? assignment.teamId} from application`}
                    >
                      <Trash2 aria-hidden="true" className="h-4 w-4" />
                    </Button>
                  ) : null}
                </div>
                <p className="mt-3 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  Assigned {formatDate(assignment.createdAt)}
                </p>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}

function getErrorMessage(error: unknown, fallback: string) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallback;
  }
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return fallback;
}
