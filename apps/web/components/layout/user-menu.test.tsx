// @vitest-environment jsdom
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { UserMenu } from "./user-menu";
import { organizationService } from "@/services/organization-service";
import { useAuthStore } from "@/store/auth-store";

const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));

vi.mock("@/services/organization-service", () => ({
  organizationService: {
    listOrganizations: vi.fn(),
  },
}));

function renderWithProviders() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <UserMenu />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  vi.mocked(organizationService.listOrganizations).mockResolvedValue({
    items: [{ id: "org-1", name: "Acme Inc", slug: "acme", description: "", ownerId: 1, createdAt: "", updatedAt: "" }],
    page: 1,
    limit: 1,
    total: 1,
    totalPages: 1,
  });
  localStorage.setItem("accessToken", "test-token");
  useAuthStore.setState({
    accessToken: "test-token",
    currentUser: { id: 1, email: "jane@example.com", name: "Jane Doe", role: "Platform Admin" },
    isBootstrapping: false,
  });
});

describe("UserMenu", () => {
  it("renders the current user's name, email, and role on the trigger", () => {
    renderWithProviders();

    expect(screen.getByText("Jane Doe")).toBeInTheDocument();
    expect(screen.getByText("jane@example.com")).toBeInTheDocument();
    expect(screen.getByText("Platform Admin")).toBeInTheDocument();
  });

  it("opens the dropdown and shows a Settings link and a Log out action, with no dead Profile link", async () => {
    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /account menu for jane doe/i }));

    expect(await screen.findByRole("menuitem", { name: /settings/i })).toHaveAttribute("href", "/settings");
    expect(screen.getByRole("menuitem", { name: /log out/i })).toBeInTheDocument();
    expect(screen.queryByText(/^profile$/i)).not.toBeInTheDocument();
  });

  it("closes the dropdown when clicking outside", async () => {
    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /account menu for jane doe/i }));
    expect(await screen.findByRole("menu")).toBeInTheDocument();

    await user.click(document.body);
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  it("logout clears the persisted token and in-memory auth state, and redirects to /login", async () => {
    const user = userEvent.setup();
    renderWithProviders();

    await user.click(screen.getByRole("button", { name: /account menu for jane doe/i }));
    await user.click(await screen.findByRole("menuitem", { name: /log out/i }));

    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(useAuthStore.getState().currentUser).toBeNull();
    expect(localStorage.getItem("accessToken")).toBeNull();
    expect(push).toHaveBeenCalledWith("/login");
  });
});
