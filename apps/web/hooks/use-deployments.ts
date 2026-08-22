"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { projectDeploymentKeys } from "@/hooks/use-project-deployments";
import { deploymentService } from "@/services/deployment-service";
import type { CreateDeploymentRequest, DeploymentQueryParams, UpdateDeploymentRequest } from "@/types/deployment-api";

const deploymentKeys = {
  all: ["deployments"] as const,
  detail: (id: string) => ["deployments", "detail", id] as const,
  list: (applicationId: string, params: DeploymentQueryParams) =>
    ["deployments", "list", applicationId, params] as const,
  latest: (applicationId: string) => ["deployments", "latest", applicationId] as const,
  history: (deploymentId: string) => ["deployments", "history", deploymentId] as const,
  historyRevision: (deploymentId: string, revision: number) =>
    ["deployments", "history", deploymentId, revision] as const,
};

export function useDeployment(id: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: deploymentKeys.detail(id ?? ""),
    queryFn: () => deploymentService.getDeployment(id ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(id),
  });
}

export function useDeploymentsByApplication(
  applicationId: string | null,
  params: DeploymentQueryParams,
  queryEnabled = true
) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: deploymentKeys.list(applicationId ?? "", params),
    queryFn: () => deploymentService.listDeploymentsByApplication(applicationId ?? "", params),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

/**
 * The application's most recently created deployment — the authoritative
 * "where is this running" answer (target cluster, namespace, environment).
 * A 404 means the application has no runtime mapping yet (never deployed),
 * which callers should render as an empty state rather than an error.
 */
export function useLatestDeployment(applicationId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: deploymentKeys.latest(applicationId ?? ""),
    queryFn: () => deploymentService.getLatestDeploymentByApplication(applicationId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
    retry: (failureCount, error) => {
      if (axios.isAxiosError(error) && error.response?.status === 404) return false;
      return failureCount < 2;
    },
  });
}

export function useDeploymentHistory(deploymentId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: deploymentKeys.history(deploymentId ?? ""),
    queryFn: () => deploymentService.getDeploymentHistory(deploymentId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(deploymentId),
  });
}

export function useDeploymentHistoryRevision(deploymentId: string | null, revision: number | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: deploymentKeys.historyRevision(deploymentId ?? "", revision ?? 0),
    queryFn: () => deploymentService.getDeploymentHistoryRevision(deploymentId ?? "", revision ?? 0),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(deploymentId) && Boolean(revision),
  });
}

export function useCreateDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateDeploymentRequest) => deploymentService.createDeployment(payload),
    onSuccess: () => {
      toast.success("Deployment created successfully.");
      return Promise.all([
        queryClient.invalidateQueries({ queryKey: deploymentKeys.all }),
        queryClient.invalidateQueries({ queryKey: projectDeploymentKeys.all }),
      ]);
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to create deployment."));
    },
  });
}

export function useUpdateDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateDeploymentRequest }) =>
      deploymentService.updateDeployment(id, payload),
    onSuccess: (deployment) => {
      toast.success("Deployment updated successfully.");
      queryClient.setQueryData(deploymentKeys.detail(deployment.id), deployment);
      return Promise.all([
        queryClient.invalidateQueries({ queryKey: deploymentKeys.all }),
        queryClient.invalidateQueries({ queryKey: projectDeploymentKeys.all }),
      ]);
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update deployment."));
    },
  });
}

export function useDeleteDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deploymentService.deleteDeployment(id),
    onSuccess: () => {
      toast.success("Deployment deleted successfully.");
      return Promise.all([
        queryClient.invalidateQueries({ queryKey: deploymentKeys.all }),
        queryClient.invalidateQueries({ queryKey: projectDeploymentKeys.all }),
      ]);
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to delete deployment."));
    },
  });
}

export function useRollbackDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, revision }: { id: string; revision: number }) =>
      deploymentService.rollbackDeployment(id, revision),
    onSuccess: () => {
      toast.success("Deployment rolled back successfully.");
      return Promise.all([
        queryClient.invalidateQueries({ queryKey: deploymentKeys.all }),
        queryClient.invalidateQueries({ queryKey: projectDeploymentKeys.all }),
      ]);
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to roll back deployment."));
    },
  });
}

export function useCancelDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deploymentService.cancelDeployment(id),
    onSuccess: () => {
      toast.success("Deployment cancelled successfully.");
      return Promise.all([
        queryClient.invalidateQueries({ queryKey: deploymentKeys.all }),
        queryClient.invalidateQueries({ queryKey: projectDeploymentKeys.all }),
      ]);
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to cancel deployment."));
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
