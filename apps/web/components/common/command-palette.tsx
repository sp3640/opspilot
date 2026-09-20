"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { AlertTriangle, FileDown, FolderKanban, FolderPlus, Search, ShieldAlert, Users, X } from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { useIncidents } from "@/hooks/use-incidents";
import { useProjects } from "@/hooks/use-projects";
import { useAlerts } from "@/hooks/use-alerts";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useCommandPaletteStore } from "@/store/command-palette-store";

type PaletteItem = {
  id: string;
  label: string;
  description: string;
  group: string;
  icon: LucideIcon;
  onSelect: () => void;
};

const staticCommands: Omit<PaletteItem, "onSelect">[] = [
  { id: "projects", label: "Projects", description: "Browse all projects", group: "Navigate", icon: FolderKanban },
  { id: "audit", label: "Audit logs", description: "Review platform activity", group: "Navigate", icon: FileDown },
  { id: "incidents", label: "Incidents", description: "Monitor active incidents", group: "Navigate", icon: ShieldAlert },
  { id: "users", label: "Users", description: "Manage team access", group: "Navigate", icon: Users },
  { id: "create-project", label: "Create project", description: "Open the projects workspace", group: "Quick actions", icon: FolderPlus },
  { id: "create-incident", label: "Create incident", description: "Open the incident workspace", group: "Quick actions", icon: ShieldAlert },
  { id: "export-audit", label: "Export audit logs", description: "Open audit exports", group: "Quick actions", icon: FileDown },
  { id: "invite-member", label: "Invite member", description: "Open workspace settings", group: "Quick actions", icon: Users },
];

const staticHrefs: Record<string, string> = {
  projects: "/projects",
  audit: "/audit",
  incidents: "/incidents",
  users: "/settings",
  "create-project": "/projects",
  "create-incident": "/incidents",
  "export-audit": "/audit",
  "invite-member": "/settings",
};

const MIN_SEARCH_LENGTH = 2;
const SEARCH_DEBOUNCE_MS = 300;

/**
 * Global keyboard-first navigation and action dialog for dashboard pages.
 * Below MIN_SEARCH_LENGTH characters it shows the static navigate/quick-
 * action shortcuts (unchanged from before); at MIN_SEARCH_LENGTH+ it
 * switches to a real, debounced search across incidents, projects, and
 * alerts using the same list endpoints and `search` query param their own
 * workspace pages already use - no new backend endpoint. Incident results
 * deep-link straight to /incidents/[id]; project/alert results open their
 * list page (neither has its own deep-linkable detail route today - see
 * Sprint report).
 */
export function CommandPalette() {
  const router = useRouter();
  const dialogRef = useRef<HTMLDialogElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const isOpen = useCommandPaletteStore((state) => state.isOpen);
  const openPalette = useCommandPaletteStore((state) => state.open);
  const closePalette = useCommandPaletteStore((state) => state.close);
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);

  const trimmedQuery = query.trim();
  const debouncedQuery = useDebouncedValue(trimmedQuery, SEARCH_DEBOUNCE_MS);
  const isSearching = debouncedQuery.length >= MIN_SEARCH_LENGTH;

  const incidentsQuery = useIncidents(
    { search: debouncedQuery, limit: 5, sort: "updated_at", order: "desc" },
    isOpen && isSearching
  );
  const projectsQuery = useProjects(
    { search: debouncedQuery, limit: 5, sort: "updated_at", order: "desc" },
    isOpen && isSearching
  );
  const alertsQuery = useAlerts(
    { search: debouncedQuery, limit: 5, sort: "updated_at", order: "desc" },
    isOpen && isSearching
  );

  const close = () => {
    dialogRef.current?.close();
    closePalette();
    setQuery("");
  };
  const select = (item: PaletteItem) => {
    item.onSelect();
    close();
  };

  useEffect(() => {
    const handleShortcut = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        openPalette();
      }
    };
    window.addEventListener("keydown", handleShortcut);
    return () => window.removeEventListener("keydown", handleShortcut);
  }, [openPalette]);

  useEffect(() => {
    if (isOpen && !dialogRef.current?.open) {
      dialogRef.current?.showModal();
      window.setTimeout(() => inputRef.current?.focus(), 0);
    } else if (!isOpen && dialogRef.current?.open) {
      dialogRef.current.close();
    }
  }, [isOpen]);

  useEffect(() => {
    setActiveIndex(0);
  }, [query, incidentsQuery.data, projectsQuery.data, alertsQuery.data]);

  const staticItems: PaletteItem[] = useMemo(
    () =>
      staticCommands.map((command) => ({
        ...command,
        onSelect: () => router.push(staticHrefs[command.id] ?? "/"),
      })),
    [router]
  );

  const filteredStaticItems = useMemo(() => {
    if (!trimmedQuery) return staticItems;
    const normalized = trimmedQuery.toLowerCase();
    return staticItems.filter((item) => `${item.label} ${item.description}`.toLowerCase().includes(normalized));
  }, [staticItems, trimmedQuery]);

  const searchResultItems: PaletteItem[] = useMemo(() => {
    if (!isSearching) return [];

    const incidents = (incidentsQuery.data?.items ?? []).map(
      (incident): PaletteItem => ({
        id: `incident-${incident.id}`,
        label: incident.title,
        description: `${incident.severity} · ${incident.status}`,
        group: "Incidents",
        icon: ShieldAlert,
        onSelect: () => router.push(`/incidents/${incident.id}`),
      })
    );

    const projects = (projectsQuery.data?.items ?? []).map(
      (project): PaletteItem => ({
        id: `project-${project.id}`,
        label: project.name,
        description: "Open in Projects",
        group: "Projects",
        icon: FolderKanban,
        onSelect: () => router.push("/projects"),
      })
    );

    const alerts = (alertsQuery.data?.items ?? []).map(
      (alert): PaletteItem => ({
        id: `alert-${alert.id}`,
        label: alert.title,
        description: `${alert.severity} · Open in Alerts`,
        group: "Alerts",
        icon: AlertTriangle,
        onSelect: () => router.push("/alerts"),
      })
    );

    return [...incidents, ...projects, ...alerts];
  }, [isSearching, incidentsQuery.data, projectsQuery.data, alertsQuery.data, router]);

  const isSearchLoading = isSearching && (incidentsQuery.isLoading || projectsQuery.isLoading || alertsQuery.isLoading);
  const isSearchError = isSearching && (incidentsQuery.isError || projectsQuery.isError || alertsQuery.isError);

  const visibleItems = isSearching ? searchResultItems : filteredStaticItems;
  const groups = isSearching ? (["Incidents", "Projects", "Alerts"] as const) : (["Navigate", "Quick actions"] as const);

  return (
    <dialog
      ref={dialogRef}
      onClose={closePalette}
      aria-labelledby="command-palette-title"
      className="m-auto w-[calc(100%-2rem)] max-w-2xl overflow-hidden rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]"
      style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}
    >
      <div className="border-b p-4" style={{ borderColor: "var(--border)" }}>
        <div className="flex items-center gap-3">
          <Search aria-hidden="true" className="h-5 w-5 shrink-0" style={{ color: "var(--muted-foreground)" }} />
          <label htmlFor="command-search" className="sr-only">
            Search command palette
          </label>
          <input
            ref={inputRef}
            id="command-search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "ArrowDown") {
                event.preventDefault();
                setActiveIndex((index) => Math.min(index + 1, visibleItems.length - 1));
              }
              if (event.key === "ArrowUp") {
                event.preventDefault();
                setActiveIndex((index) => Math.max(index - 1, 0));
              }
              if (event.key === "Enter" && visibleItems[activeIndex]) {
                event.preventDefault();
                select(visibleItems[activeIndex]);
              }
            }}
            placeholder="Search projects, incidents, alerts, and actions..."
            className="min-w-0 flex-1 bg-transparent text-base outline-none placeholder:text-[var(--muted-foreground)]"
            aria-activedescendant={visibleItems[activeIndex] ? `command-${visibleItems[activeIndex].id}` : undefined}
          />
          <button
            type="button"
            onClick={close}
            className="rounded-lg p-1.5 transition-colors hover:bg-[var(--muted)]"
            aria-label="Close command palette"
          >
            <X aria-hidden="true" className="h-4 w-4" />
          </button>
        </div>
      </div>
      <div className="max-h-[min(60vh,440px)] overflow-y-auto p-2" role="listbox" aria-label="Available commands and search results">
        <h2 id="command-palette-title" className="sr-only">
          Command palette
        </h2>

        {isSearching && isSearchLoading && (
          <p className="px-3 py-12 text-center text-sm" style={{ color: "var(--muted-foreground)" }}>
            Searching…
          </p>
        )}

        {isSearching && !isSearchLoading && isSearchError && (
          <p className="px-3 py-12 text-center text-sm" style={{ color: "var(--danger-text)" }}>
            Something went wrong running this search. Try again.
          </p>
        )}

        {!isSearchLoading &&
          !isSearchError &&
          groups.map((group) => {
            const groupItems = visibleItems.filter((item) => item.group === group);
            if (groupItems.length === 0) return null;
            return (
              <div key={group} className="py-2">
                <p className="px-3 pb-2 text-xs font-semibold uppercase tracking-[0.12em]" style={{ color: "var(--muted-foreground)" }}>
                  {group}
                </p>
                {groupItems.map((item) => {
                  const index = visibleItems.indexOf(item);
                  const Icon = item.icon;
                  const active = index === activeIndex;
                  return (
                    <button
                      id={`command-${item.id}`}
                      key={item.id}
                      type="button"
                      role="option"
                      aria-selected={active}
                      onMouseEnter={() => setActiveIndex(index)}
                      onClick={() => select(item)}
                      className="flex w-full items-center gap-3 rounded-2xl px-3 py-3 text-left transition-colors"
                      style={{ backgroundColor: active ? "var(--muted)" : "transparent" }}
                    >
                      <span
                        className="rounded-xl p-2"
                        style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}
                      >
                        <Icon aria-hidden="true" className="h-4 w-4" />
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm font-medium">{item.label}</span>
                        <span className="mt-0.5 block truncate text-xs" style={{ color: "var(--muted-foreground)" }}>
                          {item.description}
                        </span>
                      </span>
                    </button>
                  );
                })}
              </div>
            );
          })}

        {!isSearchLoading && !isSearchError && visibleItems.length === 0 && (
          <div className="px-3 py-12 text-center text-sm" style={{ color: "var(--muted-foreground)" }}>
            {isSearching ? `No incidents, projects, or alerts match "${debouncedQuery}".` : "No matching commands found."}
          </div>
        )}
      </div>
      <footer
        className="flex items-center justify-between border-t px-4 py-3 text-xs"
        style={{ color: "var(--muted-foreground)", borderColor: "var(--border)" }}
      >
        <span>↑↓ to navigate · Enter to select</span>
        <span>Esc to close</span>
      </footer>
    </dialog>
  );
}
