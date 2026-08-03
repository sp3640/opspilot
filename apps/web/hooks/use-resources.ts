"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { resourceService } from "@/services/resource-service";
import { useAuthStore } from "@/store/auth-store";
import type { CreateResourceRequest, ResourceQueryParams, SyncResourcesRequest, UpdateResourceRequest } from "@/types/resource-api";

const resourceKeys = { all: ["resources"] as const, list: (params: ResourceQueryParams) => ["resources", "list", params] as const, detail: (id: string) => ["resources", "detail", id] as const };
const message = (error: unknown, fallback: string) => axios.isAxiosError<{ message?: string }>(error) ? error.response?.data?.message ?? fallback : error instanceof Error ? error.message : fallback;

export function useResources(params: ResourceQueryParams, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);
  return useQuery({ queryKey: resourceKeys.list(params), queryFn: () => resourceService.listResources(params), enabled: queryEnabled && Boolean(accessToken) });
}
export function useResource(id: string | null) { return useQuery({ queryKey: resourceKeys.detail(id ?? ""), queryFn: () => resourceService.getResource(id ?? ""), enabled: Boolean(id) }); }
export function useCreateResource() { const client = useQueryClient(); return useMutation({ mutationFn: (payload: CreateResourceRequest) => resourceService.createResource(payload), onSuccess: () => { toast.success("Resource created successfully."); return client.invalidateQueries({ queryKey: resourceKeys.all }); }, onError: (error) => toast.error(message(error, "Failed to create resource.")) }); }
export function useUpdateResource() { const client = useQueryClient(); return useMutation({ mutationFn: ({ id, payload }: { id: string; payload: UpdateResourceRequest }) => resourceService.updateResource(id, payload), onSuccess: (resource) => { toast.success("Resource updated successfully."); client.setQueryData(resourceKeys.detail(resource.id), resource); return client.invalidateQueries({ queryKey: resourceKeys.all }); }, onError: (error) => toast.error(message(error, "Failed to update resource.")) }); }
export function useDeleteResource() { const client = useQueryClient(); return useMutation({ mutationFn: (id: string) => resourceService.deleteResource(id), onSuccess: () => { toast.success("Resource deleted successfully."); return client.invalidateQueries({ queryKey: resourceKeys.all }); }, onError: (error) => toast.error(message(error, "Failed to delete resource.")) }); }
export function useSyncResources() { const client = useQueryClient(); return useMutation({ mutationFn: (payload: SyncResourcesRequest) => resourceService.syncResources(payload), onSuccess: (result) => { toast.success(`Resources synchronized: ${result.created} created, ${result.updated} updated, ${result.restored} restored, ${result.deleted} deleted.`); return client.invalidateQueries({ queryKey: resourceKeys.all }); }, onError: (error) => toast.error(message(error, "Failed to synchronize resources.")) }); }
