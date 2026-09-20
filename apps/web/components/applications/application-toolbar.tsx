"use client";

import { Grid2X2, List, Plus } from "lucide-react";

import { Toolbar, ToolbarSearch, ViewToggle, controlClass, controlStyle } from "@/components/common/toolbar";
import { Button } from "@/components/ui/button";
import { useHasPermission } from "@/store/auth-store";

type ApplicationView = "grid" | "table";

type ApplicationToolbarProps = {
  search: string;
  onSearchChange: (value: string) => void;
  runtime?: string;
  onRuntimeChange?: (value: string) => void;
  status?: string;
  onStatusChange?: (value: string) => void;
  view: ApplicationView;
  onViewChange: (view: ApplicationView) => void;
  onCreate: () => void;
};

const RUNTIME_OPTIONS = ["NodeJS", "Go", "Java", "Python", "DotNet", "Static", "Docker"] as const;
const STATUS_OPTIONS = ["Draft", "Ready", "Archived"] as const;

export function ApplicationToolbar({
  search,
  onSearchChange,
  runtime,
  onRuntimeChange,
  status,
  onStatusChange,
  view,
  onViewChange,
  onCreate,
}: ApplicationToolbarProps) {
  const canManageApplications = useHasPermission("application:manage");
  const showFilters = Boolean(onRuntimeChange || onStatusChange);

  return (
    <Toolbar
      search={
        <ToolbarSearch
          id="application-search"
          label="Search applications"
          placeholder="Search applications"
          clearLabel="Clear application search"
          value={search}
          onChange={onSearchChange}
        />
      }
      actions={
        <>
          <ViewToggle
            label="Application view"
            value={view}
            onChange={onViewChange}
            options={[
              { value: "grid", label: "Grid", icon: Grid2X2 },
              { value: "table", label: "Table", icon: List },
            ]}
          />
          {canManageApplications ? (
            <Button type="button" onClick={onCreate}>
              <Plus aria-hidden="true" className="h-4 w-4" />
              Create application
            </Button>
          ) : null}
        </>
      }
      filters={
        showFilters ? (
          <>
            {onRuntimeChange && (
              <select aria-label="Runtime" className={controlClass} style={controlStyle} value={runtime ?? "all"} onChange={(event) => onRuntimeChange(event.target.value)}>
                <option value="all">All runtimes</option>
                {RUNTIME_OPTIONS.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
            )}
            {onStatusChange && (
              <select aria-label="Status" className={controlClass} style={controlStyle} value={status ?? "all"} onChange={(event) => onStatusChange(event.target.value)}>
                <option value="all">All statuses</option>
                {STATUS_OPTIONS.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
            )}
          </>
        ) : undefined
      }
    />
  );
}
