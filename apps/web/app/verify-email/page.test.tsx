// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import VerifyEmailPage from "./page";
import { apiError, apiOk, installMockApi } from "@/lib/testing/mock-api";
import type { MockHandler } from "@/lib/testing/mock-api";

const push = vi.fn();
const replace = vi.fn();

// Same contract as Next: the search params reflect the current URL.
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, replace }),
  useSearchParams: () => new URLSearchParams(window.location.search),
}));

const TOKEN = "A".repeat(43);
let restore: (() => void) | undefined;

function openLink(search: string) {
  window.history.replaceState(null, "", `/verify-email${search}`);
}

function serve(handler: MockHandler) {
  const mock = installMockApi(handler);
  restore = mock.restore;
  return mock.requests;
}

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  restore?.();
  restore = undefined;
  window.history.replaceState(null, "", "/");
});

describe("VerifyEmailPage", () => {
  it("verifies the emailed token, once, then offers to continue to login", async () => {
    openLink(`?token=${TOKEN}`);
    const requests = serve(() => apiOk("Email verified successfully.", { status: "verified" }));

    render(<VerifyEmailPage />);

    expect(await screen.findByRole("heading", { name: /email verified successfully/i })).toBeInTheDocument();
    expect(requests).toEqual([{ method: "POST", url: "/auth/verify-email", body: { token: TOKEN } }]);

    await userEvent.setup().click(screen.getByRole("button", { name: /continue to login/i }));
    expect(push).toHaveBeenCalledWith("/login");
  });

  it("does not create a session or store anything", async () => {
    openLink(`?token=${TOKEN}`);
    serve(() => apiOk("Email verified successfully.", { status: "verified" }));

    render(<VerifyEmailPage />);
    await screen.findByRole("heading", { name: /email verified successfully/i });

    expect(localStorage.getItem("accessToken")).toBeNull();
    expect(replace).not.toHaveBeenCalled();
  });

  it("removes the token from the address bar and never renders it", async () => {
    openLink(`?token=${TOKEN}`);
    serve(() => apiOk("Email verified successfully.", { status: "verified" }));

    const { container } = render(<VerifyEmailPage />);
    await screen.findByRole("heading", { name: /email verified successfully/i });

    expect(window.location.search).toBe("");
    expect(container.textContent).not.toContain(TOKEN);
  });

  it("says so when the account was already verified", async () => {
    openLink(`?token=${TOKEN}`);
    serve(() => apiOk("This email address is already verified.", { status: "already_verified" }));

    render(<VerifyEmailPage />);

    expect(await screen.findByRole("heading", { name: /already verified/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /continue to login/i })).toBeInTheDocument();
  });

  it("shows an invalid link state with a way to request a new one", async () => {
    openLink(`?token=${TOKEN}`);
    serve(() => apiError(400, "This verification link is invalid.", "verification_token_invalid"));

    render(<VerifyEmailPage />);

    expect(await screen.findByRole("heading", { name: /verification link is invalid/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /resend verification email/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /continue to login/i })).not.toBeInTheDocument();
  });

  it("shows an expired link state with a way to request a new one", async () => {
    openLink(`?token=${TOKEN}`);
    serve(() => apiError(410, "This verification link has expired.", "verification_token_expired"));

    render(<VerifyEmailPage />);

    expect(await screen.findByRole("heading", { name: /verification link has expired/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /resend verification email/i })).toBeInTheDocument();
  });

  it("treats a link with no token as invalid without calling the server", async () => {
    openLink("");
    const requests = serve(() => apiOk("unexpected"));

    render(<VerifyEmailPage />);

    expect(await screen.findByRole("heading", { name: /verification link is invalid/i })).toBeInTheDocument();
    expect(requests).toHaveLength(0);
  });

  it("recovers from a network failure without losing the token", async () => {
    openLink(`?token=${TOKEN}`);
    let attempt = 0;
    const requests = serve(() => {
      attempt += 1;
      return attempt === 1 ? { networkError: true } : apiOk("Email verified successfully.", { status: "verified" });
    });

    render(<VerifyEmailPage />);

    expect(await screen.findByRole("heading", { name: /couldn't verify your email right now/i })).toBeInTheDocument();
    // The URL was already cleaned, so the retry must use the token kept in memory.
    expect(window.location.search).toBe("");

    await userEvent.setup().click(screen.getByRole("button", { name: /try again/i }));

    expect(await screen.findByRole("heading", { name: /email verified successfully/i })).toBeInTheDocument();
    expect(requests.map((request) => request.body)).toEqual([{ token: TOKEN }, { token: TOKEN }]);
  });

  it("reports a server failure as retryable rather than as an invalid link", async () => {
    openLink(`?token=${TOKEN}`);
    serve(() => apiError(500, "Internal server error"));

    render(<VerifyEmailPage />);

    expect(await screen.findByRole("heading", { name: /couldn't verify your email right now/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /try again/i })).toBeInTheDocument();
  });

  it("lets someone with an expired link request a new email", async () => {
    openLink(`?token=${TOKEN}`);
    const requests = serve((request) =>
      request.url === "/auth/verify-email"
        ? apiError(410, "expired", "verification_token_expired")
        : apiOk("If that account is awaiting verification, a new verification link is on its way.")
    );

    const user = userEvent.setup();
    render(<VerifyEmailPage />);

    await user.type(await screen.findByLabelText(/^email$/i), "  Jane@Example.com ");
    await user.click(screen.getByRole("button", { name: /resend verification email/i }));

    expect(await screen.findByRole("status")).toHaveTextContent(/new link is on its way/i);
    await waitFor(() =>
      expect(requests.at(-1)).toEqual({ method: "POST", url: "/auth/resend-verification", body: { email: "Jane@Example.com" } })
    );
  });
});
