"use client";

import { useMemo } from "react";
import { Plus, Rocket } from "lucide-react";

import { EmptyState, ErrorState, PageHeader } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { Button } from "@/components/ui/button";
import { useApplicationsByProject } from "@/hooks/use-applications";
import { useHasPermission } from "@/store/auth-store";

import { ApplicationDetailsDrawer } from "./application-details-drawer";
import { ApplicationGrid } from "./application-grid";
import { ApplicationTable } from "./application-table";
import { ApplicationToolbar } from "./application-toolbar";
import { CreateApplicationModal } from "./create-application-modal";
import { DeleteApplicationDialog } from "./delete-application-dialog";
import { EditApplicationModal } from "./edit-application-modal";
import { useApplicationsWorkspace } from "./hooks";

type ApplicationWorkspaceProps = {
  projectId: string;
};

export function ApplicationWorkspace({ projectId }: ApplicationWorkspaceProps) {
  const canManageApplications = useHasPermission("application:manage");
  const workspace = useApplicationsWorkspace();

  const params = useMemo(
    () => ({
      page: workspace.page,
      limit: workspace.limit,
      search: workspace.debouncedSearch.trim() || undefined,
      sort: workspace.sort,
      order: workspace.order,
    }),
    [workspace.page, workspace.limit, workspace.debouncedSearch, workspace.sort, workspace.order]
  );

  const { data, error, isError, isLoading, refetch } = useApplicationsByProject(projectId, params);

  const applications = useMemo(() => data?.items ?? [], [data?.items]);

  const editApplication = useMemo(
    () => applications.find((application) => application.id === workspace.editApplicationID) ?? null,
    [applications, workspace.editApplicationID]
  );
  const deleteApplication = useMemo(
    () => applications.find((application) => application.id === workspace.deleteApplicationID) ?? null,
    [applications, workspace.deleteApplicationID]
  );
  const selectedApplication = useMemo(
    () => applications.find((application) => application.id === workspace.selectedApplicationID) ?? null,
    [applications, workspace.selectedApplicationID]
  );

  const errorMessage = error instanceof Error ? error.message : "Unable to load applications. Please try again.";

  return (
    <div className="mx-auto max-w-[1600px] space-y-6 lg:space-y-8">
      <PageHeader
        title="Applications"
        description="Manage the services, runtimes, and deployment configuration for this project."
        breadcrumb={[{ label: "Overview", href: "/" }, { label: "Applications" }]}
      />

      <SectionCard>
        <ApplicationToolbar
          search={workspace.search}
          onSearchChange={workspace.setSearch}
          view={workspace.view}
          onViewChange={workspace.setView}
          onCreate={workspace.openCreate}
        />

        <div className="mt-6">
          {isError ? (
            <ErrorState
              description={errorMessage}
              onRetry={() => {
                void refetch();
              }}
            />
          ) : isLoading ? (
            workspace.view === "grid" ? (
              <ApplicationGrid applications={[]} loading />
            ) : (
              <ApplicationTable applications={[]} loading />
            )
          ) : applications.length === 0 ? (
            <EmptyState
              icon={Rocket}
              title="No applications yet"
              description="Create your first application to define its runtime and deployment configuration."
              action={
                canManageApplications ? (
                  <Button type="button" onClick={workspace.openCreate}>
                    <Plus aria-hidden="true" className="h-4 w-4" />
                    Create application
                  </Button>
                ) : undefined
              }
            />
          ) : workspace.view === "grid" ? (
            <ApplicationGrid
              applications={applications}
              onCardClick={(application) => workspace.openDetails(application.id)}
              onEdit={canManageApplications ? (application) => workspace.openEdit(application.id) : undefined}
              onDelete={canManageApplications ? (application) => workspace.openDelete(application.id) : undefined}
            />
          ) : (
            <ApplicationTable
              applications={applications}
              onRowClick={(application) => workspace.openDetails(application.id)}
              onEdit={canManageApplications ? (application) => workspace.openEdit(application.id) : undefined}
              onDelete={canManageApplications ? (application) => workspace.openDelete(application.id) : undefined}
            />
          )}
        </div>

        {data && data.total > 0 ? (
          <nav
            aria-label="Applications pagination"
            className="mt-6 flex flex-col gap-3 border-t pt-5 sm:flex-row sm:items-center sm:justify-between"
            style={{ borderColor: "var(--border)" }}
          >
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
              Showing {Math.min((data.page - 1) * data.limit + 1, data.total)}-{Math.min(data.page * data.limit, data.total)} of {data.total} applications
            </p>
            <div className="flex items-center gap-2">
              <button
                type="button"
                disabled={data.page === 1}
                onClick={() => workspace.setPage(data.page - 1)}
                className="rounded-xl border px-3 py-2 text-sm font-medium disabled:opacity-40"
                style={{ borderColor: "var(--border)" }}
              >
                Previous
              </button>
              <span className="px-2 text-sm tabular-nums" style={{ color: "var(--muted-foreground)" }}>
                Page {data.page} of {data.totalPages}
              </span>
              <button
                type="button"
                disabled={data.page === data.totalPages}
                onClick={() => workspace.setPage(data.page + 1)}
                className="rounded-xl border px-3 py-2 text-sm font-medium disabled:opacity-40"
                style={{ borderColor: "var(--border)" }}
              >
                Next
              </button>
            </div>
          </nav>
        ) : null}
      </SectionCard>

      <CreateApplicationModal
        open={workspace.createOpen}
        projectId={projectId}
        onClose={workspace.closeCreate}
      />

      <EditApplicationModal
        open={workspace.editApplicationID !== null}
        application={editApplication}
        onClose={workspace.closeEdit}
      />

      <DeleteApplicationDialog
        open={workspace.deleteApplicationID !== null}
        application={deleteApplication}
        onClose={workspace.closeDelete}
      />

      <ApplicationDetailsDrawer
        open={workspace.selectedApplicationID !== null && selectedApplication !== null}
        application={selectedApplication}
        onClose={workspace.closeDetails}
      />
    </div>
  );
}
