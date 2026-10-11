// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { RESEND_COOLDOWN_SECONDS, ResendVerification } from "./resend-verification";
import { apiError, apiOk, installMockApi } from "@/lib/testing/mock-api";
import type { MockHandler } from "@/lib/testing/mock-api";

let restore: (() => void) | undefined;

function serve(handler: MockHandler) {
  const mock = installMockApi(handler);
  restore = mock.restore;
  return mock.requests;
}

afterEach(() => {
  restore?.();
  restore = undefined;
  vi.useRealTimers();
});

const accepted = () => apiOk("If that account is awaiting verification, a new verification link is on its way.");

describe("ResendVerification", () => {
  it("asks the backend for a new link for the known address and confirms without claiming delivery", async () => {
    const requests = serve(accepted);
    render(<ResendVerification email="jane@example.com" />);

    // The address is known, so there is nothing to type.
    expect(screen.queryByLabelText(/^email$/i)).not.toBeInTheDocument();

    await userEvent.setup().click(screen.getByRole("button", { name: /resend verification email/i }));

    expect(await screen.findByRole("status")).toHaveTextContent(/if that account is awaiting verification/i);
    expect(requests).toEqual([{ method: "POST", url: "/auth/resend-verification", body: { email: "jane@example.com" } }]);
  });

  it("locks the button for the cooldown, with a countdown, then re-enables it", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const requests = serve(accepted);
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<ResendVerification email="jane@example.com" />);

    await user.click(screen.getByRole("button", { name: /resend verification email/i }));

    const locked = await screen.findByRole("button", { name: /resend available in \d+s/i });
    expect(locked).toBeDisabled();
    expect(locked).toHaveTextContent(String(RESEND_COOLDOWN_SECONDS));

    await user.click(locked);
    expect(requests).toHaveLength(1);

    // The countdown re-arms itself after every render, so step it a second
    // at a time, as the clock would.
    for (let elapsed = 0; elapsed <= RESEND_COOLDOWN_SECONDS; elapsed += 1) {
      await act(async () => {
        await vi.advanceTimersByTimeAsync(1000);
      });
    }
    expect(screen.getByRole("button", { name: /resend verification email/i })).toBeEnabled();
  });

  it("asks for an address when none is known, and validates it before calling the server", async () => {
    const requests = serve(accepted);
    const user = userEvent.setup();
    render(<ResendVerification />);

    await user.click(screen.getByRole("button", { name: /resend verification email/i }));
    expect(await screen.findByRole("alert")).toHaveTextContent(/enter the email address/i);
    expect(requests).toHaveLength(0);

    await user.type(screen.getByLabelText(/^email$/i), "jane@example.com");
    await user.click(screen.getByRole("button", { name: /resend verification email/i }));
    expect(await screen.findByRole("status")).toBeInTheDocument();
    expect(requests).toHaveLength(1);
  });

  it("tells the user when they are rate limited, and locks the button", async () => {
    serve(() => apiError(429, "rate limit exceeded"));
    render(<ResendVerification email="jane@example.com" />);

    await userEvent.setup().click(screen.getByRole("button", { name: /resend verification email/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/too many requests/i);
    expect(screen.getByRole("button", { name: /resend available in/i })).toBeDisabled();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });

  it("surfaces email being unavailable instead of pretending it was sent", async () => {
    serve(() => apiError(503, "We couldn't send the verification email right now. Please try again in a few minutes.", "email_delivery_unavailable"));
    render(<ResendVerification email="jane@example.com" />);

    await userEvent.setup().click(screen.getByRole("button", { name: /resend verification email/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/couldn't send the verification email/i);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /resend verification email/i })).toBeEnabled();
  });

  it("shows a friendly message when the request never reaches the server", async () => {
    serve(() => ({ networkError: true }));
    render(<ResendVerification email="jane@example.com" />);

    await userEvent.setup().click(screen.getByRole("button", { name: /resend verification email/i }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(/couldn't send the email right now/i);
    expect(alert).not.toHaveTextContent(/network error/i);
  });
});
