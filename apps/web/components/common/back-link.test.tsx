// @vitest-environment jsdom
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { BackLink } from "./back-link";

const back = vi.fn();
const push = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ back, push }),
}));

function setHistoryLength(length: number) {
  Object.defineProperty(window.history, "length", { value: length, configurable: true });
}

function setReferrer(referrer: string) {
  Object.defineProperty(document, "referrer", { value: referrer, configurable: true });
}

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  setHistoryLength(1);
  setReferrer("");
});

describe("BackLink", () => {
  it("uses router.back() when the user navigated here from within the app", async () => {
    setHistoryLength(3);
    setReferrer(window.location.origin + "/incidents");

    const user = userEvent.setup();
    render(<BackLink fallbackHref="/incidents" label="Back to incidents" />);

    await user.click(screen.getByRole("button", { name: /back to incidents/i }));

    expect(back).toHaveBeenCalledTimes(1);
    expect(push).not.toHaveBeenCalled();
  });

  it("falls back to the fixed route when there is no in-app history", async () => {
    setHistoryLength(1);
    setReferrer("");

    const user = userEvent.setup();
    render(<BackLink fallbackHref="/incidents" label="Back to incidents" />);

    await user.click(screen.getByRole("button", { name: /back to incidents/i }));

    expect(push).toHaveBeenCalledWith("/incidents");
    expect(back).not.toHaveBeenCalled();
  });

  it("falls back to the fixed route when the referrer is a different origin", async () => {
    setHistoryLength(3);
    setReferrer("https://external-site.example/somewhere");

    const user = userEvent.setup();
    render(<BackLink fallbackHref="/incidents" label="Back to incidents" />);

    await user.click(screen.getByRole("button", { name: /back to incidents/i }));

    expect(push).toHaveBeenCalledWith("/incidents");
    expect(back).not.toHaveBeenCalled();
  });
});
