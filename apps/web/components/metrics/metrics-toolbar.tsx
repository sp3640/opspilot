"use client";
import { BarChart3, LayoutGrid, List, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { ClusterResponse } from "@/types/cluster-api";
import type { ProjectResponse } from "@/types/project-api";
import { MetricsFilter } from "./metrics-filter";
import { MetricsSearch } from "./metrics-search";
import type { MetricFilters, MetricView } from "./types";
export function MetricsToolbar({ filters, view, projects, clusters, onChange, onView, onRefresh, refreshing }: { filters: MetricFilters; view: MetricView; projects: ProjectResponse[]; clusters: ClusterResponse[]; onChange: (x: Partial<MetricFilters>) => void; onView: (x: MetricView) => void; onRefresh: () => void; refreshing: boolean }) { return <div className="space-y-3"><div className="flex flex-col gap-3 lg:flex-row"><MetricsSearch value={filters.query} onChange={(query) => onChange({ query })} /><MetricsFilter filters={filters} projects={projects} clusters={clusters} onChange={onChange} /></div><div className="flex items-center justify-between"><div className="flex rounded-2xl border p-1" style={{ borderColor: "var(--border)", backgroundColor: "var(--muted)" }}>{([{ v: "charts", Icon: BarChart3 }, { v: "cards", Icon: LayoutGrid }, { v: "table", Icon: List }] as const).map(({ v, Icon }) => <button key={v} type="button" onClick={() => onView(v)} aria-pressed={v === view} className="rounded-xl p-2" style={{ backgroundColor: view === v ? "var(--card)" : "transparent" }}><Icon className="h-4 w-4" /></button>)}</div><Button type="button" variant="ghost" loading={refreshing} onClick={onRefresh}><RefreshCw className="h-4 w-4" />Refresh</Button></div></div>; }
