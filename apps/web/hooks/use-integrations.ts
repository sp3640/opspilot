"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { integrationService } from "@/services/integration-service";
import type { CreateIntegrationRequest, UpdateIntegrationRequest } from "@/types/integration-api";

const integrationKeys = {
  all: ["integrations"] as const,
  list: () => ["integrations", "list"] as const,
  detail: (id: string) => ["integrations", "detail", id] as const,
};

export function useIntegrations(queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: integrationKeys.list(),
    queryFn: () => integrationService.listIntegrations(),
    enabled: queryEnabled && Boolean(accessToken),
  });
}

export function useIntegration(id: string | null) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: integrationKeys.detail(id ?? ""),
    queryFn: () => integrationService.getIntegration(id ?? ""),
    enabled: Boolean(accessToken) && Boolean(id),
  });
}

export function useCreateIntegration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateIntegrationRequest) => integrationService.createIntegration(payload),
    onSuccess: () => {
      toast.success("Integration created successfully.");
      return queryClient.invalidateQueries({ queryKey: integrationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to create integration."));
    },
  });
}

export function useUpdateIntegration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateIntegrationRequest }) =>
      integrationService.updateIntegration(id, payload),
    onSuccess: (integration) => {
      toast.success("Integration updated successfully.");
      queryClient.setQueryData(integrationKeys.detail(integration.id), integration);
      return queryClient.invalidateQueries({ queryKey: integrationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update integration."));
    },
  });
}

export function useDeleteIntegration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => integrationService.deleteIntegration(id),
    onSuccess: () => {
      toast.success("Integration deleted successfully.");
      return queryClient.invalidateQueries({ queryKey: integrationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to delete integration."));
    },
  });
}

export function useTestIntegration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => integrationService.testIntegration(id),
    onSuccess: (result) => {
      if (result.success) {
        toast.success(result.message || "Connection test succeeded.");
      } else {
        toast.error(result.message || "Connection test failed.");
      }
      queryClient.setQueryData(integrationKeys.detail(result.integration.id), result.integration);
      return queryClient.invalidateQueries({ queryKey: integrationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to test integration."));
    },
  });
}

export function useCheckIntegration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => integrationService.checkIntegration(id),
    onSuccess: (result) => {
      if (result.success) {
        toast.success(result.message || "Status check succeeded.");
      } else {
        toast.error(result.message || "Status check failed.");
      }
      queryClient.setQueryData(integrationKeys.detail(result.integration.id), result.integration);
      return queryClient.invalidateQueries({ queryKey: integrationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to check integration status."));
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
