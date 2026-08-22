"use client";

import { useMemo } from "react";

import { ErrorState } from "@/components/common";
import { SectionCard } from "@/components/dashboard";
import { useInvitations } from "@/hooks/use-invitations";
import { PAGINATION_DEFAULT_PAGE_SIZE } from "@/lib/constants";

import { useInvitationsWorkspace } from "./hooks";
import { InvitationCard } from "./invitation-card";
import { InvitationEmpty } from "./invitation-empty";
import { InvitationSkeleton } from "./invitation-skeleton";
import { InvitationTable } from "./invitation-table";
import { InvitationToolbar } from "./invitation-toolbar";
import { InviteUserModal } from "./invite-user-modal";
import { RevokeInvitationDialog } from "./revoke-invitation-dialog";
import type { Invitation } from "./types";

/**
 * Invitations section for the Organization page. Renders as a standalone
 * SectionCard so it slots into the existing Organization workspace layout
 * without introducing its own page header or outer container.
 */
export function InvitationWorkspace() {
  const workspace = useInvitationsWorkspace();

  // The backend's InvitationQueryParams only supports search/sort/order/page/limit —
  // the invitation repository never applies a status filter server-side, so status is
  // refined client-side over the fetched page rather than sent as a query param.
  const queryParams = useMemo(
    () => ({
      page: workspace.page,
      limit: PAGINATION_DEFAULT_PAGE_SIZE,
      search: workspace.debouncedSearch.trim() || undefined,
      sort: workspace.filters.sort,
      order: workspace.filters.order,
    }),
    [workspace.page, workspace.debouncedSearch, workspace.filters.sort, workspace.filters.order]
  );

  const { data, error, isError, isLoading, refetch } = useInvitations(queryParams);

  const invitations = useMemo(() => {
    const items = data?.items ?? [];
    return workspace.filters.status === "all"
      ? items
      : items.filter((invitation) => invitation.status === workspace.filters.status);
  }, [data?.items, workspace.filters.status]);
  const hasFilters = workspace.filters.query.length > 0 || workspace.filters.status !== "all";
  const clearFilters = () => workspace.updateFilters({ query: "", status: "all" });
  const errorMessage = error instanceof Error ? error.message : "Unable to load invitations. Please try again.";

  const revokeTarget: Invitation | null =
    invitations.find((invitation) => invitation.id === workspace.revokeTargetId) ?? null;

  return (
    <SectionCard
      title="Invitations"
      description="Invite new members and manage pending invitations for this organization."
    >
      <InvitationToolbar
        filters={workspace.filters}
        view={workspace.view}
        onFiltersChange={workspace.updateFilters}
        onViewChange={workspace.setView}
        onInvite={() => workspace.setInviteOpen(true)}
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
          <InvitationSkeleton view={workspace.view} />
        ) : invitations.length === 0 ? (
          <InvitationEmpty
            hasFilters={hasFilters}
            onClear={clearFilters}
            onInvite={() => workspace.setInviteOpen(true)}
          />
        ) : workspace.view === "grid" ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {invitations.map((invitation) => (
              <InvitationCard
                key={invitation.id}
                invitation={invitation}
                onRevoke={(target) => workspace.setRevokeTargetId(target.id)}
              />
            ))}
          </div>
        ) : (
          <InvitationTable
            invitations={invitations}
            onRevoke={(target) => workspace.setRevokeTargetId(target.id)}
          />
        )}
      </div>

      {data && data.total > 0 ? (
        <nav
          aria-label="Invitations pagination"
          className="mt-6 flex flex-col gap-3 border-t pt-5 sm:flex-row sm:items-center sm:justify-between"
          style={{ borderColor: "var(--border)" }}
        >
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
            Showing {Math.min((data.page - 1) * data.limit + 1, data.total)}–{Math.min(data.page * data.limit, data.total)} of {data.total} invitations
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

      <InviteUserModal open={workspace.inviteOpen} onClose={() => workspace.setInviteOpen(false)} />
      <RevokeInvitationDialog
        open={workspace.revokeTargetId !== null}
        invitation={revokeTarget}
        onClose={() => workspace.setRevokeTargetId(null)}
      />
    </SectionCard>
  );
}
