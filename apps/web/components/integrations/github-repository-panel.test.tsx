// @vitest-environment jsdom
import type { ReactElement } from "react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AxiosError } from "axios";

import { GitHubRepositoryPanel } from "./github-repository-panel";
import { githubService } from "@/services/github-service";
import { useAuthStore } from "@/store/auth-store";

vi.mock("@/services/github-service", () => ({
  githubService: {
    getIdentity: vi.fn(),
    listRepositories: vi.fn(),
    syncRepositories: vi.fn(),
    selectRepository: vi.fn(),
    listCommits: vi.fn(),
    listPullRequests: vi.fn(),
  },
}));

function renderWithProviders(ui: ReactElement) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>);
}

function githubUnauthorizedError() {
  const config = { url: "/integrations/int-1/github/identity" };
  return new AxiosError("Request failed", "401", config as never, null, {
    data: { success: false, message: "GitHub credentials are invalid or have been revoked" },
    status: 401,
    statusText: "",
    headers: {},
    config: config as never,
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ accessToken: "test-access-token", currentUser: null });
  vi.mocked(githubService.listRepositories).mockResolvedValue({ items: [], page: 1, limit: 20, total: 0, totalPages: 0 });
});

describe("GitHubRepositoryPanel - GitHub credential error display (C)", () => {
  it("shows the integration-specific message when the identity call 401s with an invalid GitHub credential", async () => {
    vi.mocked(githubService.getIdentity).mockRejectedValue(githubUnauthorizedError());

    renderWithProviders(<GitHubRepositoryPanel integrationId="int-1" canManage={true} />);

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("GitHub credentials are invalid. Please reconnect the GitHub integration.");
  });

  it("does not show the credential-error message once identity resolves successfully", async () => {
    vi.mocked(githubService.getIdentity).mockResolvedValue({
      id: 1,
      login: "octocat",
      avatarUrl: "",
      profileUrl: "https://github.com/octocat",
    });

    renderWithProviders(<GitHubRepositoryPanel integrationId="int-1" canManage={true} />);

    await waitFor(() => expect(screen.getByText("octocat")).toBeInTheDocument());
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("does not show the GitHub credential message for an unrelated identity error", async () => {
    const config = { url: "/integrations/int-1/github/identity" };
    vi.mocked(githubService.getIdentity).mockRejectedValue(
      new AxiosError("Request failed", "500", config as never, null, {
        data: { success: false, message: "internal server error" },
        status: 500,
        statusText: "",
        headers: {},
        config: config as never,
      })
    );

    renderWithProviders(<GitHubRepositoryPanel integrationId="int-1" canManage={true} />);

    // Give the query a tick to settle, then assert the specific credential
    // message never appears (a generic 500 is not a GitHub credential error).
    await waitFor(() => expect(githubService.getIdentity).toHaveBeenCalled());
    expect(screen.queryByText(/GitHub credentials are invalid/)).not.toBeInTheDocument();
  });
});
