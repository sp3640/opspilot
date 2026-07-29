"use client";

import { useEffect, useState } from "react";

import type { ProjectFilters, ProjectView } from "./types";

const initialFilters: ProjectFilters = { query: "", environment: "all", health: "all", sort: "updated" };

export function useProjectsWorkspace() {
  const [filters, setFilters] = useState(initialFilters);
  const [view, setView] = useState<ProjectView>("grid");
  const [page, setPage] = useState(1);
  const [selectedProjectID, setSelectedProjectID] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const debouncedSearch = useDebouncedValue(filters.query, 300);

  const updateFilters = (next: Partial<ProjectFilters>) => { setFilters((current) => ({ ...current, ...next })); setPage(1); };

  return { filters, updateFilters, view, setView, page, setPage, selectedProjectID, setSelectedProjectID, createOpen, setCreateOpen, debouncedSearch };
}

function useDebouncedValue(value: string, delay: number) {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timeout = window.setTimeout(() => setDebouncedValue(value), delay);
    return () => window.clearTimeout(timeout);
  }, [value, delay]);

  return debouncedValue;
}
