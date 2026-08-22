"use client";

import { useEffect, useState } from "react";

import type { InvitationFilters, InvitationView } from "./types";

const initialFilters: InvitationFilters = { query: "", status: "all", sort: "created_at", order: "desc" };

export function useInvitationsWorkspace() {
  const [filters, setFilters] = useState<InvitationFilters>(initialFilters);
  const [view, setView] = useState<InvitationView>("table");
  const [page, setPage] = useState(1);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [revokeTargetId, setRevokeTargetId] = useState<string | null>(null);
  const debouncedSearch = useDebouncedValue(filters.query, 300);

  const updateFilters = (next: Partial<InvitationFilters>) => {
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
    inviteOpen,
    setInviteOpen,
    revokeTargetId,
    setRevokeTargetId,
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
