// @vitest-environment jsdom
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import LoginPage from "./page";
import { authService } from "@/services/auth-service";
import { useAuthStore } from "@/store/auth-store";
import type { LoginResponse } from "@/types/auth";

const push = vi.fn();
const replace = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, replace }),
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock("@/services/auth-service", () => ({
  authService: {
    login: vi.fn(),
    me: vi.fn(),
  },
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ accessToken: null, currentUser: null, isBootstrapping: false });
});

describe("LoginPage", () => {
  it("renders the sign-in form", async () => {
    render(<LoginPage />);

    expect(await screen.findByRole("heading", { name: /sign in to opspilot/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
  });

  it("shows validation errors for empty submission", async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    await user.click(await screen.findByRole("button", { name: /sign in/i }));

    expect(await screen.findByText(/invalid email address/i)).toBeInTheDocument();
    expect(authService.login).not.toHaveBeenCalled();
  });

  it("toggles password visibility", async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    const passwordInput = await screen.findByLabelText(/^password$/i);
    expect(passwordInput).toHaveAttribute("type", "password");

    await user.click(screen.getByRole("button", { name: /show password/i }));
    expect(passwordInput).toHaveAttribute("type", "text");
  });

  it("shows a loading state while submitting", async () => {
    let resolveLogin: (value: LoginResponse) => void = () => {};
    vi.mocked(authService.login).mockReturnValue(
      new Promise<LoginResponse>((resolve) => {
        resolveLogin = resolve;
      })
    );

    const user = userEvent.setup();
    render(<LoginPage />);

    await user.type(await screen.findByLabelText(/^email$/i), "user@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    expect(await screen.findByRole("button", { name: /sign in/i })).toBeDisabled();

    resolveLogin({ success: true, message: "ok", data: { access_token: "test-token", token_type: "Bearer" } });
  });

  it("shows an error message on auth failure", async () => {
    vi.mocked(authService.login).mockRejectedValue({
      isAxiosError: true,
      response: { data: { message: "Invalid email or password" } },
    });

    const user = userEvent.setup();
    render(<LoginPage />);

    await user.type(await screen.findByLabelText(/^email$/i), "user@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "wrong-password");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Invalid email or password");
  });

  it("logs in successfully and redirects", async () => {
    vi.mocked(authService.login).mockResolvedValue({
      success: true,
      message: "ok",
      data: { access_token: "test-token", token_type: "Bearer" },
    });
    vi.mocked(authService.me).mockResolvedValue({
      success: true,
      message: "ok",
      data: { id: 1, email: "user@example.com", name: "Test User", role: "Viewer" },
    });

    const user = userEvent.setup();
    render(<LoginPage />);

    await user.type(await screen.findByLabelText(/^email$/i), "user@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() => expect(useAuthStore.getState().accessToken).toBe("test-token"));
    expect(replace).toHaveBeenCalledWith("/");
  });
});
