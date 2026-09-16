"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth-store";

import { githubService } from "@/services/github-service";

// Separate query-key namespaces per Sprint 28's caching guidance: an
// integration's repositories, one repository's commits, and one
// repository's pull requests each invalidate independently so mapping a
// repository never has to refetch commits/PRs for every other repository.
const githubKeys = {
  identity: (integrationId: string) => ["github", "identity", integrationId] as const,
  repositories: (integrationId: string, page: number, limit: number) => ["github", "repositories", integrationId, page, limit] as const,
  commits: (repositoryId: string, branch: string | undefined, page: number, limit: number) => ["github", "commits", repositoryId, branch ?? "", page, limit] as const,
  pullRequests: (repositoryId: string, state: "open" | "closed" | "all", page: number, limit: number) => ["github", "pull-requests", repositoryId, state, page, limit] as const,
  applicationMapping: (applicationId: string) => ["github", "application-mapping", applicationId] as const,
  deploymentCorrelation: (deploymentId: string) => ["github", "deployment-correlation", deploymentId] as const,
};

export function useStartGitHubOAuth() {
  return useMutation({
    mutationFn: () => githubService.startOAuth(),
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "GitHub is not configured for this deployment."));
    },
  });
}

export function useGitHubIdentity(integrationId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: githubKeys.identity(integrationId ?? ""),
    queryFn: () => githubService.getIdentity(integrationId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(integrationId),
    staleTime: 60_000,
  });
}

// GitHub repositories are a pure DB read by default (no GitHub call) - a
// long staleTime keeps the frontend from ever polling GitHub just because a
// component re-rendered. useSyncGitHubRepositories is the explicit,
// user-initiated action that actually calls GitHub.
export function useGitHubRepositories(integrationId: string | null, queryEnabled = true, page = 1, limit = 20) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: githubKeys.repositories(integrationId ?? "", page, limit),
    queryFn: () => githubService.listRepositories(integrationId ?? "", { page, limit }),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(integrationId),
    staleTime: 60_000,
  });
}

export function useSyncGitHubRepositories() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (integrationId: string) => githubService.syncRepositories(integrationId),
    onSuccess: (result, integrationId) => {
      toast.success(`Discovered ${result.total} repositories.`);
      queryClient.setQueryData(githubKeys.repositories(integrationId, result.page, result.limit), result);
      return queryClient.invalidateQueries({ queryKey: ["integrations"] });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to discover GitHub repositories."));
    },
  });
}

export function useSelectGitHubRepository() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ repositoryId, selected }: { repositoryId: string; selected: boolean }) =>
      githubService.selectRepository(repositoryId, selected),
    onSuccess: (repository) => {
      toast.success(repository.selected ? "Repository selected." : "Repository unselected.");
      return queryClient.invalidateQueries({ queryKey: ["github", "repositories", repository.integrationId] });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update repository."));
    },
  });
}

export function useGitHubCommits(repositoryId: string | null, branch?: string, queryEnabled = true, page = 1, limit = 20) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: githubKeys.commits(repositoryId ?? "", branch, page, limit),
    queryFn: () => githubService.listCommits(repositoryId ?? "", { branch, page, limit }),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(repositoryId),
    staleTime: 60_000,
  });
}

export function useGitHubPullRequests(repositoryId: string | null, queryEnabled = true, state: "open" | "closed" | "all" = "all", page = 1, limit = 20) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: githubKeys.pullRequests(repositoryId ?? "", state, page, limit),
    queryFn: () => githubService.listPullRequests(repositoryId ?? "", { state, page, limit }),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(repositoryId),
    staleTime: 60_000,
  });
}

export function useApplicationGitHubMapping(applicationId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: githubKeys.applicationMapping(applicationId ?? ""),
    queryFn: () => githubService.getApplicationMapping(applicationId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(applicationId),
    retry: false,
  });
}

export function useMapApplicationRepository() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, repositoryId }: { applicationId: string; repositoryId: string }) =>
      githubService.mapApplicationRepository(applicationId, repositoryId),
    onSuccess: (mapping) => {
      toast.success(`Mapped to ${mapping.repository.fullName}.`);
      return queryClient.invalidateQueries({ queryKey: githubKeys.applicationMapping(mapping.applicationId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to map repository."));
    },
  });
}

export function useUnmapApplicationRepository() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (applicationId: string) => githubService.unmapApplicationRepository(applicationId),
    onSuccess: (_result, applicationId) => {
      toast.success("Repository mapping removed.");
      return queryClient.invalidateQueries({ queryKey: githubKeys.applicationMapping(applicationId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to remove repository mapping."));
    },
  });
}

export function useDeploymentGitHubCorrelation(deploymentId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: githubKeys.deploymentCorrelation(deploymentId ?? ""),
    queryFn: () => githubService.getDeploymentCorrelation(deploymentId ?? ""),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(deploymentId),
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
