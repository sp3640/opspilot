import { create } from "zustand";
import type { User } from "@/types/auth";

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
