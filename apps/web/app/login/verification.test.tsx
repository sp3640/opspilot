// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import LoginPage from "./page";
import { useAuthStore } from "@/store/auth-store";
import { apiError, apiOk, installMockApi } from "@/lib/testing/mock-api";
import type { MockHandler } from "@/lib/testing/mock-api";

const push = vi.fn();
const replace = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, replace }),
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

let restore: (() => void) | undefined;

function serve(handler: MockHandler) {
  const mock = installMockApi(handler);
  restore = mock.restore;
  return mock.requests;
}

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  useAuthStore.setState({ accessToken: null, currentUser: null, isBootstrapping: false });
});

afterEach(() => {
  restore?.();
  restore = undefined;
});

async function signIn(email: string, password: string) {
  const user = userEvent.setup();
  await user.type(await screen.findByLabelText(/^email$/i), email);
  await user.type(screen.getByLabelText(/^password$/i), password);
  await user.click(screen.getByRole("button", { name: /sign in/i }));
  return user;
}

const unverified = () =>
  apiError(403, "Please verify your email before signing in.", "email_not_verified");

describe("LoginPage email verification", () => {
  it("shows a specific verification-required state, not 'invalid credentials'", async () => {
    serve((request) => (request.url === "/auth/login" ? unverified() : apiOk("accepted")));
    render(<LoginPage />);

    await signIn("Jane@Example.com", "password123");

    const alert = await screen.findByText(/please verify your email before signing in/i);
    expect(alert).toBeInTheDocument();
    expect(screen.queryByText(/invalid email or password/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/login failed/i)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /resend verification email/i })).toBeInTheDocument();

    // No session was created.
    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(localStorage.getItem("accessToken")).toBeNull();
    expect(replace).not.toHaveBeenCalled();
  });

  it("resends the verification email to the address that tried to sign in", async () => {
    const requests = serve((request) => (request.url === "/auth/login" ? unverified() : apiOk("accepted")));
    render(<LoginPage />);
    const user = await signIn("Jane@Example.com", "password123");

    await user.click(await screen.findByRole("button", { name: /resend verification email/i }));

    await waitFor(() =>
      expect(requests.at(-1)).toEqual({ method: "POST", url: "/auth/resend-verification", body: { email: "jane@example.com" } })
    );
    expect(await screen.findByText(/new link is on its way/i)).toBeInTheDocument();
  });

  it("clears the verification state on the next attempt", async () => {
    let attempt = 0;
    serve((request) => {
      if (request.url !== "/auth/login") return apiOk("ok");
      attempt += 1;
      return attempt === 1 ? unverified() : apiError(401, "invalid email or password");
    });
    render(<LoginPage />);
    const user = await signIn("jane@example.com", "password123");
    await screen.findByText(/please verify your email/i);

    await user.click(screen.getByRole("button", { name: /^sign in$/i }));

    // A wrong password gets the ordinary failure, with no verification hint.
    const alerts = await screen.findAllByRole("alert");
    expect(alerts.some((alert) => /invalid email or password/i.test(alert.textContent ?? ""))).toBe(true);
    expect(screen.queryByText(/please verify your email/i)).not.toBeInTheDocument();
  });

  it("still signs a verified user in exactly as before", async () => {
    serve((request) => {
      if (request.url === "/auth/login") {
        return apiOk("Login successful", { access_token: "jwt-for-verified-user", token_type: "Bearer" });
      }
      return apiOk("User fetched successfully", { id: 1, name: "Jane", email: "jane@example.com", role: "Viewer" });
    });
    render(<LoginPage />);

    await signIn("jane@example.com", "password123");

    await waitFor(() => expect(useAuthStore.getState().accessToken).toBe("jwt-for-verified-user"));
    expect(replace).toHaveBeenCalledWith("/");
    expect(within(document.body).queryByText(/verify your email/i)).not.toBeInTheDocument();
  });
});
