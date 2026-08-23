import { create } from "zustand";
import type { User, UserRole } from "@/types/auth";

interface AuthStore {
  accessToken: string | null;
  currentUser: User | null;
  isBootstrapping: boolean;
  setAccessToken: (token: string | null) => void;
  setCurrentUser: (user: User | null) => void;
  setBootstrapping: (isBootstrapping: boolean) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
  accessToken: null,
  currentUser: null,
  isBootstrapping: true,

  setAccessToken: (token: string | null) => {
    if (typeof window !== "undefined") {
      if (token) {
        localStorage.setItem("accessToken", token);
      } else {
        localStorage.removeItem("accessToken");
      }
    }
    set({ accessToken: token });
  },

  setCurrentUser: (user: User | null) => {
    set({ currentUser: user });
  },

  setBootstrapping: (isBootstrapping: boolean) => {
    set({ isBootstrapping });
  },

  logout: () => {
    if (typeof window !== "undefined") {
      localStorage.removeItem("accessToken");
    }
    set({
      accessToken: null,
      currentUser: null,
    });
  },
}));

export const useIsAuthenticated = () =>
  useAuthStore((state) => !!state.accessToken && !!state.currentUser);

const ROLE_PLATFORM_ADMIN: UserRole = "Platform Admin";
const ROLE_DEVOPS_ENGINEER: UserRole = "DevOps Engineer";
const ROLE_DEVELOPER: UserRole = "Developer";
const ROLE_VIEWER: UserRole = "Viewer";

/** The authenticated user's role, exactly as returned by GET /users/me. */
export const useCurrentUserRole = () => useAuthStore((state) => state.currentUser?.role?.trim());

export const useIsPlatformAdmin = () =>
  useAuthStore((state) => state.currentUser?.role?.trim() === ROLE_PLATFORM_ADMIN);

export const useIsDevOpsEngineer = () =>
  useAuthStore((state) => state.currentUser?.role?.trim() === ROLE_DEVOPS_ENGINEER);

export const useIsDeveloper = () =>
  useAuthStore((state) => state.currentUser?.role?.trim() === ROLE_DEVELOPER);

export const useIsViewer = () =>
  useAuthStore((state) => state.currentUser?.role?.trim() === ROLE_VIEWER);

/**
 * A coarse, UI-gating-only mirror of the backend's permission matrix
 * (internal/rbac). It intentionally covers just the handful of capabilities
 * the UI actually branches on, not the full backend permission taxonomy.
 * The backend remains the authoritative enforcement layer — this only
 * decides what controls to show, never what requests are allowed to
 * succeed.
 */
export type Permission =
  | "organization:manage"
  | "invitation:manage"
  | "team:manage"
  | "project:manage"
  | "project-team:manage"
  | "cluster:manage"
  | "application:manage"
  | "deployment:manage"
  | "incident:manage"
  | "incident:comment"
  | "alert:manage"
  | "notification:manage"
  | "slo:manage";

const ROLE_PERMISSIONS: Record<UserRole, ReadonlyArray<Permission>> = {
  "Platform Admin": [
    "organization:manage",
    "invitation:manage",
    "team:manage",
    "project:manage",
    "project-team:manage",
    "cluster:manage",
    "application:manage",
    "deployment:manage",
    "incident:manage",
    "incident:comment",
    "alert:manage",
    "notification:manage",
    "slo:manage",
  ],
  "DevOps Engineer": [
    "cluster:manage",
    "application:manage",
    "deployment:manage",
    "incident:manage",
    "incident:comment",
    "alert:manage",
    "notification:manage",
    "slo:manage",
  ],
  Developer: ["incident:manage", "incident:comment", "alert:manage"],
  Viewer: [],
};

export const useHasPermission = (permission: Permission) =>
  useAuthStore((state) => {
    const role = state.currentUser?.role?.trim() as UserRole | undefined;
    if (!role) return false;
    return ROLE_PERMISSIONS[role]?.includes(permission) ?? false;
  });
