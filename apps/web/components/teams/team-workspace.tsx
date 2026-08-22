"use client";

import { useMemo } from "react";

import { ErrorState, PageHeader } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { useTeams } from "@/hooks/use-teams";
import { PAGINATION_DEFAULT_PAGE_SIZE } from "@/lib/constants";

import { CreateTeamModal } from "./create-team-modal";
import { useTeamsWorkspace } from "./hooks";
import { TeamDetailsDrawer } from "./team-details-drawer";
import { TeamEmpty } from "./team-empty";
import { TeamGrid } from "./team-grid";
import { TeamSkeleton } from "./team-skeleton";
import { TeamTable } from "./team-table";
import { TeamToolbar } from "./team-toolbar";

export function TeamsWorkspace() {
  const workspace = useTeamsWorkspace();

  const queryParams = useMemo(
    () => ({
      page: workspace.page,
      limit: PAGINATION_DEFAULT_PAGE_SIZE,
      search: workspace.debouncedSearch.trim() || undefined,
      sort: workspace.filters.sort,
      order: workspace.filters.order,
    }),
    [workspace.page, workspace.debouncedSearch, workspace.filters.sort, workspace.filters.order]
  );

  const { data, error, isError, isLoading, isFetching, refetch } = useTeams(queryParams);

  const teams = data?.items ?? [];
  const hasFilters = workspace.filters.query.length > 0;
  const clearFilters = () => workspace.updateFilters({ query: "" });
  const errorMessage = error instanceof Error ? error.message : "Unable to load teams. Please try again.";

  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <PageHeader
        title="Teams"
        description="Organize members into teams and manage access across your organization."
        breadcrumb={[{ label: "Overview", href: "/" }, { label: "Teams" }]}
      />

      <SectionCard>
        <TeamToolbar
          filters={workspace.filters}
          view={workspace.view}
          onFiltersChange={workspace.updateFilters}
          onViewChange={workspace.setView}
          onRefresh={() => {
            void refetch();
          }}
          onCreate={() => workspace.setCreateOpen(true)}
          refreshing={isFetching}
        />

        <div className="mt-6">
          {isError ? (
            <ErrorState
              description={errorMessage}
              onRetry={() => {
                void refetch();
              }}
            />
          ) : isLoading ? (
            <TeamSkeleton view={workspace.view} />
          ) : teams.length === 0 ? (
            <TeamEmpty
              hasFilters={hasFilters}
              onClear={clearFilters}
              onCreate={() => workspace.setCreateOpen(true)}
            />
          ) : workspace.view === "grid" ? (
            <TeamGrid teams={teams} onOpen={workspace.setSelectedTeamID} />
          ) : (
            <TeamTable teams={teams} onOpen={workspace.setSelectedTeamID} />
          )}
        </div>

        {data && data.total > 0 ? (
          <Pagination
            page={data.page}
            pageCount={data.totalPages}
            total={data.total}
            limit={data.limit}
            onPageChange={workspace.setPage}
          />
        ) : null}
      </SectionCard>

      <TeamDetailsDrawer
        teamID={workspace.selectedTeamID}
        onClose={() => workspace.setSelectedTeamID(null)}
      />

      <CreateTeamModal
        open={workspace.createOpen}
        onClose={() => workspace.setCreateOpen(false)}
        onCreated={(team) => workspace.setSelectedTeamID(team.id)}
      />
    </div>
  );
}

function Pagination({
  page,
  pageCount,
  total,
  limit,
  onPageChange,
}: {
  page: number;
  pageCount: number;
  total: number;
  limit: number;
  onPageChange: (page: number) => void;
}) {
  return (
    <nav
      aria-label="Teams pagination"
      className="mt-6 flex flex-col gap-3 border-t pt-5 sm:flex-row sm:items-center sm:justify-between"
      style={{ borderColor: "var(--border)" }}
    >
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        Showing {Math.min((page - 1) * limit + 1, total)}–{Math.min(page * limit, total)} of {total} teams
      </p>
      <div className="flex items-center gap-2">
        <button
          type="button"
          disabled={page === 1}
          onClick={() => onPageChange(page - 1)}
          className="rounded-xl border px-3 py-2 text-sm font-medium disabled:opacity-40"
          style={{ borderColor: "var(--border)" }}
        >
          Previous
        </button>
        <span className="px-2 text-sm tabular-nums" style={{ color: "var(--muted-foreground)" }}>
          Page {page} of {pageCount}
        </span>
        <button
          type="button"
          disabled={page === pageCount}
          onClick={() => onPageChange(page + 1)}
          className="rounded-xl border px-3 py-2 text-sm font-medium disabled:opacity-40"
          style={{ borderColor: "var(--border)" }}
        >
          Next
        </button>
      </div>
    </nav>
  );
}
