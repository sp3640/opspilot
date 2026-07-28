"use client";

import { useState } from "react";
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
import type { CreateProjectInput, Project } from "./types";

type ProjectsWorkspaceProps = { initialProjects: Project[]; error?: string; loading?: boolean };

/** Complete, stateful Projects portfolio workspace composed from small feature components. */
export function ProjectsWorkspace({ initialProjects, error, loading = false }: ProjectsWorkspaceProps) {
  const [projectData, setProjectData] = useState(initialProjects);
  const [createOpen, setCreateOpen] = useState(false);
  const workspace = useProjectsWorkspace(projectData);
  const hasFilters = workspace.filters.query.length > 0 || workspace.filters.environment !== "all" || workspace.filters.health !== "all";
  const clearFilters = () => workspace.updateFilters({ query: "", environment: "all", health: "all" });
  const createProject = (input: CreateProjectInput) => {
    const project: Project = { id: input.name.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)/g, ""), name: input.name, description: input.description, environment: input.environment, health: "healthy", owner: { name: "Siddharth", initials: "S" }, members: [{ name: "Siddharth", initials: "S" }], services: 0, lastDeployment: "Not deployed", updatedAt: new Date().toISOString(), deployments: 0, icon: "folder-kanban" };
    setProjectData((projects) => [project, ...projects]);
    workspace.setSelectedProject(project);
  };

  return <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8"><PageHeader title="Projects" description="Manage the services, ownership, and operational posture of your platform portfolio." breadcrumb={[{ label: "Overview", href: "/" }, { label: "Projects" }]} /><SectionCard><ProjectToolbar filters={workspace.filters} view={workspace.view} onFiltersChange={workspace.updateFilters} onViewChange={workspace.setView} onRefresh={() => setProjectData((projects) => [...projects])} onCreate={() => setCreateOpen(true)} /><div className="mt-6">{error ? <ErrorState description={error} onRetry={() => setProjectData((projects) => [...projects])} /> : loading ? <ProjectSkeleton view={workspace.view} /> : workspace.visibleProjects.length === 0 ? <ProjectEmpty hasFilters={hasFilters} onClear={clearFilters} onCreate={() => setCreateOpen(true)} /> : workspace.view === "grid" ? <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{workspace.visibleProjects.map((project) => <ProjectCard key={project.id} project={project} onOpen={workspace.setSelectedProject} />)}</div> : <ProjectTable projects={workspace.visibleProjects} onOpen={workspace.setSelectedProject} />}</div>{workspace.filteredProjects.length > 0 && <Pagination page={workspace.page} pageCount={workspace.pageCount} total={workspace.filteredProjects.length} onPageChange={workspace.setPage} />}</SectionCard><ProjectDetailsDrawer project={workspace.selectedProject} onClose={() => workspace.setSelectedProject(null)} /><CreateProjectModal open={createOpen} onClose={() => setCreateOpen(false)} onCreate={createProject} /></div>;
}

function Pagination({ page, pageCount, total, onPageChange }: { page: number; pageCount: number; total: number; onPageChange: (page: number) => void }) {
  return <nav aria-label="Projects pagination" className="mt-6 flex flex-col gap-3 border-t pt-5 sm:flex-row sm:items-center sm:justify-between" style={{ borderColor: "var(--border)" }}><p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Showing {Math.min((page - 1) * 6 + 1, total)}–{Math.min(page * 6, total)} of {total} projects</p><div className="flex items-center gap-2"><button type="button" disabled={page === 1} onClick={() => onPageChange(page - 1)} className="rounded-xl border px-3 py-2 text-sm font-medium disabled:opacity-40" style={{ borderColor: "var(--border)" }}>Previous</button><span className="px-2 text-sm tabular-nums" style={{ color: "var(--muted-foreground)" }}>Page {page} of {pageCount}</span><button type="button" disabled={page === pageCount} onClick={() => onPageChange(page + 1)} className="rounded-xl border px-3 py-2 text-sm font-medium disabled:opacity-40" style={{ borderColor: "var(--border)" }}>Next</button></div></nav>;
}
