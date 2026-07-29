import { api } from "@/lib/api";
import type { LoginRequest, LoginResponse, MeResponse } from "@/types/auth";

export const authService = {
  login: async (payload: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>("/auth/login", payload);
    return response.data;
  },

  me: async (): Promise<MeResponse> => {
    const response = await api.get<MeResponse>("/users/me");
    return response.data;
  },
};
