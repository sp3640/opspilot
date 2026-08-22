"use client";

import { forwardRef, useMemo } from "react";
import type { SelectHTMLAttributes } from "react";

import { useTeams } from "@/hooks/use-teams";

type OwnerTeamSelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  queryEnabled?: boolean;
};

const LOOKUP_PARAMS = { page: 1, limit: 100, sort: "name" as const, order: "asc" as const };

/** Owner-team selector, org-wide - optional, since not every incident has a designated owning team yet. */
export const OwnerTeamSelect = forwardRef<HTMLSelectElement, OwnerTeamSelectProps>(function OwnerTeamSelect(
  { queryEnabled = true, disabled, ...props },
  ref
) {
  const params = useMemo(() => LOOKUP_PARAMS, []);
  const { data, isLoading, isFetching, isError } = useTeams(params, queryEnabled);

  const teams = data?.items ?? [];
  const isBusy = isLoading || isFetching;

  return (
    <select
      ref={ref}
      {...props}
      style={{ backgroundColor: "var(--card)", color: "var(--foreground)", ...(props.style ?? {}) }}
      disabled={disabled || isBusy}
    >
      <option value="">No owner team</option>
      {isError ? (
        <option value="" disabled>Unable to load teams</option>
      ) : (
        teams.map((team) => (
          <option key={team.id} value={team.id}>{team.name}</option>
        ))
      )}
    </select>
  );
});
