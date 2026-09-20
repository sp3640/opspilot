// @vitest-environment jsdom
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import RegisterPage from "./page";
import { authService } from "@/services/auth-service";
import { useAuthStore } from "@/store/auth-store";
import type { RegisterResponse } from "@/types/auth";

const push = vi.fn();
const replace = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, replace }),
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock("@/services/auth-service", () => ({
  authService: {
    register: vi.fn(),
  },
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ accessToken: null, currentUser: null, isBootstrapping: false });
});

describe("RegisterPage", () => {
  it("renders the registration form", async () => {
    render(<RegisterPage />);

    expect(await screen.findByRole("heading", { name: /create your opspilot account/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/^name$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument();
  });

  it("shows validation errors for an invalid submission", async () => {
    const user = userEvent.setup();
    render(<RegisterPage />);

    await user.click(await screen.findByRole("button", { name: /create account/i }));

    expect(await screen.findByText(/name must be at least 2 characters/i)).toBeInTheDocument();
    expect(authService.register).not.toHaveBeenCalled();
  });

  it("flags mismatched passwords", async () => {
    const user = userEvent.setup();
    render(<RegisterPage />);

    await user.type(await screen.findByLabelText(/^name$/i), "Jane Doe");
    await user.type(screen.getByLabelText(/^email$/i), "jane@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    await user.type(screen.getByLabelText(/confirm password/i), "different123");
    await user.click(screen.getByRole("button", { name: /create account/i }));

    expect(await screen.findByText(/passwords do not match/i, { selector: "#confirm-password-error" })).toBeInTheDocument();
    expect(authService.register).not.toHaveBeenCalled();
  });

  it("shows a loading state while submitting", async () => {
    let resolveRegister: (value: RegisterResponse) => void = () => {};
    vi.mocked(authService.register).mockReturnValue(
      new Promise<RegisterResponse>((resolve) => {
        resolveRegister = resolve;
      })
    );

    const user = userEvent.setup();
    render(<RegisterPage />);

    await user.type(await screen.findByLabelText(/^name$/i), "Jane Doe");
    await user.type(screen.getByLabelText(/^email$/i), "jane@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    await user.type(screen.getByLabelText(/confirm password/i), "password123");
    await user.click(screen.getByRole("button", { name: /create account/i }));

    expect(await screen.findByRole("button", { name: /create account/i })).toBeDisabled();
    resolveRegister({ success: true, message: "ok", data: null });
  });

  it("shows an error message on registration failure", async () => {
    vi.mocked(authService.register).mockRejectedValue({
      isAxiosError: true,
      response: { data: { message: "Email already registered" } },
    });

    const user = userEvent.setup();
    render(<RegisterPage />);

    await user.type(await screen.findByLabelText(/^name$/i), "Jane Doe");
    await user.type(screen.getByLabelText(/^email$/i), "jane@example.com");
    await user.type(screen.getByLabelText(/^password$/i), "password123");
    await user.type(screen.getByLabelText(/confirm password/i), "password123");
    await user.click(screen.getByRole("button", { name: /create account/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Email already registered");
  });
});
