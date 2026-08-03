import { api } from "@/lib/api";
import type {
  LoginRequest,
  LoginResponse,
  MeResponse,
  RegisterRequest,
  RegisterResponse,
} from "@/types/auth";

export const authService = {
  register: async (payload: RegisterRequest): Promise<RegisterResponse> => {
    const response = await api.post<RegisterResponse>("/auth/register", payload);
    return response.data;
  },

  login: async (payload: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>("/auth/login", payload);
    return response.data;
  },

  me: async (): Promise<MeResponse> => {
    const response = await api.get<MeResponse>("/users/me");
    return response.data;
  },
};
