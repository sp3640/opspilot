"use client";

import { useMemo, useState } from "react";

import type { Project, ProjectFilters, ProjectView } from "./types";

const pageSize = 6;
const initialFilters: ProjectFilters = { query: "", environment: "all", health: "all", sort: "updated" };

/** State and derived data for the projects workspace; replace the data input with an API query when available. */
export function useProjectsWorkspace(initialProjects: Project[]) {
  const [filters, setFilters] = useState(initialFilters);
  const [view, setView] = useState<ProjectView>("grid");
  const [page, setPage] = useState(1);
  const [selectedProject, setSelectedProject] = useState<Project | null>(null);

  const filteredProjects = useMemo(() => {
    const normalizedQuery = filters.query.trim().toLowerCase();
    return initialProjects.filter((project) => {
      const matchesQuery = !normalizedQuery || `${project.name} ${project.description} ${project.owner.name}`.toLowerCase().includes(normalizedQuery);
      return matchesQuery && (filters.environment === "all" || project.environment === filters.environment) && (filters.health === "all" || project.health === filters.health);
    }).sort((first, second) => {
      if (filters.sort === "name") return first.name.localeCompare(second.name);
      if (filters.sort === "health") return first.health.localeCompare(second.health);
      return new Date(second.updatedAt).getTime() - new Date(first.updatedAt).getTime();
    });
  }, [filters, initialProjects]);

  const pageCount = Math.max(1, Math.ceil(filteredProjects.length / pageSize));
  const currentPage = Math.min(page, pageCount);
  const visibleProjects = filteredProjects.slice((currentPage - 1) * pageSize, currentPage * pageSize);
  const updateFilters = (next: Partial<ProjectFilters>) => { setFilters((current) => ({ ...current, ...next })); setPage(1); };

  return { filters, updateFilters, view, setView, page: currentPage, setPage, pageCount, filteredProjects, visibleProjects, selectedProject, setSelectedProject };
}
