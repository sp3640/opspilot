import axios from "axios";
import { env } from "./env";
import { useAuthStore } from "@/store/auth-store";

export const api = axios.create({
  baseURL: env.API_URL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("accessToken");

    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }

  return config;
});

// Requests whose own callers already handle 401 as part of a controlled auth
// transition (login failing with a wrong password, the initial bootstrap
// check on a stale/missing token, a register call that never 401s). A 401
// from anywhere else means a previously-authenticated session's token has
// become invalid mid-use, which the global handler below reacts to.
const AUTH_TRANSITION_PATHS = ["/auth/login", "/auth/register", "/users/me"];

const PUBLIC_AUTH_PAGES = ["/login", "/register"];

// A GitHub proxy endpoint (every route containing "/github/" - identity,
// repositories, commits, pulls, application mapping, deployment
// correlation) calls out to GitHub on OpsPilot's behalf and can 401 for a
// reason that has nothing to do with the caller's own OpsPilot session: the
// *stored GitHub credential* was rejected or revoked (backend:
// apperrors.ErrGitHubUnauthorized, "GitHub credentials are invalid or have
// been revoked"). That must surface as an integration-specific error to the
// component that asked for it, never as an OpsPilot session expiry - doing
// the latter would incorrectly log the user out of OpsPilot over a GitHub-
// side problem. OpsPilot's own auth middleware never produces this message,
// so matching on it (scoped to /github/ paths, as defense in depth) safely
// distinguishes the two without touching backend behavior.
const GITHUB_CREDENTIAL_ERROR_MESSAGE = "GitHub credentials are invalid";

export function isGitHubCredentialError(error: unknown): boolean {
  if (!axios.isAxiosError<{ message?: string }>(error) || error.response?.status !== 401) {
    return false;
  }
  const url = error.config?.url ?? "";
  const message = error.response?.data?.message ?? "";
  return url.includes("/github/") && message.includes(GITHUB_CREDENTIAL_ERROR_MESSAGE);
}

// Guards against every in-flight request rejecting at once (e.g. several
// dashboard widgets losing their session together) from each independently
// clearing state and redirecting.
let isHandlingSessionExpiry = false;

function isAuthTransitionRequest(url: string | undefined): boolean {
  if (!url) return false;
  return AUTH_TRANSITION_PATHS.some((path) => url.startsWith(path));
}

function handleSessionExpired() {
  if (typeof window === "undefined" || isHandlingSessionExpiry) {
    return;
  }
  isHandlingSessionExpiry = true;

  useAuthStore.getState().logout();

  const isOnPublicAuthPage = PUBLIC_AUTH_PAGES.includes(window.location.pathname);
  if (isOnPublicAuthPage) {
    isHandlingSessionExpiry = false;
    return;
  }

  const intendedDestination = window.location.pathname + window.location.search;
  const params = new URLSearchParams();
  if (intendedDestination && intendedDestination !== "/") {
    params.set("redirect", intendedDestination);
  }
  params.set("sessionExpired", "1");

  window.location.assign(`/login?${params.toString()}`);
}

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (
      axios.isAxiosError(error) &&
      error.response?.status === 401 &&
      !isAuthTransitionRequest(error.config?.url) &&
      !isGitHubCredentialError(error)
    ) {
      handleSessionExpired();
    }

    // Never log the raw error object: axios attaches the full request
    // config to it, including the Authorization header (the OpsPilot JWT).
    // Log only what's safe to see in a console.
    if (axios.isAxiosError(error)) {
      console.error("API Error:", error.config?.method?.toUpperCase(), error.config?.url, error.response?.status, error.response?.data?.message);
    } else {
      console.error("API Error:", error);
    }

    return Promise.reject(error);
  }
);
