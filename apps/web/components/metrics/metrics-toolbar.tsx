"use client";

import { BarChart3, LayoutGrid, List, RefreshCw } from "lucide-react";

import { Toolbar, ViewToggle } from "@/components/common/toolbar";
import { Button } from "@/components/ui/button";
import type { ClusterResponse } from "@/types/cluster-api";
import type { ProjectResponse } from "@/types/project-api";

import { MetricsFilter } from "./metrics-filter";
import { MetricsSearch } from "./metrics-search";
import type { MetricFilters, MetricView } from "./types";

type MetricsToolbarProps = {
  filters: MetricFilters;
  view: MetricView;
  projects: ProjectResponse[];
  clusters: ClusterResponse[];
  onChange: (x: Partial<MetricFilters>) => void;
  onView: (x: MetricView) => void;
  onRefresh: () => void;
  refreshing: boolean;
};

export function MetricsToolbar({ filters, view, projects, clusters, onChange, onView, onRefresh, refreshing }: MetricsToolbarProps) {
  return (
    <Toolbar
      search={<MetricsSearch value={filters.query} onChange={(query) => onChange({ query })} />}
      actions={
        <>
          <ViewToggle
            label="Metrics view"
            value={view}
            onChange={onView}
            options={[
              { value: "charts", label: "Charts", icon: BarChart3 },
              { value: "cards", label: "Cards", icon: LayoutGrid },
              { value: "table", label: "Table", icon: List },
            ]}
          />
          <Button type="button" variant="ghost" loading={refreshing} onClick={onRefresh} className="px-3" aria-label="Refresh metrics">
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>
        </>
      }
      filters={<MetricsFilter filters={filters} projects={projects} clusters={clusters} onChange={onChange} />}
    />
  );
}
