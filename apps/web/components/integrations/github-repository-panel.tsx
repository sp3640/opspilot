"use client";

import { useState } from "react";
import { ExternalLink, GitCommitHorizontal, GitPullRequest } from "lucide-react";

import { EmptyState, StatusBadge } from "@/components/common";
import { Button } from "@/components/ui/button";
import {
  useGitHubCommits,
  useGitHubIdentity,
  useGitHubPullRequests,
  useGitHubRepositories,
  useSelectGitHubRepository,
  useSyncGitHubRepositories,
} from "@/hooks/use-github";
import type { GitHubRepositoryResponse } from "@/types/github-api";

/**
 * Repository discovery/selection for a connected GitHub integration
 * (Sprint 28). ListRepositories is a pure DB read (cheap, cached); "Sync"
 * is the explicit action that actually calls GitHub - the UI never hits
 * GitHub on its own render.
 */
export function GitHubRepositoryPanel({ integrationId, canManage }: { integrationId: string; canManage: boolean }) {
  const { data: identity } = useGitHubIdentity(integrationId);
  const { data, isLoading } = useGitHubRepositories(integrationId);
  const sync = useSyncGitHubRepositories();
  const [expandedRepositoryId, setExpandedRepositoryId] = useState<string | null>(null);

  const items = data?.items ?? [];

  return (
    <div className="mt-3 space-y-3 border-t pt-3" style={{ borderColor: "var(--border)" }}>
      {identity && (
        <div className="flex items-center gap-2 text-sm" style={{ color: "var(--muted-foreground)" }}>
          {identity.avatarUrl && <img src={identity.avatarUrl} alt="" className="h-6 w-6 rounded-full" />}
          <span>Connected as</span>
          {identity.profileUrl ? (
            <a href={identity.profileUrl} target="_blank" rel="noreferrer" className="font-medium" style={{ color: "var(--primary)" }}>
              {identity.login}
            </a>
          ) : (
            <span className="font-medium">{identity.login}</span>
          )}
        </div>
      )}
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wider" style={{ color: "var(--muted-foreground)" }}>
          Repositories
        </span>
        {canManage && (
          <Button
            type="button"
            variant="ghost"
            loading={sync.isPending}
            onClick={() => void sync.mutateAsync(integrationId)}
          >
            Sync repositories
          </Button>
        )}
      </div>

      {isLoading ? (
        <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
          Loading...
        </p>
      ) : items.length === 0 ? (
        <EmptyState
          icon={GitCommitHorizontal}
          title="No repositories discovered yet"
          description={canManage ? "Sync repositories to discover what this GitHub account can access." : "Ask an organization admin to sync repositories."}
        />
      ) : (
        <ul className="space-y-2">
          {items.map((repo) => (
            <li key={repo.id} className="rounded-xl border p-3" style={{ borderColor: "var(--border)" }}>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <span className="font-medium">{repo.fullName}</span>
                  {repo.private && <StatusBadge variant="archived">Private</StatusBadge>}
                  {repo.selected && <StatusBadge variant="success">Selected</StatusBadge>}
                  <a
                    href={repo.url}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center"
                    style={{ color: "var(--muted-foreground)" }}
                    aria-label={`Open ${repo.fullName} on GitHub`}
                  >
                    <ExternalLink aria-hidden="true" className="h-3.5 w-3.5" />
                  </a>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => setExpandedRepositoryId(expandedRepositoryId === repo.id ? null : repo.id)}
                  >
                    {expandedRepositoryId === repo.id ? "Hide activity" : "Recent activity"}
                  </Button>
                  {canManage && <SelectToggle repo={repo} />}
                </div>
              </div>
              <p className="mt-1 text-xs" style={{ color: "var(--muted-foreground)" }}>
                Default branch: {repo.defaultBranch}
              </p>

              {expandedRepositoryId === repo.id && <RepositoryActivity repositoryId={repo.id} branch={repo.defaultBranch} />}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function SelectToggle({ repo }: { repo: GitHubRepositoryResponse }) {
  const select = useSelectGitHubRepository();
  return (
    <Button
      type="button"
      variant="secondary"
      loading={select.isPending}
      onClick={() => void select.mutateAsync({ repositoryId: repo.id, selected: !repo.selected })}
    >
      {repo.selected ? "Unselect" : "Select"}
    </Button>
  );
}

/** Read-only recent commits + pull requests - operational, not a GitHub clone. */
function RepositoryActivity({ repositoryId, branch }: { repositoryId: string; branch: string }) {
  const { data: commitsData, isLoading: isCommitsLoading } = useGitHubCommits(repositoryId, branch);
  const { data: pullsData, isLoading: isPullsLoading } = useGitHubPullRequests(repositoryId);

  return (
    <div className="mt-3 grid gap-3 sm:grid-cols-2">
      <div>
        <p className="mb-1 flex items-center gap-1.5 text-xs font-semibold" style={{ color: "var(--muted-foreground)" }}>
          <GitCommitHorizontal aria-hidden="true" className="h-3.5 w-3.5" />
          Recent commits
        </p>
        {isCommitsLoading ? (
          <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>Loading...</p>
        ) : (commitsData?.items.length ?? 0) === 0 ? (
          <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>No commits found.</p>
        ) : (
          <ul className="space-y-1.5">
            {(commitsData?.items ?? []).slice(0, 5).map((commit) => (
              <li key={commit.sha} className="text-xs">
                <a href={commit.url} target="_blank" rel="noreferrer" className="font-mono" style={{ color: "var(--primary)" }}>
                  {commit.sha.slice(0, 7)}
                </a>{" "}
                {commit.message.split("\n")[0]}
                <span style={{ color: "var(--muted-foreground)" }}>
                  {" "}
                  · {commit.authorName} · {new Date(commit.timestamp).toLocaleString()}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>

      <div>
        <p className="mb-1 flex items-center gap-1.5 text-xs font-semibold" style={{ color: "var(--muted-foreground)" }}>
          <GitPullRequest aria-hidden="true" className="h-3.5 w-3.5" />
          Recent pull requests
        </p>
        {isPullsLoading ? (
          <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>Loading...</p>
        ) : (pullsData?.items.length ?? 0) === 0 ? (
          <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>No pull requests found.</p>
        ) : (
          <ul className="space-y-1.5">
            {(pullsData?.items ?? []).slice(0, 5).map((pr) => (
              <li key={pr.number} className="text-xs">
                <a href={pr.url} target="_blank" rel="noreferrer" style={{ color: "var(--primary)" }}>
                  #{pr.number}
                </a>{" "}
                {pr.title}
                <span style={{ color: "var(--muted-foreground)" }}>
                  {" "}
                  · {pr.state}
                  {pr.mergedAt ? ` · merged ${new Date(pr.mergedAt).toLocaleString()}` : ""}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
