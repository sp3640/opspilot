// Sprint 28: GitHub connector types. Every response here is the safe,
// backend-mapped shape - never GitHub's raw API JSON, and never a
// credential (see internal/dto/github.go for the backend contract).

export type GitHubRepositoryResponse = {
  id: string;
  integrationId: string;
  githubId: number;
  owner: string;
  name: string;
  fullName: string;
  url: string;
  defaultBranch: string;
  private: boolean;
  selected: boolean;
  createdAt: string;
  updatedAt: string;
};

export type GitHubRepositoryListResponse = {
  items: GitHubRepositoryResponse[];
  total: number;
};

export type ApplicationRepositoryMappingResponse = {
  applicationId: string;
  repository: GitHubRepositoryResponse;
  createdAt: string;
  updatedAt: string;
};

export type GitHubCommitResponse = {
  sha: string;
  message: string;
  authorName: string;
  authorEmail?: string;
  authorLogin?: string;
  timestamp: string;
  url: string;
};

export type GitHubCommitListResponse = {
  items: GitHubCommitResponse[];
};

export type GitHubPullRequestResponse = {
  number: number;
  title: string;
  state: string;
  authorLogin: string;
  createdAt: string;
  updatedAt: string;
  mergedAt?: string;
  url: string;
  headSha: string;
  baseBranch: string;
};

export type GitHubPullRequestListResponse = {
  items: GitHubPullRequestResponse[];
};

// Available=false covers every case where OpsPilot does not have enough
// trustworthy information to correlate a deployment with a commit/PR - it
// is never inferred from timestamp proximity alone. See
// docs (Sprint 28) / GitHubService.GetDeploymentCorrelation.
export type DeploymentGitHubCorrelationResponse = {
  deploymentId: string;
  commitSha?: string;
  available: boolean;
  reason?: string;
  commit?: GitHubCommitResponse;
  pullRequests?: GitHubPullRequestResponse[];
};

export type GitHubOAuthStartResponse = {
  authorizeUrl: string;
};
