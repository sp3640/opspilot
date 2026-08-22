"use client";

import { useEffect, useMemo, useState } from "react";
import axios from "axios";
import { FolderKanban, Layers3, Pencil, Trash2, UserPlus, Users, X } from "lucide-react";

import { EmptyState, ErrorState, ListSkeleton } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useApplication } from "@/hooks/use-applications";
import { useTeamApplications } from "@/hooks/use-application-teams";
import { useTeamProjects } from "@/hooks/use-project-teams";
import { useProjects } from "@/hooks/use-projects";
import { useAddTeamMember, useRemoveTeamMember, useTeam, useTeamMembers } from "@/hooks/use-teams";
import { useIsPlatformAdmin } from "@/store/auth-store";

import { DeleteTeamDialog } from "./delete-team-dialog";
import { EditTeamModal } from "./edit-team-modal";

const PROJECT_LOOKUP_QUERY = { page: 1, limit: 100, sort: "name" as const, order: "asc" as const };

export function TeamDetailsDrawer({
  teamID,
  onClose,
}: {
  teamID: string | null;
  onClose: () => void;
}) {
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [newMemberUserId, setNewMemberUserId] = useState("");
  const [memberError, setMemberError] = useState<string | null>(null);
  const isAdmin = useIsPlatformAdmin();

  const { data: team, error, isError, isLoading, refetch } = useTeam(teamID);
  const {
    data: membersData,
    error: membersErrorObj,
    isError: isMembersError,
    isLoading: isMembersLoading,
    refetch: refetchMembers,
  } = useTeamMembers(teamID);
  const addMember = useAddTeamMember();
  const removeMember = useRemoveTeamMember();

  const {
    data: ownedProjectsData,
    error: ownedProjectsErrorObj,
    isError: isOwnedProjectsError,
    isLoading: isOwnedProjectsLoading,
    refetch: refetchOwnedProjects,
  } = useTeamProjects(teamID);
  const { data: projectsLookupData } = useProjects(PROJECT_LOOKUP_QUERY, Boolean(teamID));
  const projectNameById = useMemo(
    () => new Map((projectsLookupData?.items ?? []).map((project) => [project.id, project.name])),
    [projectsLookupData]
  );
  const ownedProjects = ownedProjectsData?.items ?? [];

  const {
    data: ownedApplicationsData,
    error: ownedApplicationsErrorObj,
    isError: isOwnedApplicationsError,
    isLoading: isOwnedApplicationsLoading,
    refetch: refetchOwnedApplications,
  } = useTeamApplications(teamID);
  const ownedApplications = ownedApplicationsData?.items ?? [];

  useEffect(() => {
    setEditModalOpen(false);
    setDeleteDialogOpen(false);
    setNewMemberUserId("");
    setMemberError(null);
  }, [teamID]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      if (editModalOpen || deleteDialogOpen) return;
      onClose();
    };
    if (teamID) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [teamID, onClose, editModalOpen, deleteDialogOpen]);

  const members = membersData?.items ?? [];

  const handleAddMember = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!team) return;

    setMemberError(null);
    const userId = Number(newMemberUserId);
    if (!newMemberUserId.trim() || !Number.isInteger(userId) || userId <= 0) {
      setMemberError("Enter a valid numeric user ID.");
      return;
    }

    try {
      await addMember.mutateAsync({ teamId: team.id, payload: { userId } });
      setNewMemberUserId("");
    } catch (memberAddError) {
      setMemberError(getErrorMessage(memberAddError, "Unable to add team member. Please try again."));
    }
  };

  const handleRemoveMember = async (userId: number) => {
    if (!team) return;
    setMemberError(null);
    try {
      await removeMember.mutateAsync({ teamId: team.id, userId });
    } catch (memberRemoveError) {
      setMemberError(getErrorMessage(memberRemoveError, "Unable to remove team member. Please try again."));
    }
  };

  if (!teamID) return null;

  if (isLoading) {
    return (
      <DrawerFrame onClose={onClose} ariaLabel="Loading team details">
        <div className="flex flex-1 items-center justify-center p-5 text-sm" style={{ color: "var(--muted-foreground)" }}>
          Loading team details...
        </div>
      </DrawerFrame>
    );
  }

  if (isError || !team) {
    return (
      <DrawerFrame onClose={onClose} ariaLabel="Team details unavailable">
        <div className="p-5">
          <ErrorState
            description={error instanceof Error ? error.message : "Unable to load team details."}
            onRetry={() => {
              void refetch();
            }}
          />
        </div>
      </DrawerFrame>
    );
  }

  return (
    <div
      className="fixed inset-0 z-[60] flex justify-end bg-[color:color-mix(in_srgb,var(--background)_72%,transparent)]"
      role="presentation"
      onMouseDown={onClose}
    >
      <aside
        role="dialog"
        aria-modal="true"
        aria-labelledby="team-details-title"
        onMouseDown={(event) => event.stopPropagation()}
        className="flex h-full w-full max-w-xl flex-col border-l shadow-[var(--shadow-lg)]"
        style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
      >
        <header className="border-b p-5" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 items-center gap-3">
              <div
                className="rounded-2xl p-3"
                style={{
                  color: "var(--primary)",
                  backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)",
                }}
              >
                <Users aria-hidden="true" className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <h2 id="team-details-title" className="truncate text-xl font-semibold tracking-tight">
                  {team.name}
                </h2>
                <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                  Team workspace
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              {isAdmin ? (
                <>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => setEditModalOpen(true)}
                    className="px-3"
                    aria-label="Edit team"
                  >
                    <Pencil aria-hidden="true" className="h-4 w-4" />
                    <span className="hidden sm:inline">Edit</span>
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => setDeleteDialogOpen(true)}
                    className="px-3 text-[var(--danger)]"
                    aria-label="Delete team"
                  >
                    <Trash2 aria-hidden="true" className="h-4 w-4" />
                    <span className="hidden sm:inline">Delete</span>
                  </Button>
                </>
              ) : null}
              <Button
                type="button"
                variant="ghost"
                onClick={onClose}
                className="h-9 w-9 rounded-xl p-0"
                aria-label="Close team details"
              >
                <X aria-hidden="true" className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </header>

        <div className="flex-1 space-y-6 overflow-y-auto p-5">
          <section>
            <h3 className="text-sm font-semibold">Team information</h3>
            <dl className="mt-3 grid gap-3">
              <DrawerStat label="Description" value={team.description || "No description provided"} />
              <DrawerStat label="Created" value={formatDateTime(team.createdAt)} />
              <DrawerStat label="Updated" value={formatDateTime(team.updatedAt)} />
            </dl>
          </section>

          <section>
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold">Members{membersData ? ` (${membersData.total})` : ""}</h3>
            </div>

            {isAdmin ? (
              <form onSubmit={handleAddMember} className="mt-3 flex items-end gap-2">
                <div className="flex-1">
                  <label htmlFor="new-member-user-id" className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                    User ID
                  </label>
                  <input
                    id="new-member-user-id"
                    type="number"
                    min={1}
                    value={newMemberUserId}
                    onChange={(event) => setNewMemberUserId(event.target.value)}
                    placeholder="e.g. 42"
                    disabled={addMember.isPending}
                    className="mt-1 h-10 w-full rounded-2xl border bg-transparent px-3 text-sm outline-none transition-colors focus:border-[var(--primary)]"
                    style={{ borderColor: "var(--border)" }}
                  />
                </div>
                <Button type="submit" loading={addMember.isPending} disabled={addMember.isPending}>
                  <UserPlus aria-hidden="true" className="h-4 w-4" />
                  Add
                </Button>
              </form>
            ) : null}
            {memberError ? (
              <p className="mt-2 text-xs" style={{ color: "var(--danger)" }}>
                {memberError}
              </p>
            ) : null}

            <div className="mt-4">
              {isMembersError ? (
                <ErrorState
                  description={membersErrorObj instanceof Error ? membersErrorObj.message : "Unable to load team members."}
                  onRetry={() => {
                    void refetchMembers();
                  }}
                />
              ) : isMembersLoading ? (
                <ListSkeleton rows={3} />
              ) : members.length === 0 ? (
                <EmptyState
                  icon={Users}
                  title="No members yet"
                  description="Add a member above by their user ID to get started."
                />
              ) : (
                <ul className="space-y-2">
                  {members.map((member) => (
                    <li
                      key={member.id}
                      className="flex items-center justify-between gap-3 rounded-2xl border p-3"
                      style={{ borderColor: "var(--border)" }}
                    >
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium">{member.name}</p>
                        <p className="truncate text-xs" style={{ color: "var(--muted-foreground)" }}>
                          {member.email} · User ID {member.userId}
                        </p>
                      </div>
                      {isAdmin ? (
                        <Button
                          type="button"
                          variant="ghost"
                          onClick={() => {
                            void handleRemoveMember(member.userId);
                          }}
                          loading={removeMember.isPending && removeMember.variables?.userId === member.userId}
                          disabled={removeMember.isPending}
                          className="h-9 w-9 shrink-0 rounded-xl p-0 text-[var(--danger)]"
                          aria-label={`Remove ${member.name} from team`}
                        >
                          <Trash2 aria-hidden="true" className="h-4 w-4" />
                        </Button>
                      ) : null}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </section>

          <section>
            <h3 className="text-sm font-semibold">
              Owned projects{ownedProjectsData ? ` (${ownedProjectsData.total})` : ""}
            </h3>
            <div className="mt-4">
              {isOwnedProjectsError ? (
                <ErrorState
                  description={
                    ownedProjectsErrorObj instanceof Error
                      ? ownedProjectsErrorObj.message
                      : "Unable to load owned projects."
                  }
                  onRetry={() => {
                    void refetchOwnedProjects();
                  }}
                />
              ) : isOwnedProjectsLoading ? (
                <ListSkeleton rows={2} />
              ) : ownedProjects.length === 0 ? (
                <EmptyState
                  icon={FolderKanban}
                  title="No projects owned"
                  description="Assign this team to a project from the project's Members tab."
                />
              ) : (
                <ul className="space-y-2">
                  {ownedProjects.map((assignment) => (
                    <li
                      key={assignment.id}
                      className="rounded-2xl border p-3 text-sm font-medium"
                      style={{ borderColor: "var(--border)" }}
                    >
                      {projectNameById.get(assignment.projectId) ?? assignment.projectId}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </section>

          <section>
            <h3 className="text-sm font-semibold">
              Owned applications{ownedApplicationsData ? ` (${ownedApplicationsData.total})` : ""}
            </h3>
            <div className="mt-4">
              {isOwnedApplicationsError ? (
                <ErrorState
                  description={
                    ownedApplicationsErrorObj instanceof Error
                      ? ownedApplicationsErrorObj.message
                      : "Unable to load owned applications."
                  }
                  onRetry={() => {
                    void refetchOwnedApplications();
                  }}
                />
              ) : isOwnedApplicationsLoading ? (
                <ListSkeleton rows={2} />
              ) : ownedApplications.length === 0 ? (
                <EmptyState
                  icon={Layers3}
                  title="No applications owned"
                  description="Assign this team to an application from the application's Teams tab."
                />
              ) : (
                <ul className="space-y-2">
                  {ownedApplications.map((assignment) => (
                    <OwnedApplicationRow key={assignment.id} applicationId={assignment.applicationId} />
                  ))}
                </ul>
              )}
            </div>
          </section>
        </div>
      </aside>

      <EditTeamModal open={editModalOpen} team={team} onClose={() => setEditModalOpen(false)} />
      <DeleteTeamDialog
        open={deleteDialogOpen}
        team={team}
        onClose={() => setDeleteDialogOpen(false)}
        onSuccess={onClose}
      />
    </div>
  );
}

function OwnedApplicationRow({ applicationId }: { applicationId: string }) {
  const { data: application, isLoading } = useApplication(applicationId);

  return (
    <li className="rounded-2xl border p-3 text-sm font-medium" style={{ borderColor: "var(--border)" }}>
      {isLoading ? "Loading application..." : application?.name ?? applicationId}
    </li>
  );
}

function DrawerFrame({
  ariaLabel,
  children,
  onClose,
}: {
  ariaLabel: string;
  children: React.ReactNode;
  onClose: () => void;
}) {
  return (
    <div
      className="fixed inset-0 z-[60] flex justify-end bg-[color:color-mix(in_srgb,var(--background)_72%,transparent)]"
      role="presentation"
      onMouseDown={onClose}
    >
      <aside
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel}
        onMouseDown={(event) => event.stopPropagation()}
        className="flex h-full w-full max-w-xl flex-col border-l shadow-[var(--shadow-lg)]"
        style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
      >
        {children}
      </aside>
    </div>
  );
}

function DrawerStat({ label, value }: { label: string; value: string }) {
  return (
    <div
      className="rounded-2xl border p-4"
      style={{
        borderColor: "var(--border)",
        backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)",
      }}
    >
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        {label}
      </dt>
      <dd className="mt-1 break-all font-semibold">{value}</dd>
    </div>
  );
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString();
}

function getErrorMessage(error: unknown, fallback: string) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallback;
  }
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return fallback;
}
