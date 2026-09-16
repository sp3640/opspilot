import { api } from "@/lib/api";
import type {
  ApplicationRepositoryMappingResponse,
  DeploymentGitHubCorrelationResponse,
  GitHubIdentityResponse,
  GitHubCommitListResponse,
  GitHubOAuthStartResponse,
  GitHubPullRequestListResponse,
  GitHubRepositoryListResponse,
  GitHubRepositoryResponse,
} from "@/types/github-api";

type APIResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export const githubService = {
  async startOAuth(): Promise<GitHubOAuthStartResponse> {
    const response = await api.get<APIResponse<GitHubOAuthStartResponse>>("/github/oauth/start");
    return response.data.data;
  },

  async getIdentity(integrationId: string): Promise<GitHubIdentityResponse> {
    const response = await api.get<APIResponse<GitHubIdentityResponse>>(
      `/integrations/${integrationId}/github/identity`
    );
    return response.data.data;
  },

  async listRepositories(integrationId: string, params?: { page?: number; limit?: number }): Promise<GitHubRepositoryListResponse> {
    const response = await api.get<APIResponse<GitHubRepositoryListResponse>>(
      `/integrations/${integrationId}/github/repositories`,
      { params }
    );
    return response.data.data;
  },

  async syncRepositories(integrationId: string): Promise<GitHubRepositoryListResponse> {
    const response = await api.post<APIResponse<GitHubRepositoryListResponse>>(
      `/integrations/${integrationId}/github/repositories/sync`
    );
    return response.data.data;
  },

  async selectRepository(repositoryId: string, selected: boolean): Promise<GitHubRepositoryResponse> {
    const response = await api.patch<APIResponse<GitHubRepositoryResponse>>(
      `/github/repositories/${repositoryId}/select`,
      { selected }
    );
    return response.data.data;
  },

  async listCommits(repositoryId: string, params?: { branch?: string; page?: number; limit?: number }): Promise<GitHubCommitListResponse> {
    const response = await api.get<APIResponse<GitHubCommitListResponse>>(
      `/github/repositories/${repositoryId}/commits`,
      { params }
    );
    return response.data.data;
  },

  async listPullRequests(repositoryId: string, params?: { page?: number; limit?: number; state?: "open" | "closed" | "all" }): Promise<GitHubPullRequestListResponse> {
    const response = await api.get<APIResponse<GitHubPullRequestListResponse>>(
      `/github/repositories/${repositoryId}/pulls`,
      { params }
    );
    return response.data.data;
  },

  async getApplicationMapping(applicationId: string): Promise<ApplicationRepositoryMappingResponse> {
    const response = await api.get<APIResponse<ApplicationRepositoryMappingResponse>>(
      `/applications/${applicationId}/github/repository`
    );
    return response.data.data;
  },

  async mapApplicationRepository(applicationId: string, repositoryId: string): Promise<ApplicationRepositoryMappingResponse> {
    const response = await api.put<APIResponse<ApplicationRepositoryMappingResponse>>(
      `/applications/${applicationId}/github/repository`,
      { repository_id: repositoryId }
    );
    return response.data.data;
  },

  async unmapApplicationRepository(applicationId: string): Promise<void> {
    await api.delete<APIResponse<null>>(`/applications/${applicationId}/github/repository`);
  },

  async getDeploymentCorrelation(deploymentId: string): Promise<DeploymentGitHubCorrelationResponse> {
    const response = await api.get<APIResponse<DeploymentGitHubCorrelationResponse>>(
      `/deployments/${deploymentId}/github/correlation`
    );
    return response.data.data;
  },
};
