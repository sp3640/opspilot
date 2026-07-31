"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";

import { incidentService } from "@/services/incident-service";
import type {
  CreateIncidentRequest,
  IncidentQueryParams,
  UpdateIncidentRequest,
} from "@/types/incident-api";

const incidentKeys = {
  all: ["incidents"] as const,
  list: (params: IncidentQueryParams) => ["incidents", "list", params] as const,
  detail: (id: number) => ["incidents", "detail", id] as const,
};

export function useIncidents(params: IncidentQueryParams) {
  return useQuery({
    queryKey: incidentKeys.list(params),
    queryFn: () => incidentService.listIncidents(params),
  });
}

export function useIncident(id: number | null) {
  return useQuery({
    queryKey: incidentKeys.detail(id ?? 0),
    queryFn: () => incidentService.getIncident(id ?? 0),
    enabled: Boolean(id),
  });
}

export function useCreateIncident() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateIncidentRequest) => incidentService.createIncident(payload),
    onSuccess: () => {
      toast.success("Incident created successfully.");
      return queryClient.invalidateQueries({ queryKey: incidentKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to create incident."));
    },
  });
}

export function useUpdateIncident() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: UpdateIncidentRequest }) =>
      incidentService.updateIncident(id, payload),
    onSuccess: (incident) => {
      toast.success("Incident updated successfully.");
      queryClient.setQueryData(incidentKeys.detail(incident.id), incident);
      return queryClient.invalidateQueries({ queryKey: incidentKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update incident."));
    },
  });
}

export function useDeleteIncident() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => incidentService.deleteIncident(id),
    onSuccess: () => {
      toast.success("Incident deleted successfully.");
      return queryClient.invalidateQueries({ queryKey: incidentKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to delete incident."));
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
