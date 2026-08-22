"use client";

import { useEffect, useState } from "react";

import type { TeamFilters, TeamView } from "./types";

const initialFilters: TeamFilters = { query: "", sort: "updated_at", order: "desc" };

export function useTeamsWorkspace() {
  const [filters, setFilters] = useState<TeamFilters>(initialFilters);
  const [view, setView] = useState<TeamView>("grid");
  const [page, setPage] = useState(1);
  const [selectedTeamID, setSelectedTeamID] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const debouncedSearch = useDebouncedValue(filters.query, 300);

  const updateFilters = (next: Partial<TeamFilters>) => {
    setFilters((current) => ({ ...current, ...next }));
    setPage(1);
  };

  return {
    filters,
    updateFilters,
    view,
    setView,
    page,
    setPage,
    selectedTeamID,
    setSelectedTeamID,
    createOpen,
    setCreateOpen,
    debouncedSearch,
  };
}

function useDebouncedValue(value: string, delay: number) {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timeout = window.setTimeout(() => setDebouncedValue(value), delay);
    return () => window.clearTimeout(timeout);
  }, [value, delay]);

  return debouncedValue;
}
