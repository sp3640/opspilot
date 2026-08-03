"use client";
import { useEffect, useState } from "react";
import type { MetricFilters, MetricView } from "./types";
const initial: MetricFilters = { query: "", projectId: "all", clusterId: "all", resourceId: "all", metricType: "all", timeRange: "24h", sort: "timestamp", order: "desc" };
export function useMetricsWorkspace() { const [filters, setFilters] = useState(initial); const [view, setView] = useState<MetricView>("charts"); const [page, setPage] = useState(1); const [selectedMetricID, setSelectedMetricID] = useState<string | null>(null); const [debouncedSearch, setSearch] = useState(""); useEffect(() => { const id = window.setTimeout(() => setSearch(filters.query), 300); return () => window.clearTimeout(id); }, [filters.query]); return { filters, updateFilters: (next: Partial<MetricFilters>) => { setFilters((current) => ({ ...current, ...next })); setPage(1); }, view, setView, page, setPage, selectedMetricID, setSelectedMetricID, debouncedSearch }; }
