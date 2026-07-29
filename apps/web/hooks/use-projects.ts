"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { projectService } from "@/services/project-service";
import type {
  CreateProjectRequest,
  ProjectQueryParams,
  UpdateProjectRequest,
} from "@/types/project-api";

const projectKeys = {
  all: ["projects"] as const,
  list: (params: ProjectQueryParams) => ["projects", "list", params] as const,
  detail: (id: string) => ["projects", "detail", id] as const,
};

export function useProjects(params: ProjectQueryParams) {
  return useQuery({
    queryKey: projectKeys.list(params),
    queryFn: () => projectService.listProjects(params),
  });
}

export function useProject(id: string | null) {
  return useQuery({
    queryKey: projectKeys.detail(id ?? ""),
    queryFn: () => projectService.getProject(id ?? ""),
    enabled: Boolean(id),
  });
}

export function useCreateProject() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateProjectRequest) => projectService.createProject(payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: projectKeys.all }),
  });
}

export function useUpdateProject() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateProjectRequest }) => projectService.updateProject(id, payload),
    onSuccess: (project) => {
      queryClient.setQueryData(projectKeys.detail(project.id), project);
      return queryClient.invalidateQueries({ queryKey: projectKeys.all });
    },
  });
}

export function useDeleteProject() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => projectService.deleteProject(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: projectKeys.all }),
  });
}
