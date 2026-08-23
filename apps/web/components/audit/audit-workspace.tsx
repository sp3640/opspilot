"use client";

import { useEffect, useMemo, useState } from "react";
import { ClipboardList, Search, X } from "lucide-react";
import { useAuditLogs } from "@/hooks/use-audit";
import { useIncidents } from "@/hooks/use-incidents";
import { useProjects } from "@/hooks/use-projects";
import { useApplications } from "@/hooks/use-applications";
import {
  PAGINATION_DEFAULT_PAGE,
  PAGINATION_DEFAULT_PAGE_SIZE,
  PROJECT_LOOKUP_QUERY,
  SORT_ORDERS,
} from "@/lib/constants/pagination";
import { ErrorState, EmptyState, PageHeader, StatusBadge, TableSkeleton } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { Button } from "@/components/ui/button";
import type {
  AuditAction,
  AuditLogResponse,
  AuditQueryParams,
  AuditResult,
  AuditScope,
} from "@/types/audit-api";

type ViewScope = Extract<AuditScope, "organization" | "project" | "incident">;

type AuditFilters = {
  scope: ViewScope;
  action: "all" | AuditAction;
  result: "all" | AuditResult;
  entityType: string;
  applicationId: string;
  userId: string;
  dateFrom: string;
  dateTo: string;
  search: string;
  sort: "created_at" | "entity_type" | "action";
  order: "asc" | "desc";
};

const initialFilters: AuditFilters = {
  scope: "organization",
  action: "all",
  result: "all",
  entityType: "",
  applicationId: "",
  userId: "",
  dateFrom: "",
  dateTo: "",
  search: "",
  sort: "created_at",
  order: SORT_ORDERS.DESC,
};

// Local <input type="datetime-local"> values have no timezone; the audit
// API expects RFC3339, so treat the picker value as local time and convert.
function toRFC3339(localDateTime: string): string | undefined {
  if (!localDateTime) return undefined;
  const parsed = new Date(localDateTime);
  if (Number.isNaN(parsed.getTime())) return undefined;
  return parsed.toISOString();
}

export function AuditWorkspace() {
  const [filters, setFilters] = useState(initialFilters);
  const [page, setPage] = useState(PAGINATION_DEFAULT_PAGE);
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);
  const [selectedIncidentId, setSelectedIncidentId] = useState<number | null>(null);
  const [selectedLog, setSelectedLog] = useState<AuditLogResponse | null>(null);

  const { data: projectsData, isLoading: isProjectsLoading } = useProjects({
    page: PROJECT_LOOKUP_QUERY.page,
    limit: PROJECT_LOOKUP_QUERY.limit,
    sort: PROJECT_LOOKUP_QUERY.sort,
    order: PROJECT_LOOKUP_QUERY.order,
  });

  const { data: incidentsData, isLoading: isIncidentsLoading } = useIncidents({
    page: 1,
    limit: 100,
    sort: "updated_at",
    order: "desc",
  });

  const { data: applicationsData, isLoading: isApplicationsLoading } = useApplications({
    page: PROJECT_LOOKUP_QUERY.page,
    limit: PROJECT_LOOKUP_QUERY.limit,
    sort: "name",
    order: SORT_ORDERS.ASC,
  });

  const projects = useMemo(() => projectsData?.items ?? [], [projectsData?.items]);
  const incidents = useMemo(() => incidentsData?.items ?? [], [incidentsData?.items]);
  const applications = useMemo(() => applicationsData?.items ?? [], [applicationsData?.items]);

  useEffect(() => {
    if (filters.scope === "project" && !selectedProjectId && projects.length > 0) {
      setSelectedProjectId(projects[0]?.id ?? null);
    }
  }, [filters.scope, projects, selectedProjectId]);

  useEffect(() => {
    if (filters.scope === "incident" && selectedIncidentId === null && incidents.length > 0) {
      setSelectedIncidentId(incidents[0]?.id ?? null);
    }
  }, [filters.scope, incidents, selectedIncidentId]);

  const scopeID =
    filters.scope === "project" ? selectedProjectId : filters.scope === "incident" ? selectedIncidentId : null;

  const queryParams: AuditQueryParams = useMemo(
    () => ({
      page,
      limit: PAGINATION_DEFAULT_PAGE_SIZE,
      search: filters.search.trim() || undefined,
      sort: filters.sort,
      order: filters.order,
      action: filters.action === "all" ? undefined : filters.action,
      result: filters.result === "all" ? undefined : filters.result,
      entityType: filters.entityType.trim() || undefined,
      applicationId: filters.applicationId || undefined,
      userId: filters.userId.trim() ? Number(filters.userId.trim()) : undefined,
      dateFrom: toRFC3339(filters.dateFrom),
      dateTo: toRFC3339(filters.dateTo),
    }),
    [
      filters.action,
      filters.applicationId,
      filters.dateFrom,
      filters.dateTo,
      filters.entityType,
      filters.order,
      filters.result,
      filters.search,
      filters.sort,
      filters.userId,
      page,
    ]
  );

  const {
    data: auditData,
    error,
    isError,
    isLoading,
    isFetching,
    refetch,
  } = useAuditLogs(filters.scope, scopeID, queryParams);

  const items = auditData?.items ?? [];

  const resetPage = () => setPage(PAGINATION_DEFAULT_PAGE);

  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <PageHeader
        title="Audit Logs"
        description="Trace operational and security-relevant actions across your organization, projects, and incidents."
        breadcrumb={[{ label: "Overview", href: "/" }, { label: "Audit Logs" }]}
      />

      <SectionCard>
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex flex-wrap items-center gap-2">
            <FilterSelect
              id="audit-scope"
              value={filters.scope}
              onChange={(value) => {
                setFilters((current) => ({ ...current, scope: value as ViewScope }));
                resetPage();
              }}
            >
              <option value="organization">Organization audit</option>
              <option value="project">Project scope</option>
              <option value="incident">Incident scope</option>
            </FilterSelect>

            {filters.scope === "project" ? (
              <FilterSelect
                id="audit-project"
                value={selectedProjectId ?? ""}
                onChange={(value) => {
                  setSelectedProjectId(value || null);
                  resetPage();
                }}
                disabled={isProjectsLoading || projects.length === 0}
              >
                {projects.length === 0 ? (
                  <option value="">No projects available</option>
                ) : (
                  projects.map((project) => (
                    <option key={project.id} value={project.id}>
                      {project.name}
                    </option>
                  ))
                )}
              </FilterSelect>
            ) : (
              <FilterSelect
                id="audit-incident"
                value={selectedIncidentId?.toString() ?? ""}
                onChange={(value) => {
                  setSelectedIncidentId(value ? Number(value) : null);
                  resetPage();
                }}
                disabled={isIncidentsLoading || incidents.length === 0}
              >
                {incidents.length === 0 ? (
                  <option value="">No incidents available</option>
                ) : (
                  incidents.map((incident) => (
                    <option key={incident.id} value={incident.id}>
                      #{incident.id} {incident.title}
                    </option>
                  ))
                )}
              </FilterSelect>
            )}

            <FilterSelect
              id="audit-action"
              value={filters.action}
              onChange={(value) => {
                setFilters((current) => ({ ...current, action: value as AuditFilters["action"] }));
                resetPage();
              }}
            >
              <option value="all">All actions</option>
              <option value="CREATE">Create</option>
              <option value="UPDATE">Update</option>
              <option value="DELETE">Delete</option>
              <option value="LOGIN">Login</option>
            </FilterSelect>

            <FilterSelect
              id="audit-result"
              value={filters.result}
              onChange={(value) => {
                setFilters((current) => ({ ...current, result: value as AuditFilters["result"] }));
                resetPage();
              }}
            >
              <option value="all">All results</option>
              <option value="SUCCESS">Success</option>
              <option value="FAILURE">Failure</option>
            </FilterSelect>

            <FilterSelect
              id="audit-application"
              value={filters.applicationId}
              onChange={(value) => {
                setFilters((current) => ({ ...current, applicationId: value }));
                resetPage();
              }}
              disabled={isApplicationsLoading}
            >
              <option value="">All applications</option>
              {applications.map((application) => (
                <option key={application.id} value={application.id}>
                  {application.name}
                </option>
              ))}
            </FilterSelect>

            <FilterSelect
              id="audit-sort"
              value={filters.sort}
              onChange={(value) => {
                setFilters((current) => ({ ...current, sort: value as AuditFilters["sort"] }));
                resetPage();
              }}
            >
              <option value="created_at">Sort by time</option>
              <option value="entity_type">Sort by entity</option>
              <option value="action">Sort by action</option>
            </FilterSelect>

            <FilterSelect
              id="audit-order"
              value={filters.order}
              onChange={(value) => {
                setFilters((current) => ({ ...current, order: value as AuditFilters["order"] }));
                resetPage();
              }}
            >
              <option value="desc">Newest first</option>
              <option value="asc">Oldest first</option>
            </FilterSelect>
          </div>

          <div className="flex flex-1 items-center gap-2 lg:max-w-lg">
            <div className="relative flex-1">
              <Search
                aria-hidden="true"
                className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2"
                style={{ color: "var(--muted-foreground)" }}
              />
              <input
                value={filters.search}
                onChange={(event) => {
                  setFilters((current) => ({ ...current, search: event.target.value }));
                  resetPage();
                }}
                placeholder="Search entity type or action"
                className="h-11 w-full rounded-2xl border bg-transparent py-2 pl-10 pr-10 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]"
                style={{ borderColor: "var(--border)" }}
              />
              {filters.search && (
                <button
                  type="button"
                  onClick={() => {
                    setFilters((current) => ({ ...current, search: "" }));
                    resetPage();
                  }}
                  className="absolute right-2 top-1/2 rounded-lg p-1 -translate-y-1/2 hover:bg-[var(--muted)]"
                  aria-label="Clear audit search"
                >
                  <X aria-hidden="true" className="h-4 w-4" />
                </button>
              )}
            </div>
            <input
              value={filters.entityType}
              onChange={(event) => {
                setFilters((current) => ({ ...current, entityType: event.target.value }));
                resetPage();
              }}
              placeholder="Resource type"
              className="h-11 w-40 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]"
              style={{ borderColor: "var(--border)" }}
            />
            <Button
              type="button"
              variant="ghost"
              loading={isFetching}
              onClick={() => {
                void refetch();
              }}
            >
              Refresh
            </Button>
          </div>
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-2">
          <input
            value={filters.userId}
            onChange={(event) => {
              setFilters((current) => ({ ...current, userId: event.target.value.replace(/\D/g, "") }));
              resetPage();
            }}
            placeholder="User ID"
            inputMode="numeric"
            className="h-11 w-28 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]"
            style={{ borderColor: "var(--border)" }}
          />
          <label className="flex items-center gap-2 text-sm" style={{ color: "var(--muted-foreground)" }}>
            From
            <input
              type="datetime-local"
              value={filters.dateFrom}
              onChange={(event) => {
                setFilters((current) => ({ ...current, dateFrom: event.target.value }));
                resetPage();
              }}
              className="h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]"
              style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
            />
          </label>
          <label className="flex items-center gap-2 text-sm" style={{ color: "var(--muted-foreground)" }}>
            To
            <input
              type="datetime-local"
              value={filters.dateTo}
              onChange={(event) => {
                setFilters((current) => ({ ...current, dateTo: event.target.value }));
                resetPage();
              }}
              className="h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]"
              style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
            />
          </label>
          {(filters.userId || filters.dateFrom || filters.dateTo) && (
            <button
              type="button"
              onClick={() => {
                setFilters((current) => ({ ...current, userId: "", dateFrom: "", dateTo: "" }));
                resetPage();
              }}
              className="rounded-lg px-2 py-1 text-xs hover:bg-[var(--muted)]"
              style={{ color: "var(--muted-foreground)" }}
            >
              Clear
            </button>
          )}
        </div>

        <div className="mt-6">
          {filters.scope !== "organization" && scopeID === null ? (
            <EmptyState
              icon={ClipboardList}
              title="Scope is required"
              description={`Select a ${filters.scope} to load audit logs.`}
            />
          ) : isError ? (
            <ErrorState
              description={error instanceof Error ? error.message : "Unable to load audit logs."}
              onRetry={() => {
                void refetch();
              }}
            />
          ) : isLoading ? (
            <TableSkeleton columns={7} rows={6} />
          ) : items.length === 0 ? (
            <EmptyState
              icon={ClipboardList}
              title="No audit logs found"
              description="Try adjusting your search or filter criteria."
            />
          ) : (
            <AuditTable items={items} onSelect={setSelectedLog} />
          )}
        </div>

        {auditData && auditData.total > 0 && (
          <Pagination
            page={auditData.page}
            pageCount={auditData.totalPages}
            total={auditData.total}
            limit={auditData.limit}
            onPageChange={setPage}
          />
        )}
      </SectionCard>

      <AuditDetailDrawer log={selectedLog} onClose={() => setSelectedLog(null)} />
    </div>
  );
}

function FilterSelect({
  id,
  value,
  onChange,
  disabled,
  children,
}: {
  id: string;
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  children: React.ReactNode;
}) {
  return (
    <select
      id={id}
      value={value}
      onChange={(event) => onChange(event.target.value)}
      disabled={disabled}
      className="h-11 rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)] disabled:opacity-60"
      style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
    >
      {children}
    </select>
  );
}

function AuditTable({
  items,
  onSelect,
}: {
  items: AuditLogResponse[];
  onSelect: (item: AuditLogResponse) => void;
}) {
  return (
    <div className="overflow-x-auto rounded-2xl border" style={{ borderColor: "var(--border)" }}>
      <table className="w-full min-w-[860px] text-left text-sm">
        <thead
          className="text-xs uppercase tracking-[0.1em]"
          style={{
            color: "var(--muted-foreground)",
            backgroundColor: "color-mix(in srgb, var(--muted) 55%, transparent)",
          }}
        >
          <tr>
            <th className="px-5 py-3 font-semibold">Time</th>
            <th className="px-5 py-3 font-semibold">Action</th>
            <th className="px-5 py-3 font-semibold">Result</th>
            <th className="px-5 py-3 font-semibold">Resource</th>
            <th className="px-5 py-3 font-semibold">Resource ID</th>
            <th className="px-5 py-3 font-semibold">Field</th>
            <th className="px-5 py-3 font-semibold">User</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr
              key={item.id}
              role="button"
              tabIndex={0}
              className="cursor-pointer border-t transition-colors hover:bg-[var(--muted)] focus:bg-[var(--muted)] focus:outline-none"
              style={{ borderColor: "var(--border)" }}
              onClick={() => onSelect(item)}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onSelect(item);
                }
              }}
            >
              <td className="px-5 py-4">{new Date(item.created_at).toLocaleString()}</td>
              <td className="px-5 py-4">
                <StatusBadge variant={badgeVariantForAction(item.action)}>{item.action}</StatusBadge>
              </td>
              <td className="px-5 py-4">
                <StatusBadge variant={item.result === "FAILURE" ? "critical" : "success"}>{item.result}</StatusBadge>
              </td>
              <td className="px-5 py-4">{item.entity_type}</td>
              <td className="px-5 py-4">{item.entity_id}</td>
              <td className="px-5 py-4">{item.field_name || "—"}</td>
              <td className="px-5 py-4">{item.user_id}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function Pagination({
  page,
  pageCount,
  total,
  limit,
  onPageChange,
}: {
  page: number;
  pageCount: number;
  total: number;
  limit: number;
  onPageChange: (next: number) => void;
}) {
  return (
    <nav
      aria-label="Audit pagination"
      className="mt-6 flex flex-col gap-3 border-t pt-5 sm:flex-row sm:items-center sm:justify-between"
      style={{ borderColor: "var(--border)" }}
    >
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        Showing {Math.min((page - 1) * limit + 1, total)}–{Math.min(page * limit, total)} of {total} logs
      </p>
      <div className="flex items-center gap-2">
        <Button
          type="button"
          variant="secondary"
          onClick={() => onPageChange(page - 1)}
          disabled={page === 1}
          className="px-3 py-2"
        >
          Previous
        </Button>
        <span className="px-2 text-sm tabular-nums" style={{ color: "var(--muted-foreground)" }}>
          Page {page} of {pageCount}
        </span>
        <Button
          type="button"
          variant="secondary"
          onClick={() => onPageChange(page + 1)}
          disabled={page === pageCount}
          className="px-3 py-2"
        >
          Next
        </Button>
      </div>
    </nav>
  );
}

function AuditDetailDrawer({
  log,
  onClose,
}: {
  log: AuditLogResponse | null;
  onClose: () => void;
}) {
  if (!log) return null;

  return (
    <div
      className="fixed inset-0 z-[60] flex justify-end bg-[color:color-mix(in_srgb,var(--background)_72%,transparent)]"
      role="presentation"
      onMouseDown={onClose}
    >
      <aside
        role="dialog"
        aria-modal="true"
        aria-label="Audit log details"
        onMouseDown={(event) => event.stopPropagation()}
        className="flex h-full w-full max-w-xl flex-col border-l shadow-[var(--shadow-lg)]"
        style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
      >
        <header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div>
            <h2 className="text-lg font-semibold">Audit Log #{log.id}</h2>
            <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
              {new Date(log.created_at).toLocaleString()}
            </p>
          </div>
          <Button type="button" variant="ghost" onClick={onClose} className="h-9 w-9 rounded-xl p-0" aria-label="Close audit detail">
            <X aria-hidden="true" className="h-4 w-4" />
          </Button>
        </header>
        <div className="space-y-4 overflow-y-auto p-5">
          <DetailField label="Action" value={log.action} />
          <DetailField label="Result" value={log.result} />
          <DetailField label="Resource Type" value={log.entity_type} />
          <DetailField label="Resource ID" value={log.entity_id} />
          <DetailField label="Field" value={log.field_name || "—"} />
          <DetailField label="User ID" value={String(log.user_id)} />
          <DetailField label="Project ID" value={log.project_id || "—"} />
          <DetailField label="Application ID" value={log.application_id || "—"} />
          <DetailField label="Incident ID" value={log.incident_id ? String(log.incident_id) : "—"} />
          <DetailField label="Old Value" value={log.old_value || "—"} multiline />
          <DetailField label="New Value" value={log.new_value || "—"} multiline />
          <DetailField label="Before State" value={log.before_state || "—"} multiline />
          <DetailField label="After State" value={log.after_state || "—"} multiline />
          <DetailField label="IP Address" value={log.ip_address || "—"} />
          <DetailField label="User Agent" value={log.user_agent || "—"} multiline />
        </div>
      </aside>
    </div>
  );
}

function DetailField({
  label,
  value,
  multiline = false,
}: {
  label: string;
  value: string;
  multiline?: boolean;
}) {
  return (
    <div
      className="rounded-2xl border p-4"
      style={{
        borderColor: "var(--border)",
        backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)",
      }}
    >
      <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        {label}
      </p>
      <p className={`mt-1 font-medium ${multiline ? "whitespace-pre-wrap break-all" : "break-all"}`}>{value}</p>
    </div>
  );
}

function badgeVariantForAction(action: AuditAction): "success" | "warning" | "critical" {
  switch (action) {
    case "CREATE":
      return "success";
    case "DELETE":
      return "critical";
    default:
      return "warning";
  }
}
