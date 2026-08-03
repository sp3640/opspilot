"use client";
import { useEffect, useState } from "react";
import type { ResourceFilters, ResourceView } from "./types";
const initialFilters: ResourceFilters = { query: "", projectId: "all", cluster: "all", namespace: "all", provider: "all", kind: "all", status: "all", health: "all", region: "all", sort: "updated_at", order: "desc" };
export function useResourcesWorkspace() {
 const [filters, setFilters] = useState<ResourceFilters>(initialFilters); const [view, setView] = useState<ResourceView>("grid"); const [page, setPage] = useState(1); const [selectedResourceID, setSelectedResourceID] = useState<string | null>(null); const [createOpen, setCreateOpen] = useState(false); const [syncOpen, setSyncOpen] = useState(false); const debouncedSearch = useDebouncedValue(filters.query, 300);
 return { filters, updateFilters: (next: Partial<ResourceFilters>) => { setFilters((current) => ({ ...current, ...next })); setPage(1); }, view, setView, page, setPage, selectedResourceID, setSelectedResourceID, createOpen, setCreateOpen, syncOpen, setSyncOpen, debouncedSearch };
}
function useDebouncedValue(value: string, delay: number) { const [result, setResult] = useState(value); useEffect(() => { const timeout = window.setTimeout(() => setResult(value), delay); return () => window.clearTimeout(timeout); }, [value, delay]); return result; }
