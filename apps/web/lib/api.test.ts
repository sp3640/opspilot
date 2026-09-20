// @vitest-environment jsdom
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import type { AxiosRequestConfig, AxiosResponse } from "axios";
import { AxiosError } from "axios";

// A per-request adapter override runs the real interceptor pipeline (the
// exact code under test) against a synthetic response, with no network call
// and no need to reimplement or mock the interceptor itself.
function respondWith(status: number, data: unknown) {
  return async (config: AxiosRequestConfig): Promise<AxiosResponse> => {
    const response: AxiosResponse = { data, status, statusText: "", headers: {}, config: config as never };
    if (status >= 200 && status < 300) {
      return response;
    }
    throw new AxiosError("Request failed", String(status), config as never, null, response);
  };
}

// lib/api.ts tracks "already redirecting to login" as module-level state
// that (by design - see its own comment) only resets on a real page
// navigation, which jsdom doesn't perform for a mocked
// window.location.assign. Each test below therefore gets its own fresh
// module instance (and a fresh, independent auth store) via
// vi.resetModules() + a dynamic import, exactly like a real fresh page load
// would - so one test triggering the redirect can never mask whether a
// later test would have triggered it too.
async function freshApi() {
  vi.resetModules();
  const apiModule = await import("@/lib/api");
  const authModule = await import("@/store/auth-store");
  return { ...apiModule, useAuthStore: authModule.useAuthStore };
}

beforeEach(() => {
  localStorage.setItem("accessToken", "existing-opspilot-token");

  Object.defineProperty(window, "location", {
    writable: true,
    configurable: true,
    value: { ...window.location, pathname: "/organization", search: "", assign: vi.fn() },
  });
});

afterEach(() => {
  localStorage.clear();
  vi.restoreAllMocks();
});

describe("api response interceptor - OpsPilot session expiry vs. GitHub credential errors", () => {
  it("A: a normal OpsPilot 401 (invalid/expired token) still clears the session and redirects to login", async () => {
    const { api, useAuthStore } = await freshApi();
    useAuthStore.setState({ accessToken: "existing-opspilot-token" });

    await expect(
      api.get("/integrations", {
        adapter: respondWith(401, { success: false, message: "invalid or expired token" }),
      })
    ).rejects.toBeTruthy();

    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(localStorage.getItem("accessToken")).toBeNull();
    expect(window.location.assign).toHaveBeenCalledTimes(1);
    expect(window.location.assign).toHaveBeenCalledWith(expect.stringContaining("/login"));
  });

  it("B: a GitHub identity 401 does not clear the token, redirect to login, or navigate at all", async () => {
    const { api, useAuthStore } = await freshApi();
    useAuthStore.setState({ accessToken: "existing-opspilot-token" });

    await expect(
      api.get("/integrations/int-1/github/identity", {
        adapter: respondWith(401, { success: false, message: "GitHub credentials are invalid or have been revoked" }),
      })
    ).rejects.toBeTruthy();

    expect(useAuthStore.getState().accessToken).toBe("existing-opspilot-token");
    expect(localStorage.getItem("accessToken")).toBe("existing-opspilot-token");
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("B (other GitHub proxy routes): a repositories-sync 401 with the same GitHub message is also exempted", async () => {
    const { api, useAuthStore } = await freshApi();
    useAuthStore.setState({ accessToken: "existing-opspilot-token" });

    await expect(
      api.post("/integrations/int-1/github/repositories/sync", undefined, {
        adapter: respondWith(401, { success: false, message: "GitHub credentials are invalid or have been revoked" }),
      })
    ).rejects.toBeTruthy();

    expect(useAuthStore.getState().accessToken).toBe("existing-opspilot-token");
    expect(window.location.assign).not.toHaveBeenCalled();
  });

  it("does not exempt a genuine OpsPilot 401 just because it happens to hit a /github/ path", async () => {
    const { api, useAuthStore } = await freshApi();
    useAuthStore.setState({ accessToken: "existing-opspilot-token" });

    await expect(
      api.get("/integrations/int-1/github/identity", {
        adapter: respondWith(401, { success: false, message: "invalid or expired token" }),
      })
    ).rejects.toBeTruthy();

    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(window.location.assign).toHaveBeenCalledTimes(1);
  });
});

describe("isGitHubCredentialError", () => {
  function axiosError401(url: string, message: string): unknown {
    const config = { url } as AxiosRequestConfig;
    return new AxiosError("Request failed", "401", config as never, null, {
      data: { message },
      status: 401,
      statusText: "",
      headers: {},
      config: config as never,
    });
  }

  it("is true for a GitHub proxy path with the GitHub credential message", async () => {
    const { isGitHubCredentialError } = await freshApi();
    expect(
      isGitHubCredentialError(axiosError401("/integrations/int-1/github/identity", "GitHub credentials are invalid or have been revoked"))
    ).toBe(true);
  });

  it("is false for a non-github path even with a similar message", async () => {
    const { isGitHubCredentialError } = await freshApi();
    expect(isGitHubCredentialError(axiosError401("/integrations", "GitHub credentials are invalid or have been revoked"))).toBe(false);
  });

  it("is false for a github path with an unrelated 401 message", async () => {
    const { isGitHubCredentialError } = await freshApi();
    expect(isGitHubCredentialError(axiosError401("/integrations/int-1/github/identity", "invalid or expired token"))).toBe(false);
  });

  it("is false for a non-error value", async () => {
    const { isGitHubCredentialError } = await freshApi();
    expect(isGitHubCredentialError(new Error("boom"))).toBe(false);
    expect(isGitHubCredentialError(null)).toBe(false);
  });
});
