"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { clusterService } from "@/services/cluster-service";
import type {
  ClusterQueryParams,
  CreateClusterRequest,
  UpdateClusterRequest,
} from "@/types/cluster-api";

const clusterKeys = {
  all: ["clusters"] as const,
  list: (params: ClusterQueryParams) => ["clusters", "list", params] as const,
  detail: (id: string) => ["clusters", "detail", id] as const,
};

export function useClusters(params: ClusterQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: clusterKeys.list(params),
    queryFn: () => clusterService.listClusters(params),
    enabled: queryEnabled && Boolean(accessToken),
  });
}

export function useCluster(id: string | null) {
  return useQuery({
    queryKey: clusterKeys.detail(id ?? ""),
    queryFn: () => clusterService.getCluster(id ?? ""),
    enabled: Boolean(id),
  });
}

export function useCreateCluster() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateClusterRequest) => clusterService.createCluster(payload),
    onSuccess: () => {
      toast.success("Cluster created successfully.");
      return queryClient.invalidateQueries({ queryKey: clusterKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to create cluster."));
    },
  });
}

export function useUpdateCluster() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateClusterRequest }) =>
      clusterService.updateCluster(id, payload),
    onSuccess: (cluster) => {
      toast.success("Cluster updated successfully.");
      queryClient.setQueryData(clusterKeys.detail(cluster.id), cluster);
      return queryClient.invalidateQueries({ queryKey: clusterKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update cluster."));
    },
  });
}

export function useDeleteCluster() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => clusterService.deleteCluster(id),
    onSuccess: () => {
      toast.success("Cluster deleted successfully.");
      return queryClient.invalidateQueries({ queryKey: clusterKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to delete cluster."));
    },
  });
}

export function useValidateCluster() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => clusterService.validateCluster(id),
    onSuccess: (result) => {
      if (result.connected) {
        toast.success("Cluster validation successful.");
      } else {
        toast.warning("Cluster validation failed.");
      }
      if (result.cluster) {
        queryClient.setQueryData(clusterKeys.detail(result.cluster.id), result.cluster);
      }
      return queryClient.invalidateQueries({ queryKey: clusterKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to validate cluster."));
    },
  });
}

export function useSetDefaultCluster() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => clusterService.setDefaultCluster(id),
    onSuccess: (cluster) => {
      toast.success("Default cluster updated.");
      queryClient.setQueryData(clusterKeys.detail(cluster.id), cluster);
      return queryClient.invalidateQueries({ queryKey: clusterKeys.all });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to set default cluster."));
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
