"use client";

import { Plus, RefreshCw } from "lucide-react";

import { Toolbar } from "@/components/common/toolbar";
import { Button } from "@/components/ui/button";
import { useHasPermission } from "@/store/auth-store";

import { IncidentSearch } from "./incident-search";
import type { IncidentFilters } from "./hooks";

type IncidentToolbarProps = {
  filters: IncidentFilters;
  onFiltersChange: (filters: Partial<IncidentFilters>) => void;
  onRefresh: () => void;
  onCreate: () => void;
  refreshing?: boolean;
};

export function IncidentToolbar({ filters, onFiltersChange, onRefresh, onCreate, refreshing = false }: IncidentToolbarProps) {
  const canManageIncidents = useHasPermission("incident:manage");
  return (
    <Toolbar
      search={<IncidentSearch value={filters.query} onChange={(query) => onFiltersChange({ query })} />}
      actions={
        <>
          <Button type="button" variant="ghost" onClick={onRefresh} loading={refreshing} className="px-3" aria-label="Refresh incidents">
            <RefreshCw aria-hidden="true" className="h-4 w-4" />
            <span className="hidden sm:inline">Refresh</span>
          </Button>
          {canManageIncidents ? (
            <Button type="button" onClick={onCreate}>
              <Plus aria-hidden="true" className="h-4 w-4" />
              Create incident
            </Button>
          ) : null}
        </>
      }
    />
  );
}
