"use client";

import { useMemo } from "react";

import { ErrorState, PageHeader } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { useClusters } from "@/hooks/use-clusters";
import { useLatestMetrics, useMetrics } from "@/hooks/use-metrics";
import { useProjects } from "@/hooks/use-projects";
import { METRIC_TYPES, PAGINATION_DEFAULT_PAGE_SIZE, PROJECT_LOOKUP_QUERY } from "@/lib/constants";

import { MetricCharts } from "./metric-charts";
import { MetricDetailsDrawer } from "./metric-details-drawer";
import { MetricEmpty } from "./metric-empty";
import { MetricsGrid } from "./metrics-grid";
import { MetricSkeleton } from "./metric-skeleton";
import { MetricsTable } from "./metrics-table";
import { MetricsToolbar } from "./metrics-toolbar";
import { format } from "./metrics-card";
import { useMetricsWorkspace } from "./hooks";

export function MetricsWorkspace() {
  const workspace = useMetricsWorkspace();
  const { data: projectsData } = useProjects(PROJECT_LOOKUP_QUERY);
  const { data: clustersData } = useClusters(PROJECT_LOOKUP_QUERY);

  const projects = useMemo(() => projectsData?.items ?? [], [projectsData?.items]);
  const clusters = useMemo(() => clustersData?.items ?? [], [clustersData?.items]);

  const queryParams = useMemo(
    () => ({
      page: workspace.page,
      limit: PAGINATION_DEFAULT_PAGE_SIZE,
      search: workspace.debouncedSearch.trim() || undefined,
      sort: workspace.filters.sort,
      order: workspace.filters.order,
      projectId: workspace.filters.projectId === "all" ? undefined : workspace.filters.projectId,
    }),
    [
      workspace.page,
      workspace.debouncedSearch,
      workspace.filters.sort,
      workspace.filters.order,
      workspace.filters.projectId,
    ]
  );

  const metricsQuery = useMetrics(queryParams);
  const allMetrics = metricsQuery.data?.items ?? [];
  const metrics = allMetrics.filter(
    (metric) =>
      (workspace.filters.clusterId === "all" || metric.clusterId === workspace.filters.clusterId) &&
      (workspace.filters.resourceId === "all" || metric.resourceId === workspace.filters.resourceId) &&
      (workspace.filters.metricType === "all" || metric.metricType === workspace.filters.metricType)
  );

  const selectedMetric = metrics.find((metric) => metric.id === workspace.selectedMetricID) ?? metrics[0] ?? null;

  const latestMetric = useLatestMetrics(
    selectedMetric
      ? {
          projectId: selectedMetric.projectId,
          clusterId: selectedMetric.clusterId,
          resourceId: selectedMetric.resourceId,
          metricType: selectedMetric.metricType,
          metricName: selectedMetric.metricName,
        }
      : null
  );

  const projectNames = useMemo(() => new Map(projects.map((project) => [project.id, project.name])), [projects]);
  const clusterNames = useMemo(() => new Map(clusters.map((cluster) => [cluster.id, cluster.name])), [clusters]);

  const summaryProjectId =
    workspace.filters.projectId !== "all"
      ? workspace.filters.projectId
      : selectedMetric?.projectId ?? projects[0]?.id ?? null;

  const hasFilters =
    workspace.filters.query.length > 0 ||
    workspace.filters.projectId !== "all" ||
    workspace.filters.clusterId !== "all" ||
    workspace.filters.resourceId !== "all" ||
    workspace.filters.metricType !== "all" ||
    workspace.filters.timeRange !== "24h" ||
    workspace.filters.sort !== "timestamp" ||
    workspace.filters.order !== "desc";

  const clearFilters = () => {
    workspace.updateFilters({
      query: "",
      projectId: "all",
      clusterId: "all",
      resourceId: "all",
      metricType: "all",
      timeRange: "24h",
      sort: "timestamp",
      order: "desc",
    });
  };

  const errorMessage = metricsQuery.error instanceof Error ? metricsQuery.error.message : "Unable to load metrics.";

  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <PageHeader
        title="Metrics"
        description="Explore live, historical, and aggregated infrastructure metrics."
        breadcrumb={[{ label: "Overview", href: "/" }, { label: "Metrics" }]}
      />

      <SectionCard>
        <MetricsToolbar
          filters={workspace.filters}
          view={workspace.view}
          projects={projects}
          clusters={clusters}
          onChange={workspace.updateFilters}
          onView={workspace.setView}
          onRefresh={() => {
            void metricsQuery.refetch();
          }}
          refreshing={metricsQuery.isFetching}
        />

        <div className="mt-6">
          {metricsQuery.isError ? (
            <ErrorState
              description={errorMessage}
              onRetry={() => {
                void metricsQuery.refetch();
              }}
            />
          ) : metricsQuery.isLoading ? (
            <MetricSkeleton view={workspace.view} />
          ) : metrics.length === 0 ? (
            <MetricEmpty hasFilters={hasFilters} onClear={clearFilters} />
          ) : workspace.view === "charts" ? (
            <div className="space-y-6">
              <section>
                <h2 className="text-base font-semibold">Metrics Dashboard</h2>
                <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                  Latest metrics, historical trend, and aggregate values for the selected scope.
                </p>
              </section>

              <section>
                <h3 className="text-sm font-semibold">Latest Metrics</h3>
                <LatestMetricsSummary projectId={summaryProjectId} />
              </section>

              <section className="rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
                <p className="text-xs uppercase tracking-wide" style={{ color: "var(--muted-foreground)" }}>
                  Latest metric for current selection
                </p>
                <p className="mt-1 text-3xl font-semibold">
                  {latestMetric.data ? `${format(latestMetric.data.value)} ${latestMetric.data.unit}` : "Loading..."}
                </p>
                <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                  {selectedMetric?.metricName} · {selectedMetric?.resourceKind}
                </p>
              </section>

              <section>
                <h3 className="mb-3 text-sm font-semibold">Historical Metrics and Aggregated Metrics</h3>
                <MetricCharts metric={selectedMetric} range={workspace.filters.timeRange} />
              </section>
            </div>
          ) : workspace.view === "cards" ? (
            <MetricsGrid
              metrics={metrics}
              projectNames={projectNames}
              clusterNames={clusterNames}
              onOpen={workspace.setSelectedMetricID}
            />
          ) : (
            <MetricsTable
              metrics={metrics}
              projectNames={projectNames}
              clusterNames={clusterNames}
              timeRange={workspace.filters.timeRange}
              onOpen={workspace.setSelectedMetricID}
            />
          )}
        </div>

        {metricsQuery.data && metricsQuery.data.total > 0 ? (
          <Pagination
            page={metricsQuery.data.page}
            pages={metricsQuery.data.totalPages}
            total={metricsQuery.data.total}
            limit={metricsQuery.data.limit}
            onChange={workspace.setPage}
          />
        ) : null}
      </SectionCard>

      <MetricDetailsDrawer metric={metrics.find((metric) => metric.id === workspace.selectedMetricID) ?? null} onClose={() => workspace.setSelectedMetricID(null)} />
    </div>
  );
}

function LatestMetricsSummary({ projectId }: { projectId: string | null }) {
  if (!projectId) {
    return (
      <p className="mt-2 text-sm" style={{ color: "var(--muted-foreground)" }}>
        Select a project to load latest metrics.
      </p>
    );
  }

  return (
    <div className="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      {METRIC_TYPES.map((metricType) => (
        <LatestMetricCard key={metricType} projectId={projectId} metricType={metricType} />
      ))}
    </div>
  );
}

function LatestMetricCard({ projectId, metricType }: { projectId: string; metricType: string }) {
  const latest = useLatestMetrics({ projectId, metricType });

  return (
    <article className="rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
      <p className="text-xs uppercase tracking-wide" style={{ color: "var(--muted-foreground)" }}>
        {metricType}
      </p>
      <p className="mt-1 text-lg font-semibold">
        {latest.data ? `${format(latest.data.value)} ${latest.data.unit}` : latest.isLoading ? "Loading..." : "Unavailable"}
      </p>
      <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
        {latest.data ? new Date(latest.data.timestamp).toLocaleString() : ""}
      </p>
    </article>
  );
}

function Pagination({
  page,
  pages,
  total,
  limit,
  onChange,
}: {
  page: number;
  pages: number;
  total: number;
  limit: number;
  onChange: (page: number) => void;
}) {
  return (
    <nav
      aria-label="Metrics pagination"
      className="mt-6 flex items-center justify-between border-t pt-5"
      style={{ borderColor: "var(--border)" }}
    >
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        Showing {Math.min((page - 1) * limit + 1, total)}-{Math.min(page * limit, total)} of {total} metrics
      </p>
      <div className="flex gap-2">
        <button
          type="button"
          disabled={page === 1}
          onClick={() => onChange(page - 1)}
          className="rounded-xl border px-3 py-2 disabled:opacity-40"
          style={{ borderColor: "var(--border)" }}
        >
          Previous
        </button>
        <span className="px-2 py-2 text-sm">Page {page} of {pages}</span>
        <button
          type="button"
          disabled={page === pages}
          onClick={() => onChange(page + 1)}
          className="rounded-xl border px-3 py-2 disabled:opacity-40"
          style={{ borderColor: "var(--border)" }}
        >
          Next
        </button>
      </div>
    </nav>
  );
}
