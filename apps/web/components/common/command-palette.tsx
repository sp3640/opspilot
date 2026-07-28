"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { FileDown, FolderKanban, FolderPlus, Search, ShieldAlert, Users, X } from "lucide-react";
import type { LucideIcon } from "lucide-react";

type Command = { id: string; label: string; description: string; href: string; group: "Navigate" | "Quick actions"; icon: LucideIcon };

const commands: Command[] = [
  { id: "projects", label: "Projects", description: "Browse all projects", href: "/projects", group: "Navigate", icon: FolderKanban },
  { id: "audit", label: "Audit logs", description: "Review platform activity", href: "/audit", group: "Navigate", icon: FileDown },
  { id: "incidents", label: "Incidents", description: "Monitor active incidents", href: "/incidents", group: "Navigate", icon: ShieldAlert },
  { id: "users", label: "Users", description: "Manage team access", href: "/settings", group: "Navigate", icon: Users },
  { id: "create-project", label: "Create project", description: "Open the projects workspace", href: "/projects", group: "Quick actions", icon: FolderPlus },
  { id: "create-incident", label: "Create incident", description: "Open the incident workspace", href: "/incidents", group: "Quick actions", icon: ShieldAlert },
  { id: "export-audit", label: "Export audit logs", description: "Open audit exports", href: "/audit", group: "Quick actions", icon: FileDown },
  { id: "invite-member", label: "Invite member", description: "Open workspace settings", href: "/settings", group: "Quick actions", icon: Users },
];

/** Global keyboard-first navigation and action dialog for dashboard pages. */
export function CommandPalette() {
  const router = useRouter();
  const dialogRef = useRef<HTMLDialogElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const filteredCommands = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    return normalized ? commands.filter((command) => `${command.label} ${command.description}`.toLowerCase().includes(normalized)) : commands;
  }, [query]);

  const close = () => { dialogRef.current?.close(); setOpen(false); setQuery(""); };
  const select = (command: Command) => { router.push(command.href); close(); };

  useEffect(() => {
    const handleShortcut = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        if (!dialogRef.current?.open) dialogRef.current?.showModal();
        setOpen(true);
      }
    };
    window.addEventListener("keydown", handleShortcut);
    return () => window.removeEventListener("keydown", handleShortcut);
  }, []);

  useEffect(() => { setActiveIndex(0); }, [query]);
  useEffect(() => { if (open) window.setTimeout(() => inputRef.current?.focus(), 0); }, [open]);

  return (
    <dialog ref={dialogRef} onClose={() => setOpen(false)} aria-labelledby="command-palette-title" className="m-auto w-[calc(100%-2rem)] max-w-2xl overflow-hidden rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]" style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}>
      <div className="border-b p-4" style={{ borderColor: "var(--border)" }}>
        <div className="flex items-center gap-3"><Search aria-hidden="true" className="h-5 w-5 shrink-0" style={{ color: "var(--muted-foreground)" }} /><label htmlFor="command-search" className="sr-only">Search commands</label><input ref={inputRef} id="command-search" value={query} onChange={(event) => setQuery(event.target.value)} onKeyDown={(event) => { if (event.key === "ArrowDown") { event.preventDefault(); setActiveIndex((index) => Math.min(index + 1, filteredCommands.length - 1)); } if (event.key === "ArrowUp") { event.preventDefault(); setActiveIndex((index) => Math.max(index - 1, 0)); } if (event.key === "Enter" && filteredCommands[activeIndex]) { event.preventDefault(); select(filteredCommands[activeIndex]); } }} placeholder="Search projects, incidents, and actions..." className="min-w-0 flex-1 bg-transparent text-base outline-none placeholder:text-[var(--muted-foreground)]" aria-activedescendant={filteredCommands[activeIndex] ? `command-${filteredCommands[activeIndex].id}` : undefined} /><button type="button" onClick={close} className="rounded-lg p-1.5 transition-colors hover:bg-[var(--muted)]" aria-label="Close command palette"><X aria-hidden="true" className="h-4 w-4" /></button></div>
      </div>
      <div className="max-h-[min(60vh,440px)] overflow-y-auto p-2" role="listbox" aria-label="Available commands">
        <h2 id="command-palette-title" className="sr-only">Command palette</h2>
        {(["Navigate", "Quick actions"] as const).map((group) => { const groupCommands = filteredCommands.filter((command) => command.group === group); return groupCommands.length > 0 ? <div key={group} className="py-2"><p className="px-3 pb-2 text-xs font-semibold uppercase tracking-[0.12em]" style={{ color: "var(--muted-foreground)" }}>{group}</p>{groupCommands.map((command) => { const index = filteredCommands.indexOf(command); const Icon = command.icon; const active = index === activeIndex; return <button id={`command-${command.id}`} key={command.id} type="button" role="option" aria-selected={active} onMouseEnter={() => setActiveIndex(index)} onClick={() => select(command)} className="flex w-full items-center gap-3 rounded-2xl px-3 py-3 text-left transition-colors" style={{ backgroundColor: active ? "var(--muted)" : "transparent" }}><span className="rounded-xl p-2" style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}><Icon aria-hidden="true" className="h-4 w-4" /></span><span className="min-w-0 flex-1"><span className="block text-sm font-medium">{command.label}</span><span className="mt-0.5 block truncate text-xs" style={{ color: "var(--muted-foreground)" }}>{command.description}</span></span></button>; })}</div> : null; })}
        {filteredCommands.length === 0 && <div className="px-3 py-12 text-center text-sm" style={{ color: "var(--muted-foreground)" }}>No matching commands found.</div>}
      </div>
      <footer className="flex items-center justify-between border-t px-4 py-3 text-xs" style={{ color: "var(--muted-foreground)", borderColor: "var(--border)" }}><span>↑↓ to navigate · Enter to select</span><span>Esc to close</span></footer>
    </dialog>
  );
}
