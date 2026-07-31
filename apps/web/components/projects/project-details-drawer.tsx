"use client";

import { useEffect, useState } from "react";
import { Activity, ClipboardList, Rocket, Settings, Users, type LucideIcon, X } from "lucide-react";

import { ErrorState } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useProject } from "@/hooks/use-projects";

import { getProjectIcon } from "./project-icon";
import { ProjectEnvironmentBadge, ProjectHealthBadge } from "./project-status";

const tabs: ReadonlyArray<{ label: string; icon: LucideIcon }> = [
  { label: "Overview", icon: Activity },
  { label: "Deployments", icon: Rocket },
  { label: "Members", icon: Users },
  { label: "Audit", icon: ClipboardList },
  { label: "Settings", icon: Settings },
];

/** Right-side project context loaded from the Project detail API. */
export function ProjectDetailsDrawer({ projectID, onClose }: { projectID: string | null; onClose: () => void }) {
  const [activeTab, setActiveTab] = useState("Overview");
  const { data: project, error, isError, isLoading, refetch } = useProject(projectID);

  useEffect(() => {
    setActiveTab("Overview");
  }, [projectID]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    if (projectID) window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [projectID, onClose]);

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
            <div className="space-y-6">
              <p className="text-sm leading-6" style={{ color: "var(--muted-foreground)" }}>
                {project.description}
              </p>
              <dl className="grid grid-cols-2 gap-3">
                <DrawerStat label="Services" value={String(project.services)} />
                <DrawerStat label="Members" value={String(project.members)} />
                <DrawerStat label="Created" value={formatDate(project.createdAt)} />
                <DrawerStat label="Project owner" value={project.owner.name} />
              </dl>
              <div>
                <h3 className="text-sm font-semibold">Project identifier</h3>
                <p
                  className="mt-3 rounded-2xl border p-3 text-sm"
                  style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}
                >
                  {project.slug}
                </p>
              </div>
            </div>
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

function DrawerStat({ label, value }: { label: string; value: string }) {
  return (
    <div
      className="rounded-2xl border p-4"
      style={{
        borderColor: "var(--border)",
        backgroundColor: "color-mix(in srgb, var(--muted) 45%, transparent)",
      }}
    >
      <dt className="text-xs" style={{ color: "var(--muted-foreground)" }}>{label}</dt>
      <dd className="mt-1 font-semibold">{value}</dd>
    </div>
  );
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString();
}
