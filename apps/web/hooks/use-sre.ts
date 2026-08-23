"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { sreService } from "@/services/sre-service";
import type { ConfigureApplicationSLORequest } from "@/types/sre-api";

const sreKeys = {
  all: ["sre"] as const,
  slo: (applicationId: string) => ["sre", "slo", applicationId] as const,
  metrics: (applicationId: string) => ["sre", "metrics", applicationId] as const,
};

// GetSLO 404s when no SLO has been configured yet for the application - that
// is a normal, expected state (not an error to alarm on), so the component
// checks isError + this same "not configured" signal rather than the hook
// swallowing it, mirroring how ApplicationHealth treats a 404 "not deployed
// yet" response.
export function useApplicationSLO(applicationId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: sreKeys.slo(applicationId ?? ""),
    queryFn: () => sreService.getSLO(applicationId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
    retry: false,
  });
}

export function useApplicationSREMetrics(applicationId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: sreKeys.metrics(applicationId ?? ""),
    queryFn: () => sreService.getMetrics(applicationId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
  });
}

export function useConfigureApplicationSLO() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, payload }: { applicationId: string; payload: ConfigureApplicationSLORequest }) =>
      sreService.configureSLO(applicationId, payload),
    onSuccess: (slo, variables) => {
      toast.success("SLO configuration saved successfully.");
      queryClient.setQueryData(sreKeys.slo(variables.applicationId), slo);
      return queryClient.invalidateQueries({ queryKey: sreKeys.metrics(variables.applicationId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to save SLO configuration."));
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
