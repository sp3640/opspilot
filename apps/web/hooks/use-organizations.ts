"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";

import { organizationService } from "@/services/organization-service";
import { useAuthStore } from "@/store/auth-store";
import type {
  CreateOrganizationRequest,
  OrganizationQueryParams,
  UpdateOrganizationRequest,
} from "@/types/organization-api";

const organizationKeys = {
  all: ["organizations"] as const,
  list: (params: OrganizationQueryParams) => ["organizations", "list", params] as const,
  detail: (id: string) => ["organizations", "detail", id] as const,
};

export function useOrganizations(params: OrganizationQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: organizationKeys.list(params),
    queryFn: () => organizationService.listOrganizations(params),
    enabled: queryEnabled && Boolean(accessToken),
  });
}

export function useOrganization(id: string | null) {
  return useQuery({
    queryKey: organizationKeys.detail(id ?? ""),
    queryFn: () => organizationService.getOrganization(id ?? ""),
    enabled: Boolean(id),
  });
}

export function useCreateOrganization() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateOrganizationRequest) => organizationService.createOrganization(payload),
    onSuccess: () => {
      toast.success("Organization created successfully.");
      return queryClient.invalidateQueries({ queryKey: organizationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to create organization."));
    },
  });
}

export function useUpdateOrganization() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateOrganizationRequest }) =>
      organizationService.updateOrganization(id, payload),
    onSuccess: (organization) => {
      toast.success("Organization updated successfully.");
      queryClient.setQueryData(organizationKeys.detail(organization.id), organization);
      return queryClient.invalidateQueries({ queryKey: organizationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update organization."));
    },
  });
}

export function useDeleteOrganization() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => organizationService.deleteOrganization(id),
    onSuccess: () => {
      toast.success("Organization deleted successfully.");
      return queryClient.invalidateQueries({ queryKey: organizationKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to delete organization."));
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
