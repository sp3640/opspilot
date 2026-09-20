// @vitest-environment jsdom
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { CommandPalette } from "./command-palette";
import { SearchBar } from "@/components/layout/search-bar";
import { incidentService } from "@/services/incident-service";
import { projectService } from "@/services/project-service";
import { alertService } from "@/services/alert-service";
import { useCommandPaletteStore } from "@/store/command-palette-store";

const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));

vi.mock("@/services/incident-service", () => ({
  incidentService: { listIncidents: vi.fn() },
}));
vi.mock("@/services/project-service", () => ({
  projectService: { listProjects: vi.fn() },
}));
vi.mock("@/services/alert-service", () => ({
  alertService: { listAlerts: vi.fn() },
}));

function renderWithProviders() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <SearchBar />
      <CommandPalette />
    </QueryClientProvider>
  );
}

const emptyList = { items: [], page: 1, limit: 5, total: 0, totalPages: 0 };

beforeEach(() => {
  vi.clearAllMocks();
  useCommandPaletteStore.setState({ isOpen: false });
  vi.mocked(incidentService.listIncidents).mockResolvedValue(emptyList);
  vi.mocked(projectService.listProjects).mockResolvedValue(emptyList);
  vi.mocked(alertService.listAlerts).mockResolvedValue(emptyList);
});

describe("SearchBar + CommandPalette", () => {
  it("the header search control is a real, labeled button that opens the palette (previously a decorative dead input)", async () => {
    const user = userEvent.setup();
    renderWithProviders();

    const trigger = screen.getByRole("button", { name: /search projects, incidents, and alerts/i });
    await user.click(trigger);

    expect(await screen.findByRole("dialog", { name: /command palette/i })).toBeInTheDocument();
  });

  it("shows the static navigate/quick-action commands when no query is entered", async () => {
    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /search projects, incidents, and alerts/i }));

    expect(await screen.findByRole("option", { name: /incidents/i })).toBeInTheDocument();
    expect(incidentService.listIncidents).not.toHaveBeenCalled();
  });

  it("accepts typed input and does not fire a search API call below the minimum query length", async () => {
    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /search projects, incidents, and alerts/i }));
    const input = await screen.findByLabelText(/search command palette/i);
    await user.type(input, "a");

    expect(input).toHaveValue("a");
    await new Promise((resolve) => setTimeout(resolve, 350));
    expect(incidentService.listIncidents).not.toHaveBeenCalled();
  });

  it("runs a debounced search across incidents/projects/alerts and does not call on every keystroke", async () => {
    vi.mocked(incidentService.listIncidents).mockResolvedValue({
      items: [{ id: 42, title: "Database outage", description: "", severity: "P0", status: "OPEN", projectId: "p1", createdAt: "", updatedAt: "" }],
      page: 1,
      limit: 5,
      total: 1,
      totalPages: 1,
    });

    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /search projects, incidents, and alerts/i }));
    const input = await screen.findByLabelText(/search command palette/i);
    await user.type(input, "outage");

    expect(await screen.findByText("Database outage")).toBeInTheDocument();
    // Debounced to one call for the whole typed string, not one per keystroke.
    expect(incidentService.listIncidents).toHaveBeenCalledTimes(1);
    expect(incidentService.listIncidents).toHaveBeenCalledWith(expect.objectContaining({ search: "outage" }));
  });

  it("selecting an incident result navigates to its detail route", async () => {
    vi.mocked(incidentService.listIncidents).mockResolvedValue({
      items: [{ id: 42, title: "Database outage", description: "", severity: "P0", status: "OPEN", projectId: "p1", createdAt: "", updatedAt: "" }],
      page: 1,
      limit: 5,
      total: 1,
      totalPages: 1,
    });

    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /search projects, incidents, and alerts/i }));
    await user.type(await screen.findByLabelText(/search command palette/i), "outage");
    await user.click(await screen.findByText("Database outage"));

    expect(push).toHaveBeenCalledWith("/incidents/42");
  });

  it("shows a no-results state when nothing matches", async () => {
    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /search projects, incidents, and alerts/i }));
    await user.type(await screen.findByLabelText(/search command palette/i), "nonexistent");

    expect(await screen.findByText(/no incidents, projects, or alerts match/i)).toBeInTheDocument();
  });

  it("shows an error state when the search request fails", async () => {
    vi.mocked(incidentService.listIncidents).mockRejectedValue(new Error("network error"));

    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /search projects, incidents, and alerts/i }));
    await user.type(await screen.findByLabelText(/search command palette/i), "outage");

    await waitFor(() => {
      expect(screen.getByText(/something went wrong running this search/i)).toBeInTheDocument();
    });
  });
});
