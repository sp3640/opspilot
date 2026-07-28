"use client";

import { useEffect, useRef } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { FolderPlus, X } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";

import type { CreateProjectInput, ProjectEnvironment } from "./types";

const projectSchema = z.object({
  name: z.string().trim().min(3, "Project name must contain at least 3 characters.").max(80, "Project name must be 80 characters or fewer."),
  description: z.string().trim().min(12, "Add a short description of at least 12 characters.").max(300, "Description must be 300 characters or fewer."),
  environment: z.enum(["production", "staging", "development"]),
});

type CreateProjectModalProps = { open: boolean; onClose: () => void; onCreate: (project: CreateProjectInput) => void };

/** Validated project-creation workflow that can later submit to the projects API. */
export function CreateProjectModal({ open, onClose, onCreate }: CreateProjectModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm<CreateProjectInput>({ resolver: zodResolver(projectSchema), defaultValues: { name: "", description: "", environment: "development" } });
  useEffect(() => { if (open && !dialogRef.current?.open) dialogRef.current?.showModal(); if (!open && dialogRef.current?.open) dialogRef.current.close(); }, [open]);
  const close = () => { reset(); onClose(); };
  const submit = (input: CreateProjectInput) => { onCreate(input); close(); };
  const inputClass = "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

  return <dialog ref={dialogRef} onClose={onClose} aria-labelledby="create-project-title" className="m-auto w-[calc(100%-2rem)] max-w-lg rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]" style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}><form onSubmit={handleSubmit(submit)}><header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}><div className="flex items-center gap-3"><div className="rounded-2xl p-3" style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}><FolderPlus aria-hidden="true" className="h-5 w-5" /></div><div><h2 id="create-project-title" className="text-lg font-semibold">Create project</h2><p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>Start with an environment and a clear service boundary.</p></div></div><Button type="button" variant="ghost" onClick={close} className="h-9 w-9 rounded-xl p-0" aria-label="Close create project dialog"><X aria-hidden="true" className="h-4 w-4" /></Button></header><div className="space-y-5 p-5"><div><label htmlFor="project-name" className="text-sm font-medium">Project name</label><input id="project-name" {...register("name")} placeholder="e.g. Customer API" className={inputClass} style={{ borderColor: errors.name ? "var(--danger)" : "var(--border)" }} />{errors.name && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.name.message}</p>}</div><div><label htmlFor="project-description" className="text-sm font-medium">Description</label><textarea id="project-description" {...register("description")} placeholder="Describe the services and responsibility of this project." rows={4} className={inputClass} style={{ borderColor: errors.description ? "var(--danger)" : "var(--border)" }} />{errors.description && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.description.message}</p>}</div><div><label htmlFor="project-environment" className="text-sm font-medium">Primary environment</label><select id="project-environment" {...register("environment")} className={inputClass} style={{ borderColor: "var(--border)" }}>{(["production", "staging", "development"] as ProjectEnvironment[]).map((environment) => <option key={environment} value={environment}>{environment.charAt(0).toUpperCase() + environment.slice(1)}</option>)}</select></div></div><footer className="flex justify-end gap-3 border-t p-5" style={{ borderColor: "var(--border)" }}><Button type="button" variant="secondary" onClick={close}>Cancel</Button><Button type="submit" loading={isSubmitting}>Create project</Button></footer></form></dialog>;
}
