"use client";

import { useEffect, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import axios from "axios";
import { FolderPlus, X } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { useCreateProject } from "@/hooks/use-projects";
import type { CreateProjectRequest, ProjectResponse } from "@/types/project-api";

const projectSchema = z.object({
	name: z.string().trim().min(3, "Project name must contain at least 3 characters.").max(100, "Project name must be 100 characters or fewer."),
	description: z.string().trim().max(300, "Description must be 300 characters or fewer."),
});

type CreateProjectModalProps = { open: boolean; onClose: () => void; onCreated: (project: ProjectResponse) => void };

/** Validated project-creation workflow backed by the projects API. */
export function CreateProjectModal({ open, onClose, onCreated }: CreateProjectModalProps) {
	const dialogRef = useRef<HTMLDialogElement>(null);
	const createProject = useCreateProject();
	const [submitError, setSubmitError] = useState<string | null>(null);
	const { register, handleSubmit, reset, formState: { errors } } = useForm<CreateProjectRequest>({ resolver: zodResolver(projectSchema), defaultValues: { name: "", description: "" } });

	useEffect(() => { if (open && !dialogRef.current?.open) dialogRef.current?.showModal(); if (!open && dialogRef.current?.open) dialogRef.current.close(); }, [open]);

	const close = () => { if (createProject.isPending) return; reset(); setSubmitError(null); onClose(); };
	const submit = async (input: CreateProjectRequest) => {
		setSubmitError(null);
		try {
			const project = await createProject.mutateAsync(input);
			reset();
			onCreated(project);
			onClose();
		} catch (error) {
			setSubmitError(getProjectErrorMessage(error));
		}
	};
	const inputClass = "mt-2 w-full rounded-2xl border bg-transparent px-3 py-2.5 text-sm outline-none transition-colors placeholder:text-[var(--muted-foreground)] focus:border-[var(--primary)]";

	return <dialog ref={dialogRef} onClose={onClose} aria-labelledby="create-project-title" className="m-auto w-[calc(100%-2rem)] max-w-lg rounded-3xl border p-0 text-[var(--foreground)] shadow-[var(--shadow-lg)] backdrop:bg-[var(--background)]" style={{ backgroundColor: "var(--card)", borderColor: "var(--border)" }}><form onSubmit={handleSubmit(submit)}><header className="flex items-start justify-between border-b p-5" style={{ borderColor: "var(--border)" }}><div className="flex items-center gap-3"><div className="rounded-2xl p-3" style={{ color: "var(--primary)", backgroundColor: "color-mix(in srgb, var(--primary) 12%, transparent)" }}><FolderPlus aria-hidden="true" className="h-5 w-5" /></div><div><h2 id="create-project-title" className="text-lg font-semibold">Create project</h2><p className="mt-1 text-sm" style={{ color: "var(--muted-foreground)" }}>Start with a clear service boundary.</p></div></div><Button type="button" variant="ghost" onClick={close} disabled={createProject.isPending} className="h-9 w-9 rounded-xl p-0" aria-label="Close create project dialog"><X aria-hidden="true" className="h-4 w-4" /></Button></header><div className="space-y-5 p-5"><div><label htmlFor="project-name" className="text-sm font-medium">Project name</label><input id="project-name" {...register("name")} placeholder="e.g. Customer API" className={inputClass} style={{ borderColor: errors.name ? "var(--danger)" : "var(--border)" }} />{errors.name && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.name.message}</p>}</div><div><label htmlFor="project-description" className="text-sm font-medium">Description</label><textarea id="project-description" {...register("description")} placeholder="Describe the services and responsibility of this project." rows={4} className={inputClass} style={{ borderColor: errors.description ? "var(--danger)" : "var(--border)" }} />{errors.description && <p className="mt-1.5 text-xs" style={{ color: "var(--danger)" }}>{errors.description.message}</p>}</div>{submitError && <p role="alert" className="text-sm" style={{ color: "var(--danger)" }}>{submitError}</p>}</div><footer className="flex justify-end gap-3 border-t p-5" style={{ borderColor: "var(--border)" }}><Button type="button" variant="secondary" onClick={close} disabled={createProject.isPending}>Cancel</Button><Button type="submit" loading={createProject.isPending}>Create project</Button></footer></form></dialog>;
}

function getProjectErrorMessage(error: unknown) {
	if (axios.isAxiosError<{ message?: string }>(error)) return error.response?.data?.message ?? "Unable to create project. Please try again.";
	return "Unable to create project. Please try again.";
}
