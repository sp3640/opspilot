"use client";

import { useMemo } from "react";

import { ErrorState, PageHeader } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { useClusters } from "@/hooks/use-clusters";
import { useProjects } from "@/hooks/use-projects";
import {
  PAGINATION_DEFAULT_PAGE_SIZE,
  PROJECT_LOOKUP_QUERY,
} from "@/lib/constants";
import { CLUSTER_SORT_FIELDS } from "@/lib/constants/cluster";

import { ClusterDetailsDrawer } from "./cluster-details-drawer";
import { ClusterEmpty } from "./cluster-empty";
import { ClusterGrid } from "./cluster-grid";
import { ClusterSkeleton } from "./cluster-skeleton";
import { ClusterTable } from "./cluster-table";
import { ClusterToolbar } from "./cluster-toolbar";
import { CreateClusterModal } from "./create-cluster-modal";
import { useClustersWorkspace } from "./hooks";

export function ClustersWorkspace() {
  const workspace = useClustersWorkspace();

  const { data: projectsData } = useProjects({
    page: PROJECT_LOOKUP_QUERY.page,
    limit: PROJECT_LOOKUP_QUERY.limit,
    sort: PROJECT_LOOKUP_QUERY.sort,
    order: PROJECT_LOOKUP_QUERY.order,
  });

  const queryParams = useMemo(
    () => ({
      page: workspace.page,
      limit: PAGINATION_DEFAULT_PAGE_SIZE,
      search: workspace.debouncedSearch.trim() || undefined,
      sort: workspace.filters.sort,
      order: workspace.filters.order,
      provider: workspace.filters.provider === "all" ? undefined : workspace.filters.provider,
      status: workspace.filters.status === "all" ? undefined : workspace.filters.status,
      projectId: workspace.filters.projectId === "all" ? undefined : workspace.filters.projectId,
    }),
    [
      workspace.page,
      workspace.debouncedSearch,
      workspace.filters.sort,
      workspace.filters.order,
      workspace.filters.provider,
      workspace.filters.status,
      workspace.filters.projectId,
    ]
  );

  const { data, error, isError, isLoading, isFetching, refetch } = useClusters(queryParams);

  const clusters = data?.items ?? [];
  const projects = useMemo(() => projectsData?.items ?? [], [projectsData?.items]);
  const projectNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const project of projects) {
      map.set(project.id, project.name);
    }
    return map;
  }, [projects]);

  const hasFilters =
    workspace.filters.query.length > 0 ||
    workspace.filters.provider !== "all" ||
    workspace.filters.status !== "all" ||
    workspace.filters.projectId !== "all" ||
    workspace.filters.sort !== CLUSTER_SORT_FIELDS.UPDATED_AT ||
    workspace.filters.order !== "desc";

  const clearFilters = () => {
    workspace.updateFilters({
      query: "",
      provider: "all",
      status: "all",
      projectId: "all",
      sort: CLUSTER_SORT_FIELDS.UPDATED_AT,
      order: "desc",
    });
  };

  const errorMessage = error instanceof Error ? error.message : "Unable to load clusters. Please try again.";

  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <PageHeader
        title="Clusters"
        description="Manage runtime clusters, connection health, and default routing across projects."
        breadcrumb={[{ label: "Overview", href: "/" }, { label: "Clusters" }]}
      />

      <SectionCard>
        <ClusterToolbar
          filters={workspace.filters}
          view={workspace.view}
          projects={projects}
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
            <ClusterSkeleton view={workspace.view} />
          ) : clusters.length === 0 ? (
            <ClusterEmpty
              hasFilters={hasFilters}
              onClear={clearFilters}
              onCreate={() => workspace.setCreateOpen(true)}
            />
          ) : workspace.view === "grid" ? (
            <ClusterGrid clusters={clusters} projectNameById={projectNameById} onOpen={workspace.setSelectedClusterID} />
          ) : (
            <ClusterTable clusters={clusters} projectNameById={projectNameById} onOpen={workspace.setSelectedClusterID} />
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

      <ClusterDetailsDrawer
        clusterID={workspace.selectedClusterID}
        projectNameById={projectNameById}
        onClose={() => workspace.setSelectedClusterID(null)}
      />

      <CreateClusterModal
        open={workspace.createOpen}
        onClose={() => workspace.setCreateOpen(false)}
        onCreated={(cluster) => workspace.setSelectedClusterID(cluster.id)}
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
      aria-label="Clusters pagination"
      className="mt-6 flex flex-col gap-3 border-t pt-5 sm:flex-row sm:items-center sm:justify-between"
      style={{ borderColor: "var(--border)" }}
    >
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        Showing {Math.min((page - 1) * limit + 1, total)}-{Math.min(page * limit, total)} of {total} clusters
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
