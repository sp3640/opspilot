// @vitest-environment jsdom
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { IntegrationWorkspace } from "./integration-workspace";
import { integrationService } from "@/services/integration-service";
import { useAuthStore } from "@/store/auth-store";
import type { IntegrationResponse } from "@/types/integration-api";

vi.mock("@/services/integration-service", () => ({
  integrationService: {
    listIntegrations: vi.fn(),
    getIntegration: vi.fn(),
    createIntegration: vi.fn(),
    updateIntegration: vi.fn(),
    deleteIntegration: vi.fn(),
    testIntegration: vi.fn(),
    checkIntegration: vi.fn(),
  },
}));

vi.mock("@/services/github-service", () => ({
  githubService: {
    startOAuth: vi.fn(),
    getIdentity: vi.fn(),
    listRepositories: vi.fn(),
  },
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>);
}

function prometheusIntegration(overrides: Partial<IntegrationResponse> = {}): IntegrationResponse {
  return {
    id: "int-prometheus-1",
    organizationId: "org-1",
    type: "prometheus",
    name: "Prod Prometheus",
    status: "CONNECTED",
    metadata: { api_url: "https://prometheus.example.com" },
    hasCredentials: true,
    connectorImplemented: true,
    connectorDescription: "Query Prometheus-compatible time series data in read-only mode via the HTTP API.",
    lastCheckedAt: "2026-09-17T00:00:00Z",
    createdAt: "2026-09-01T00:00:00Z",
    updatedAt: "2026-09-17T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({
    accessToken: "test-access-token",
    currentUser: {
      id: 1,
      email: "admin@example.com",
      name: "Admin",
      role: "Platform Admin",
    } as never,
  });
});

describe("IntegrationWorkspace", () => {
  it("shows Prometheus as an available integration type and in the list", async () => {
    vi.mocked(integrationService.listIntegrations).mockResolvedValue({
      items: [prometheusIntegration()],
      page: 1,
      limit: 100,
      total: 1,
      totalPages: 1,
    });

    renderWithProviders(<IntegrationWorkspace />);

    // Appears as a connected row.
    await waitFor(() => {
      expect(screen.getByText("Prod Prometheus")).toBeInTheDocument();
    });
    expect(screen.getByText("Prometheus")).toBeInTheDocument();
    expect(screen.getByText("CONNECTED")).toBeInTheDocument();
    // A real connector now exists, so it must never be labeled unsupported.
    expect(screen.queryByText("Not yet supported")).not.toBeInTheDocument();

    // Also selectable as a type when adding a new integration.
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Add integration" }));
    const typeSelect = screen.getByLabelText("Type", { exact: false }) as HTMLSelectElement;
    expect(within(typeSelect).getByRole("option", { name: "Prometheus" })).toBeInTheDocument();
  });

  it("runs the connection test action against the mocked API", async () => {
    vi.mocked(integrationService.listIntegrations).mockResolvedValue({
      items: [prometheusIntegration({ status: "ERROR", lastError: "prometheus api returned status 401" })],
      page: 1,
      limit: 100,
      total: 1,
      totalPages: 1,
    });
    vi.mocked(integrationService.testIntegration).mockResolvedValue({
      success: true,
      message: "Prometheus API is reachable and responding successfully.",
      checkedAt: "2026-09-17T01:00:00Z",
      integration: prometheusIntegration({ status: "CONNECTED" }),
    });

    renderWithProviders(<IntegrationWorkspace />);
    await waitFor(() => expect(screen.getByText("Prod Prometheus")).toBeInTheDocument());

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Test connection" }));

    await waitFor(() => {
      expect(integrationService.testIntegration).toHaveBeenCalledWith("int-prometheus-1");
    });
  });

  it("never renders the stored bearer token after credentials are saved", async () => {
    vi.mocked(integrationService.listIntegrations).mockResolvedValue({
      items: [prometheusIntegration({ hasCredentials: true })],
      page: 1,
      limit: 100,
      total: 1,
      totalPages: 1,
    });

    renderWithProviders(<IntegrationWorkspace />);
    await waitFor(() => expect(screen.getByText("Prod Prometheus")).toBeInTheDocument());

    // The response type never carries a credential value, and the row only
    // ever renders the boolean fact that credentials exist.
    expect(screen.getByText(/Credentials configured/)).toBeInTheDocument();
    expect(screen.queryByText(/bearer/i)).not.toBeInTheDocument();

    // The create form's token field is a masked password input, and it
    // starts (and stays, until the user types) empty - never pre-filled
    // from anything the server returned.
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Add integration" }));
    const typeSelect = screen.getByLabelText("Type", { exact: false });
    await user.selectOptions(typeSelect, "prometheus");

    const tokenInput = screen.getByPlaceholderText("Leave blank if Prometheus requires no auth") as HTMLInputElement;
    expect(tokenInput).toHaveAttribute("type", "password");
    expect(tokenInput.value).toBe("");
  });
});
