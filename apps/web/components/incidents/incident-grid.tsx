"use client";

import { useState } from "react";
import { useIncidents } from "@/hooks/use-incidents";
import { ErrorState, PageHeader } from "@/components/common";
import { SectionCard } from "@/components/dashboard";

import { CreateIncidentModal } from "./create-incident-modal";
import { IncidentDetailsDrawer } from "./incident-details-drawer";
import { useIncidentsWorkspace } from "./hooks";
import { IncidentCard } from "./incident-card";
import { IncidentEmpty } from "./incident-empty";
import { IncidentSkeleton } from "./incident-skeleton";
import { IncidentToolbar } from "./incident-toolbar";

export function IncidentsWorkspace() {
  const workspace = useIncidentsWorkspace();
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const { data, error, isError, isLoading, isFetching, refetch } = useIncidents({
    page: workspace.page,
    limit: 6,
    search: workspace.debouncedSearch.trim() || undefined,
  });

  const incidents = data?.items ?? [];
  const hasFilters = workspace.filters.query.length > 0;
  const clearFilters = () => workspace.updateFilters({ query: "" });
  const errorMessage =
    error instanceof Error ? error.message : "Unable to load incidents. Please try again.";

  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <PageHeader
        title="Incidents"
        description="Monitor and manage active incidents across your platform."
        breadcrumb={[
          { label: "Overview", href: "/" },
          { label: "Incidents" },
        ]}
      />

      <SectionCard>
        <IncidentToolbar
          filters={workspace.filters}
          onFiltersChange={workspace.updateFilters}
          onRefresh={() => {
            void refetch();
          }}
          onCreate={() => {
            setCreateModalOpen(true);
          }}
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
            <IncidentSkeleton />
          ) : incidents.length === 0 ? (
            <IncidentEmpty hasFilters={hasFilters} onClear={clearFilters} />
          ) : (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {incidents.map((incident) => (
                <IncidentCard
                  key={incident.id}
                  incident={incident}
                  onSelect={workspace.setSelectedIncidentID}
                />
              ))}
            </div>
          )}
        </div>

        {data && data.total > 0 && (
          <Pagination
            page={data.page}
            pageCount={data.totalPages}
            total={data.total}
            limit={data.limit}
            onPageChange={workspace.setPage}
          />
        )}
      </SectionCard>

      <CreateIncidentModal open={createModalOpen} onClose={() => setCreateModalOpen(false)} />
      <IncidentDetailsDrawer
        incidentID={workspace.selectedIncidentID}
        onClose={() => workspace.setSelectedIncidentID(null)}
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
      aria-label="Incidents pagination"
      className="mt-6 flex flex-col gap-3 border-t pt-5 sm:flex-row sm:items-center sm:justify-between"
      style={{ borderColor: "var(--border)" }}
    >
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        Showing {Math.min((page - 1) * limit + 1, total)}–{Math.min(page * limit, total)} of{" "}
        {total} incidents
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
