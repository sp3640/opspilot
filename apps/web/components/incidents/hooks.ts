"use client";

import { useEffect, useState } from "react";

export type IncidentFilters = {
  query: string;
};

export type IncidentView = "grid";

export function useIncidentsWorkspace() {
  const [filters, setFilters] = useState<IncidentFilters>({ query: "" });
  const [page, setPage] = useState(1);
  const [selectedIncidentID, setSelectedIncidentID] = useState<number | null>(null);
  const debouncedSearch = useDebouncedValue(filters.query, 300);

  const updateFilters = (next: Partial<IncidentFilters>) => {
    setFilters((current) => ({ ...current, ...next }));
    setPage(1);
  };

  return {
    filters,
    updateFilters,
    page,
    setPage,
    selectedIncidentID,
    setSelectedIncidentID,
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
