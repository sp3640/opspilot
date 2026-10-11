// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import RegisterPage from "./page";
import { useAuthStore } from "@/store/auth-store";
import { apiError, apiOk, installMockApi } from "@/lib/testing/mock-api";
import type { MockHandler } from "@/lib/testing/mock-api";

const push = vi.fn();
const replace = vi.fn();
let search = "";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, replace }),
  useSearchParams: () => new URLSearchParams(search),
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
  search = "";
  useAuthStore.setState({ accessToken: null, currentUser: null, isBootstrapping: false });
});

afterEach(() => {
  restore?.();
  restore = undefined;
});

async function fillAndSubmit(email = "Jane@Example.com") {
  const user = userEvent.setup();
  await user.type(await screen.findByLabelText(/^name$/i), "Jane Doe");
  await user.type(screen.getByLabelText(/^email$/i), email);
  await user.type(screen.getByLabelText(/^password$/i), "password123");
  await user.type(screen.getByLabelText(/confirm password/i), "password123");
  await user.click(screen.getByRole("button", { name: /create account/i }));
  return user;
}

describe("RegisterPage email verification", () => {
  it("moves to a 'Check your email' state instead of signing in or redirecting", async () => {
    const requests = serve((request) =>
      request.url === "/auth/register" ? { status: 201, body: { success: true, message: "Check your email for a link to verify your account." } } : apiOk("resent")
    );
    render(<RegisterPage />);

    await fillAndSubmit();

    expect(await screen.findByRole("heading", { name: /check your email/i })).toBeInTheDocument();
    expect(screen.getByText(/we've sent a verification link to/i)).toHaveTextContent("jane@example.com");
    expect(screen.getByRole("button", { name: /resend verification email/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /back to login/i })).toHaveAttribute("href", "/login");

    // The form is gone, no session was created, and nothing redirected.
    expect(screen.queryByLabelText(/^password$/i)).not.toBeInTheDocument();
    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(replace).not.toHaveBeenCalled();
    expect(push).not.toHaveBeenCalled();
    expect(requests).toHaveLength(1);
    expect(requests[0]).toMatchObject({ method: "POST", url: "/auth/register" });
  });

  it("resends the verification email for the address just registered", async () => {
    const requests = serve((request) =>
      request.url === "/auth/register" ? { status: 201, body: { success: true, message: "ok" } } : apiOk("accepted")
    );
    render(<RegisterPage />);
    const user = await fillAndSubmit();

    await user.click(await screen.findByRole("button", { name: /resend verification email/i }));

    expect(await screen.findByRole("status")).toHaveTextContent(/new link is on its way/i);
    expect(requests.at(-1)).toEqual({ method: "POST", url: "/auth/resend-verification", body: { email: "jane@example.com" } });
  });

  it("keeps the invitation redirect on 'Back to login'", async () => {
    search = "redirect=%2Faccept-invitation%3Ftoken%3Dinvite-token";
    serve(() => ({ status: 201, body: { success: true, message: "ok" } }));
    render(<RegisterPage />);

    await fillAndSubmit();

    const link = await screen.findByRole("link", { name: /back to login/i });
    expect(link).toHaveAttribute("href", `/login?redirect=${encodeURIComponent("/accept-invitation?token=invite-token")}`);
  });

  it("does not claim success when the verification email could not be sent", async () => {
    serve(() => apiError(503, "We couldn't send the verification email right now. Please try again in a few minutes.", "email_delivery_unavailable"));
    render(<RegisterPage />);

    await fillAndSubmit();

    expect(await screen.findByRole("alert")).toHaveTextContent(/couldn't send the verification email/i);
    expect(screen.queryByRole("heading", { name: /check your email/i })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create account/i })).toBeEnabled();
  });
});
