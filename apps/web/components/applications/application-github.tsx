"use client";

import { useState } from "react";
import axios from "axios";
import { ExternalLink, GitCommitHorizontal, GitPullRequest } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useIntegrations } from "@/hooks/use-integrations";
import { useGitHubRepositories } from "@/hooks/use-github";
import {
  useApplicationGitHubMapping,
  useGitHubCommits,
  useGitHubPullRequests,
  useMapApplicationRepository,
  useUnmapApplicationRepository,
} from "@/hooks/use-github";
import { useHasPermission } from "@/store/auth-store";

/**
 * Application <-> GitHub repository mapping and read-only recent activity
 * (Sprint 28). Reuses the exact same hooks as the Integrations workspace -
 * no separate GitHub data-fetching logic lives here.
 */
export function ApplicationGitHub({ applicationId }: { applicationId: string }) {
  const canManage = useHasPermission("organization:manage");
  const {
    data: mapping,
    error: mappingError,
    isError: isMappingError,
    isLoading: isMappingLoading,
  } = useApplicationGitHubMapping(applicationId);
  const notMapped = axios.isAxiosError(mappingError) && mappingError.response?.status === 404;
  const unmap = useUnmapApplicationRepository();

  if (isMappingLoading) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        Loading...
      </p>
    );
  }

  if (isMappingError && !notMapped) {
    return (
      <p className="text-sm" style={{ color: "var(--danger)" }}>
        {mappingError instanceof Error ? mappingError.message : "Unable to load the GitHub repository mapping."}
      </p>
    );
  }

  return (
    <div className="space-y-4">
      {mapping ? (
        <div className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border p-4" style={{ borderColor: "var(--border)" }}>
          <div>
            <p className="font-semibold">{mapping.repository.fullName}</p>
            <a
              href={mapping.repository.url}
              target="_blank"
              rel="noreferrer"
              className="mt-1 inline-flex items-center gap-1 text-xs"
              style={{ color: "var(--primary)" }}
            >
              View on GitHub <ExternalLink aria-hidden="true" className="h-3 w-3" />
            </a>
          </div>
          {canManage && (
            <Button
              type="button"
              variant="ghost"
              className="text-[var(--danger)]"
              loading={unmap.isPending}
              onClick={() => void unmap.mutateAsync(applicationId)}
            >
              Unmap
            </Button>
          )}
        </div>
      ) : (
        <RepositoryPicker applicationId={applicationId} canManage={canManage} />
      )}

      {mapping && <RecentActivity repositoryId={mapping.repository.id} branch={mapping.repository.defaultBranch} />}
    </div>
  );
}

function RepositoryPicker({ applicationId, canManage }: { applicationId: string; canManage: boolean }) {
  const { data: integrationsData } = useIntegrations();
  const githubIntegration = (integrationsData?.items ?? []).find((integration) => integration.type === "github");
  const { data: reposData, isLoading } = useGitHubRepositories(githubIntegration?.id ?? null, Boolean(githubIntegration));
  const map = useMapApplicationRepository();
  const [selectedRepositoryId, setSelectedRepositoryId] = useState("");

  if (!githubIntegration) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        No GitHub integration is connected for this organization yet. Connect one from Organization settings.
      </p>
    );
  }

  const repos = reposData?.items ?? [];

  return (
    <div className="space-y-2">
      <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        No GitHub repository is mapped to this application yet.
      </p>
      {canManage && (
        <div className="flex flex-wrap items-center gap-2">
          <select
            value={selectedRepositoryId}
            onChange={(event) => setSelectedRepositoryId(event.target.value)}
            disabled={isLoading || repos.length === 0}
            className="h-10 rounded-xl border bg-transparent px-3 text-sm outline-none focus:border-[var(--primary)] disabled:opacity-60"
            style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
          >
            <option value="">{repos.length === 0 ? "No repositories discovered" : "Select a repository"}</option>
            {repos.map((repo) => (
              <option key={repo.id} value={repo.id}>
                {repo.fullName}
              </option>
            ))}
          </select>
          <Button
            type="button"
            variant="secondary"
            loading={map.isPending}
            disabled={!selectedRepositoryId}
            onClick={() => void map.mutateAsync({ applicationId, repositoryId: selectedRepositoryId })}
          >
            Map repository
          </Button>
        </div>
      )}
    </div>
  );
}

function RecentActivity({ repositoryId, branch }: { repositoryId: string; branch: string }) {
  const { data: commitsData, isLoading: isCommitsLoading } = useGitHubCommits(repositoryId, branch);
  const { data: pullsData, isLoading: isPullsLoading } = useGitHubPullRequests(repositoryId);

  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <div>
        <p className="mb-2 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          <GitCommitHorizontal aria-hidden="true" className="h-3.5 w-3.5" />
          Recent commits
        </p>
        {isCommitsLoading ? (
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Loading...</p>
        ) : (commitsData?.items.length ?? 0) === 0 ? (
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>No commits found.</p>
        ) : (
          <ul className="space-y-2">
            {(commitsData?.items ?? []).slice(0, 5).map((commit) => (
              <li key={commit.sha} className="rounded-xl border p-3 text-sm" style={{ borderColor: "var(--border)" }}>
                <a href={commit.url} target="_blank" rel="noreferrer" className="font-mono text-xs" style={{ color: "var(--primary)" }}>
                  {commit.sha.slice(0, 7)}
                </a>
                <p className="mt-1">{commit.message.split("\n")[0]}</p>
                <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  {commit.authorName} · {new Date(commit.timestamp).toLocaleString()}
                </p>
              </li>
            ))}
          </ul>
        )}
      </div>

      <div>
        <p className="mb-2 flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          <GitPullRequest aria-hidden="true" className="h-3.5 w-3.5" />
          Recent pull requests
        </p>
        {isPullsLoading ? (
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Loading...</p>
        ) : (pullsData?.items.length ?? 0) === 0 ? (
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>No pull requests found.</p>
        ) : (
          <ul className="space-y-2">
            {(pullsData?.items ?? []).slice(0, 5).map((pr) => (
              <li key={pr.number} className="rounded-xl border p-3 text-sm" style={{ borderColor: "var(--border)" }}>
                <a href={pr.url} target="_blank" rel="noreferrer" style={{ color: "var(--primary)" }}>
                  #{pr.number}
                </a>
                <p className="mt-1">{pr.title}</p>
                <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  {pr.state}
                  {pr.mergedAt ? ` · merged ${new Date(pr.mergedAt).toLocaleString()}` : ""}
                </p>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
