"use client";

import { useEffect, useState } from "react";
import { Activity, ClipboardList, Layers3, Pencil, Rocket, Settings, Trash2, Users, type LucideIcon, X } from "lucide-react";

import { ApplicationWorkspace } from "@/components/applications/application-workspace";
import { ErrorState } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useProject } from "@/hooks/use-projects";
import { useIsPlatformAdmin } from "@/store/auth-store";

import { DeleteProjectDialog } from "./delete-project-dialog";
import { EditProjectModal } from "./edit-project-modal";
import { getProjectIcon } from "./project-icon";
import { ProjectAuditLog } from "./project-audit-log";
import { ProjectDeployments } from "./project-deployments";
import { ProjectEnvironmentBadge, ProjectHealthBadge } from "./project-status";
import { ProjectOverview } from "./project-overview";
import { ProjectSettings } from "./project-settings";
import { ProjectTeams } from "./project-teams";

const tabs: ReadonlyArray<{ label: string; icon: LucideIcon }> = [
  { label: "Overview", icon: Activity },
  { label: "Applications", icon: Layers3 },
  { label: "Deployments", icon: Rocket },
  { label: "Members", icon: Users },
  { label: "Audit", icon: ClipboardList },
  { label: "Settings", icon: Settings },
];

/** Right-side project context loaded from the Project detail API. */
export function ProjectDetailsDrawer({ projectID, onClose }: { projectID: string | null; onClose: () => void }) {
  const [activeTab, setActiveTab] = useState("Overview");
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const { data: project, error, isError, isLoading, refetch } = useProject(projectID);
  const isAdmin = useIsPlatformAdmin();

  useEffect(() => {
    setActiveTab("Overview");
    setEditModalOpen(false);
    setDeleteDialogOpen(false);
  }, [projectID]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      if (editModalOpen || deleteDialogOpen) return;
      onClose();
    };
    if (projectID) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [projectID, onClose, editModalOpen, deleteDialogOpen]);

  if (!projectID) return null;

  if (isLoading) {
    return (
      <DrawerFrame onClose={onClose} ariaLabel="Loading project details">
        <div className="flex flex-1 items-center justify-center p-5 text-sm" style={{ color: "var(--muted-foreground)" }}>
          Loading project details...
        </div>
      </DrawerFrame>
    );
  }

  if (isError || !project) {
    return (
      <DrawerFrame onClose={onClose} ariaLabel="Project details unavailable">
        <div className="p-5">
          <ErrorState
            description={error instanceof Error ? error.message : "Unable to load project details."}
            onRetry={() => {
              void refetch();
            }}
          />
        </div>
      </DrawerFrame>
    );
  }

  const Icon = getProjectIcon("folder-kanban");
  const titleId = "project-details-title";
  const tabListId = "project-details-tablist";
  const activeTabKey = activeTab.toLowerCase();
  const activePanelId = `project-details-panel-${activeTabKey}`;

  return (
    <div
      className="fixed inset-0 z-[60] flex justify-end bg-[color:color-mix(in_srgb,var(--background)_72%,transparent)]"
      role="presentation"
      onMouseDown={onClose}
    >
      <aside
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
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
                <Icon aria-hidden="true" className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <h2 id={titleId} className="truncate text-xl font-semibold tracking-tight">{project.name}</h2>
                <p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>
                  Project workspace
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
                    aria-label="Edit project"
                  >
                    <Pencil aria-hidden="true" className="h-4 w-4" />
                    <span className="hidden sm:inline">Edit</span>
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => setDeleteDialogOpen(true)}
                    className="px-3 text-[var(--danger)]"
                    aria-label="Delete project"
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
                aria-label="Close project details"
              >
                <X aria-hidden="true" className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <div className="mt-5 flex gap-2">
            <ProjectHealthBadge health={project.health} />
            <ProjectEnvironmentBadge environment={project.environment} />
          </div>
        </header>

        <div className="overflow-x-auto border-b px-3" style={{ borderColor: "var(--border)" }}>
          <div className="flex gap-1" role="tablist" id={tabListId} aria-label="Project details sections">
            {tabs.map(({ label, icon: TabIcon }, index) => {
              const isActive = activeTab === label;
              const tabId = `project-details-tab-${label.toLowerCase()}`;
              const panelId = `project-details-panel-${label.toLowerCase()}`;

              return (
                <button
                  key={label}
                  id={tabId}
                  type="button"
                  role="tab"
                  tabIndex={isActive ? 0 : -1}
                  aria-selected={isActive}
                  aria-controls={panelId}
                  onClick={() => setActiveTab(label)}
                  onKeyDown={(event) => {
                    if (!"ArrowLeft ArrowRight Home End".includes(event.key)) return;
                    event.preventDefault();

                    const maxIndex = tabs.length - 1;
                    let nextIndex = index;
                    if (event.key === "ArrowRight") nextIndex = index === maxIndex ? 0 : index + 1;
                    if (event.key === "ArrowLeft") nextIndex = index === 0 ? maxIndex : index - 1;
                    if (event.key === "Home") nextIndex = 0;
                    if (event.key === "End") nextIndex = maxIndex;

                    const nextTab = tabs[nextIndex];
                    if (!nextTab) return;
                    setActiveTab(nextTab.label);
                    const nextTabButton = document.getElementById(`project-details-tab-${nextTab.label.toLowerCase()}`);
                    nextTabButton?.focus();
                  }}
                  className="inline-flex shrink-0 items-center gap-1.5 border-b-2 px-3 py-3 text-sm font-medium transition-colors"
                  style={{
                    color: isActive ? "var(--primary)" : "var(--muted-foreground)",
                    borderColor: isActive ? "var(--primary)" : "transparent",
                  }}
                >
                  <TabIcon aria-hidden="true" className="h-3.5 w-3.5" />
                  {label}
                </button>
              );
            })}
          </div>
        </div>

        <div
          id={activePanelId}
          role="tabpanel"
          aria-labelledby={`project-details-tab-${activeTabKey}`}
          aria-describedby={tabListId}
          className="flex-1 overflow-y-auto p-5"
        >
          {activeTab === "Overview" ? (
            <ProjectOverview project={project} />
          ) : activeTab === "Applications" ? (
            <ApplicationWorkspace projectId={project.id} />
          ) : activeTab === "Deployments" ? (
            <ProjectDeployments projectId={project.id} />
          ) : activeTab === "Members" ? (
            <ProjectTeams projectId={project.id} />
          ) : activeTab === "Audit" ? (
            <ProjectAuditLog projectId={project.id} />
          ) : activeTab === "Settings" ? (
            <ProjectSettings
              project={project}
              isAdmin={isAdmin}
              onEdit={() => setEditModalOpen(true)}
              onDelete={() => setDeleteDialogOpen(true)}
            />
          ) : (
            <div
              className="flex min-h-56 items-center justify-center rounded-2xl border p-6 text-center text-sm"
              style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}
            >
              {activeTab} data is not available for {project.name}.
            </div>
          )}
        </div>
      </aside>

      <EditProjectModal open={editModalOpen} project={project} onClose={() => setEditModalOpen(false)} />
      <DeleteProjectDialog
        open={deleteDialogOpen}
        project={project}
        onClose={() => setDeleteDialogOpen(false)}
        onSuccess={onClose}
      />
    </div>
  );
}

function DrawerFrame({ ariaLabel, children, onClose }: { ariaLabel: string; children: React.ReactNode; onClose: () => void }) {
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

