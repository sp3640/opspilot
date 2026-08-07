"use client";

import { useEffect, useState } from "react";

import { PAGINATION_DEFAULT_PAGE_SIZE } from "@/lib/constants";
import type { ApplicationQueryParams } from "@/types/application-api";

export type ApplicationView = "grid" | "table";
export type ApplicationSort = NonNullable<ApplicationQueryParams["sort"]>;
export type ApplicationOrder = NonNullable<ApplicationQueryParams["order"]>;

export function useApplicationsWorkspace() {
  const [search, setSearchState] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimitState] = useState<number>(PAGINATION_DEFAULT_PAGE_SIZE);
  const [sort, setSortState] = useState<ApplicationSort>("updated_at");
  const [order, setOrderState] = useState<ApplicationOrder>("desc");
  const [view, setView] = useState<ApplicationView>("grid");
  const [selectedApplicationID, setSelectedApplicationID] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [editApplicationID, setEditApplicationID] = useState<string | null>(null);
  const [deleteApplicationID, setDeleteApplicationID] = useState<string | null>(null);

  const debouncedSearch = useDebouncedValue(search, 300);

  const setSearch = (value: string) => {
    setSearchState(value);
    setPage(1);
  };

  const setLimit = (value: number) => {
    setLimitState(value);
    setPage(1);
  };

  const setSort = (value: ApplicationSort) => {
    setSortState(value);
    setPage(1);
  };

  const setOrder = (value: ApplicationOrder) => {
    setOrderState(value);
    setPage(1);
  };

  const openCreate = () => setCreateOpen(true);
  const closeCreate = () => setCreateOpen(false);

  const openEdit = (id: string) => setEditApplicationID(id);
  const closeEdit = () => setEditApplicationID(null);

  const openDelete = (id: string) => setDeleteApplicationID(id);
  const closeDelete = () => setDeleteApplicationID(null);

  const openDetails = (id: string) => setSelectedApplicationID(id);
  const closeDetails = () => setSelectedApplicationID(null);

  return {
    search,
    debouncedSearch,
    page,
    limit,
    sort,
    order,
    view,
    selectedApplicationID,
    createOpen,
    editApplicationID,
    deleteApplicationID,
    setSearch,
    setPage,
    setLimit,
    setSort,
    setOrder,
    setView,
    openCreate,
    closeCreate,
    openEdit,
    closeEdit,
    openDelete,
    closeDelete,
    openDetails,
    closeDetails,
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
