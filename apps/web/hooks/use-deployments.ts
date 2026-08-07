"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { deploymentService } from "@/services/deployment-service";
import type { DeploymentQueryParams } from "@/types/deployment-api";

const deploymentKeys = {
  all: ["deployments"] as const,
  list: (applicationId: string, params: DeploymentQueryParams) =>
    ["deployments", "list", applicationId, params] as const,
  history: (deploymentId: string) => ["deployments", "history", deploymentId] as const,
};

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

export function useDeploymentHistory(deploymentId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: deploymentKeys.history(deploymentId ?? ""),
    queryFn: () => deploymentService.getDeploymentHistory(deploymentId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(deploymentId),
  });
}

export function useRollbackDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, revision }: { id: string; revision: number }) =>
      deploymentService.rollbackDeployment(id, revision),
    onSuccess: () => {
      toast.success("Deployment rolled back successfully.");
      return queryClient.invalidateQueries({ queryKey: deploymentKeys.all });
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
      return queryClient.invalidateQueries({ queryKey: deploymentKeys.all });
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
