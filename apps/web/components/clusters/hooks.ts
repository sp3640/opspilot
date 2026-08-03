"use client";

import { useEffect, useState } from "react";
import type { ClusterFilters, ClusterView } from "./types";

const initialFilters: ClusterFilters = {
  query: "",
  provider: "all",
  status: "all",
  projectId: "all",
  sort: "updated_at",
  order: "desc",
};

export function useClustersWorkspace() {
  const [filters, setFilters] = useState<ClusterFilters>(initialFilters);
  const [view, setView] = useState<ClusterView>("grid");
  const [page, setPage] = useState(1);
  const [selectedClusterID, setSelectedClusterID] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const debouncedSearch = useDebouncedValue(filters.query, 300);

  const updateFilters = (next: Partial<ClusterFilters>) => {
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
    selectedClusterID,
    setSelectedClusterID,
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
