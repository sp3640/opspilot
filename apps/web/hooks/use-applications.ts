"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { applicationService } from "@/services/application-service";
import type {
  ApplicationQueryParams,
  CreateApplicationRequest,
  UpdateApplicationRequest,
} from "@/types/application-api";

const applicationKeys = {
  all: ["applications"] as const,
  list: (projectId: string, params: ApplicationQueryParams) =>
    ["applications", "list", projectId, params] as const,
  detail: (id: string) => ["applications", "detail", id] as const,
};

export function useApplicationsByProject(
  projectId: string | null,
  params: ApplicationQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: applicationKeys.list(projectId ?? "", params),
    queryFn: () => applicationService.listApplicationsByProject(projectId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(projectId),
  });
}

export function useApplication(applicationId: string | null) {
  return useQuery({
    queryKey: applicationKeys.detail(applicationId ?? ""),
    queryFn: () => applicationService.getApplication(applicationId ?? ""),
    enabled: Boolean(applicationId),
  });
}

export function useCreateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ projectId, payload }: { projectId: string; payload: CreateApplicationRequest }) =>
      applicationService.createApplication(projectId, payload),
    onSuccess: () => {
      toast.success("Application created successfully.");
      return queryClient.invalidateQueries({ queryKey: applicationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to create application."));
    },
  });
}

export function useUpdateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateApplicationRequest }) =>
      applicationService.updateApplication(id, payload),
    onSuccess: (application) => {
      toast.success("Application updated successfully.");
      queryClient.setQueryData(applicationKeys.detail(application.id), application);
      return queryClient.invalidateQueries({ queryKey: applicationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update application."));
    },
  });
}

export function useDeleteApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => applicationService.deleteApplication(id),
    onSuccess: () => {
      toast.success("Application deleted successfully.");
      return queryClient.invalidateQueries({ queryKey: applicationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to delete application."));
    },
  });
}

function getMutationErrorMessage(error: unknown, fallback: string) {
  if (axios.isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallback;
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return fallback;
}
