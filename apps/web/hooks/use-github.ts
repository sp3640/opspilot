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
  repositories: (integrationId: string) => ["github", "repositories", integrationId] as const,
  commits: (repositoryId: string, branch?: string) => ["github", "commits", repositoryId, branch ?? ""] as const,
  pullRequests: (repositoryId: string) => ["github", "pull-requests", repositoryId] as const,
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

// GitHub repositories are a pure DB read by default (no GitHub call) - a
// long staleTime keeps the frontend from ever polling GitHub just because a
// component re-rendered. useSyncGitHubRepositories is the explicit,
// user-initiated action that actually calls GitHub.
export function useGitHubRepositories(integrationId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: githubKeys.repositories(integrationId ?? ""),
    queryFn: () => githubService.listRepositories(integrationId ?? ""),
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
      queryClient.setQueryData(githubKeys.repositories(integrationId), result);
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
      return queryClient.invalidateQueries({ queryKey: githubKeys.repositories(repository.integrationId) });
    },
    onError: (error) => {
      toast.error(getMutationErrorMessage(error, "Failed to update repository."));
    },
  });
}

export function useGitHubCommits(repositoryId: string | null, branch?: string, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: githubKeys.commits(repositoryId ?? "", branch),
    queryFn: () => githubService.listCommits(repositoryId ?? "", { branch, limit: 20 }),
    enabled: queryEnabled && Boolean(accessToken) && Boolean(repositoryId),
    staleTime: 60_000,
  });
}

export function useGitHubPullRequests(repositoryId: string | null, queryEnabled = true) {
  const accessToken = useAuthStore((state) => state.accessToken);

  return useQuery({
    queryKey: githubKeys.pullRequests(repositoryId ?? ""),
    queryFn: () => githubService.listPullRequests(repositoryId ?? "", { limit: 20 }),
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
