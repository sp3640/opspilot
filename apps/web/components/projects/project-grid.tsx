"use client";

import { useProjects } from "@/hooks/use-projects";
import {
  PAGINATION_DEFAULT_PAGE_SIZE,
  PROJECT_SORT_FIELDS,
  SORT_ORDERS,
} from "@/lib/constants";
import { ErrorState, PageHeader } from "@/components/common";
import { SectionCard } from "@/components/dashboard";

import { CreateProjectModal } from "./create-project-modal";
import { useProjectsWorkspace } from "./hooks";
import { ProjectCard } from "./project-card";
import { ProjectDetailsDrawer } from "./project-details-drawer";
import { ProjectEmpty } from "./project-empty";
import { ProjectSkeleton } from "./project-skeleton";
import { ProjectTable } from "./project-table";
import { ProjectToolbar } from "./project-toolbar";
/** Complete, stateful Projects portfolio workspace composed from small feature components. */
export function ProjectsWorkspace() {
  const workspace = useProjectsWorkspace();
  const sort = workspace.filters.sort === PROJECT_SORT_FIELDS.NAME ? PROJECT_SORT_FIELDS.NAME : PROJECT_SORT_FIELDS.UPDATED_AT;
  const { data, error, isError, isLoading, isFetching, refetch } = useProjects({ page: workspace.page, limit: PAGINATION_DEFAULT_PAGE_SIZE, search: workspace.debouncedSearch.trim() || undefined, sort, order: sort === PROJECT_SORT_FIELDS.NAME ? SORT_ORDERS.ASC : SORT_ORDERS.DESC });
  const projects = data?.items ?? [];
  const hasFilters = workspace.filters.query.length > 0;
  const clearFilters = () => workspace.updateFilters({ query: "", environment: "all", health: "all" });
  const errorMessage = error instanceof Error ? error.message : "Unable to load projects. Please try again.";

  return <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8"><PageHeader title="Projects" description="Manage the services, ownership, and operational posture of your platform portfolio." breadcrumb={[{ label: "Overview", href: "/" }, { label: "Projects" }]} /><SectionCard><ProjectToolbar filters={workspace.filters} view={workspace.view} onFiltersChange={workspace.updateFilters} onViewChange={workspace.setView} onRefresh={() => { void refetch(); }} onCreate={() => workspace.setCreateOpen(true)} refreshing={isFetching} /><div className="mt-6">{isError ? <ErrorState description={errorMessage} onRetry={() => { void refetch(); }} /> : isLoading ? <ProjectSkeleton view={workspace.view} /> : projects.length === 0 ? <ProjectEmpty hasFilters={hasFilters} onClear={clearFilters} onCreate={() => workspace.setCreateOpen(true)} /> : workspace.view === "grid" ? <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{projects.map((project) => <ProjectCard key={project.id} project={project} onOpen={workspace.setSelectedProjectID} />)}</div> : <ProjectTable projects={projects} onOpen={workspace.setSelectedProjectID} />}</div>{data && data.total > 0 && <Pagination page={data.page} pageCount={data.totalPages} total={data.total} limit={data.limit} onPageChange={workspace.setPage} />}</SectionCard><ProjectDetailsDrawer projectID={workspace.selectedProjectID} onClose={() => workspace.setSelectedProjectID(null)} /><CreateProjectModal open={workspace.createOpen} onClose={() => workspace.setCreateOpen(false)} onCreated={(project) => workspace.setSelectedProjectID(project.id)} /></div>;
}

function Pagination({ page, pageCount, total, limit, onPageChange }: { page: number; pageCount: number; total: number; limit: number; onPageChange: (page: number) => void }) {
  return <nav aria-label="Projects pagination" className="mt-6 flex flex-col gap-3 border-t pt-5 sm:flex-row sm:items-center sm:justify-between" style={{ borderColor: "var(--border)" }}><p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Showing {Math.min((page - 1) * limit + 1, total)}–{Math.min(page * limit, total)} of {total} projects</p><div className="flex items-center gap-2"><button type="button" disabled={page === 1} onClick={() => onPageChange(page - 1)} className="rounded-xl border px-3 py-2 text-sm font-medium disabled:opacity-40" style={{ borderColor: "var(--border)" }}>Previous</button><span className="px-2 text-sm tabular-nums" style={{ color: "var(--muted-foreground)" }}>Page {page} of {pageCount}</span><button type="button" disabled={page === pageCount} onClick={() => onPageChange(page + 1)} className="rounded-xl border px-3 py-2 text-sm font-medium disabled:opacity-40" style={{ borderColor: "var(--border)" }}>Next</button></div></nav>;
}
