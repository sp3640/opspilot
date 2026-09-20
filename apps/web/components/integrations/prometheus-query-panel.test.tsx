// @vitest-environment jsdom
import type { ReactElement } from "react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { PrometheusQueryPanel } from "./prometheus-query-panel";
import { prometheusService } from "@/services/prometheus-service";
import { useAuthStore } from "@/store/auth-store";
import type { PrometheusQueryResponse } from "@/types/prometheus-api";

vi.mock("@/services/prometheus-service", () => ({
  prometheusService: {
    queryInstant: vi.fn(),
  },
}));

function renderWithProviders(ui: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>);
}

function axiosError(status: number, message: string) {
  return { isAxiosError: true, response: { status, data: { message } } };
}

function successResponse(overrides: Partial<PrometheusQueryResponse> = {}): PrometheusQueryResponse {
  return {
    query: "rate(http_requests_total[5m])",
    samples: [
      {
        timestamp: "2026-09-17T12:00:00Z",
        value: 4.2,
        labels: { __name__: "http_requests_total", pod: "api-7f9", namespace: "prod" },
      },
    ],
    correlation: { available: false, reason: "no deterministic application match was found" },
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ accessToken: "test-access-token", currentUser: null });
});

async function runQuery(queryText = "rate(http_requests_total[5m])") {
  const user = userEvent.setup();
  const input = screen.getByLabelText("PromQL query");
  await user.type(input, queryText);
  await user.click(screen.getByRole("button", { name: /run query/i }));
  return user;
}

describe("PrometheusQueryPanel", () => {
  it("shows an 'integration not connected' state and never calls the backend when disconnected", () => {
    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="DISCONNECTED" />);

    expect(screen.getByText(/Connect this integration before running Prometheus queries/)).toBeInTheDocument();
    expect(screen.queryByLabelText("PromQL query")).not.toBeInTheDocument();
    expect(prometheusService.queryInstant).not.toHaveBeenCalled();
  });

  it("calls the correct backend service when Run query is clicked", async () => {
    vi.mocked(prometheusService.queryInstant).mockResolvedValue(successResponse());

    renderWithProviders(<PrometheusQueryPanel integrationId="int-42" integrationStatus="CONNECTED" />);
    await runQuery("up");

    await waitFor(() => {
      expect(prometheusService.queryInstant).toHaveBeenCalledWith("int-42", "up");
    });
  });

  it("shows a loading state while the query is in flight", async () => {
    let resolveQuery: (value: PrometheusQueryResponse) => void = () => {};
    vi.mocked(prometheusService.queryInstant).mockReturnValue(
      new Promise((resolve) => {
        resolveQuery = resolve;
      })
    );

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery();

    expect(await screen.findByRole("status")).toHaveTextContent(/Running query/);

    resolveQuery(successResponse());
    await waitFor(() => expect(screen.queryByRole("status")).not.toBeInTheDocument());
  });

  it("renders a successful metric result", async () => {
    vi.mocked(prometheusService.queryInstant).mockResolvedValue(successResponse());

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery();

    expect(await screen.findByText("http_requests_total")).toBeInTheDocument();
    expect(screen.getByText(/pod="api-7f9"/)).toBeInTheDocument();
    expect(screen.getByText(/namespace="prod"/)).toBeInTheDocument();
    expect(screen.getByText("4.2")).toBeInTheDocument();
  });

  it("shows an empty-result message when no samples are returned", async () => {
    vi.mocked(prometheusService.queryInstant).mockResolvedValue(successResponse({ samples: [] }));

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery();

    expect(await screen.findByText("No data returned for this query.")).toBeInTheDocument();
  });

  it("shows a human-readable message for invalid PromQL", async () => {
    vi.mocked(prometheusService.queryInstant).mockRejectedValue(
      axiosError(502, "prometheus query error: parse error at char 3: unexpected end of input")
    );

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery("rate(");

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("Invalid PromQL query. Check the query syntax and try again.");
    expect(alert.textContent).not.toContain("parse error");
  });

  it("shows a human-readable message when Prometheus is unavailable", async () => {
    vi.mocked(prometheusService.queryInstant).mockRejectedValue(
      axiosError(502, "call Prometheus API: dial tcp 10.0.0.5:9090: connect: connection refused")
    );

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery("up");

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("Prometheus is unavailable right now. Please try again later.");
    expect(alert.textContent).not.toContain("10.0.0.5");
  });

  it("shows a human-readable message for an unauthorized response, without leaking credentials", async () => {
    vi.mocked(prometheusService.queryInstant).mockRejectedValue(
      axiosError(502, "prometheus api returned status 401")
    );

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery("up");

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("Prometheus rejected the request as unauthorized. Check the integration's credentials.");
    expect(alert.textContent).not.toMatch(/bearer|token/i);
  });

  it("shows a generic message for other backend errors, without a stack trace", async () => {
    vi.mocked(prometheusService.queryInstant).mockRejectedValue(
      axiosError(500, "internal server error")
    );

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery("up");

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("Something went wrong running this query. Please try again.");
  });

  it("displays a correlated resource when the backend reports one", async () => {
    vi.mocked(prometheusService.queryInstant).mockResolvedValue(
      successResponse({
        correlation: {
          available: true,
          resource: { id: "res-1", kind: "Pod", name: "api-7f9", namespace: "prod" },
        },
      })
    );

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery();

    expect(await screen.findByText(/Resource: Pod api-7f9 \(prod\)/)).toBeInTheDocument();
    expect(screen.queryByText("Correlation unavailable")).not.toBeInTheDocument();
  });

  it("displays a correlated application when the backend reports one", async () => {
    vi.mocked(prometheusService.queryInstant).mockResolvedValue(
      successResponse({
        correlation: {
          available: true,
          application: { id: "app-1", name: "checkout-service" },
        },
      })
    );

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery();

    expect(await screen.findByText(/Application: checkout-service/)).toBeInTheDocument();
    expect(screen.queryByText("Correlation unavailable")).not.toBeInTheDocument();
  });

  it("displays 'Correlation unavailable' when the backend cannot correlate", async () => {
    vi.mocked(prometheusService.queryInstant).mockResolvedValue(
      successResponse({ correlation: { available: false, reason: "ambiguous match" } })
    );

    renderWithProviders(<PrometheusQueryPanel integrationId="int-1" integrationStatus="CONNECTED" />);
    await runQuery();

    expect(await screen.findByText("Correlation unavailable")).toBeInTheDocument();
  });
});
