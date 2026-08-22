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
    if (axios.isAxiosError(error) && error.response?.status === 401 && !isAuthTransitionRequest(error.config?.url)) {
      handleSessionExpired();
    }

    console.error("API Error:", error);

    return Promise.reject(error);
  }
);
