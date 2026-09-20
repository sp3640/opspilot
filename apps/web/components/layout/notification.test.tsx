// @vitest-environment jsdom
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { NotificationButton } from "./notification";
import { incidentService } from "@/services/incident-service";
import type { IncidentListResponse } from "@/types/incident-api";

const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));

vi.mock("@/services/incident-service", () => ({
  incidentService: { listIncidents: vi.fn() },
}));

function renderWithProviders() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <NotificationButton />
    </QueryClientProvider>
  );
}

function incident(overrides: Partial<{ id: number; title: string; status: string; severity: string; updatedAt: string }> = {}) {
  return {
    id: 1,
    title: "API latency spike",
    description: "",
    severity: "P1",
    status: "OPEN",
    projectId: "p1",
    createdAt: "2026-09-20T00:00:00Z",
    updatedAt: "2026-09-20T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
});

describe("NotificationButton", () => {
  it("opens reliably when clicked", async () => {
    vi.mocked(incidentService.listIncidents).mockResolvedValue({ items: [], page: 1, limit: 10, total: 0, totalPages: 0 });

    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /notifications/i }));

    expect(await screen.findByRole("dialog", { name: /notifications/i })).toBeInTheDocument();
  });

  it("shows a loading state, then real open-incident data (not fake notifications)", async () => {
    let resolveList: (value: IncidentListResponse) => void = () => {};
    vi.mocked(incidentService.listIncidents).mockReturnValue(
      new Promise<IncidentListResponse>((resolve) => {
        resolveList = resolve;
      })
    );

    const user = userEvent.setup();
    renderWithProviders();
    await user.click(screen.getByRole("button", { name: /notifications/i }));

    expect(await screen.findByText(/loading/i)).toBeInTheDocument();

    resolveList({ items: [incident({ id: 7, title: "Database outage" })], page: 1, limit: 10, total: 1, totalPages: 1 });

    expect(await screen.findByText("Database outage")).toBeInTheDocument();
  });

  it("filters out resolved incidents and shows an empty state when nothing is open", async () => {
    vi.mocked(incidentService.listIncidents).mockResolvedValue({
      items: [incident({ id: 1, status: "RESOLVED" })],
      page: 1,
      limit: 10,
      total: 1,
      totalPages: 1,
    });

    const user = userEvent.setup();
    renderWithProviders();
    await user.click(screen.getByRole("button", { name: /notifications/i }));

    expect(await screen.findByText(/no open incidents right now/i)).toBeInTheDocument();
  });

  it("shows an error state with retry when the request fails", async () => {
    vi.mocked(incidentService.listIncidents).mockRejectedValueOnce(new Error("network error"));

    const user = userEvent.setup();
    renderWithProviders();
    await user.click(screen.getByRole("button", { name: /notifications/i }));

    expect(await screen.findByText(/unable to load incidents/i)).toBeInTheDocument();

    vi.mocked(incidentService.listIncidents).mockResolvedValueOnce({
      items: [incident()],
      page: 1,
      limit: 10,
      total: 1,
      totalPages: 1,
    });
    await user.click(screen.getByRole("button", { name: /try again/i }));

    expect(await screen.findByText("API latency spike")).toBeInTheDocument();
  });

  it("selecting a notification navigates to the incident and closes the panel", async () => {
    vi.mocked(incidentService.listIncidents).mockResolvedValue({
      items: [incident({ id: 99, title: "Payment service down" })],
      page: 1,
      limit: 10,
      total: 1,
      totalPages: 1,
    });

    const user = userEvent.setup();
    renderWithProviders();
    await user.click(screen.getByRole("button", { name: /notifications/i }));
    await user.click(await screen.findByText("Payment service down"));

    expect(push).toHaveBeenCalledWith("/incidents/99");
    await waitFor(() => expect(screen.queryByRole("dialog", { name: /notifications/i })).not.toBeInTheDocument());
  });

  it("closes when clicking outside", async () => {
    vi.mocked(incidentService.listIncidents).mockResolvedValue({ items: [], page: 1, limit: 10, total: 0, totalPages: 0 });

    const user = userEvent.setup();
    renderWithProviders();
    await user.click(screen.getByRole("button", { name: /notifications/i }));
    expect(await screen.findByRole("dialog", { name: /notifications/i })).toBeInTheDocument();

    await user.click(document.body);
    expect(screen.queryByRole("dialog", { name: /notifications/i })).not.toBeInTheDocument();
  });
});
