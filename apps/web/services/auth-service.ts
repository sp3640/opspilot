import { api } from "@/lib/api";
import type {
  LoginRequest,
  LoginResponse,
  MeResponse,
  RegisterRequest,
  RegisterResponse,
  ResendVerificationRequest,
  ResendVerificationResponse,
  VerifyEmailRequest,
  VerifyEmailResponse,
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

  /**
   * Redeems the token from an emailed verification link. Sent in the body
   * (never the URL) so it does not end up in request logs.
   */
  verifyEmail: async (payload: VerifyEmailRequest): Promise<VerifyEmailResponse> => {
    const response = await api.post<VerifyEmailResponse>("/auth/verify-email", payload);
    return response.data;
  },

  /**
   * Asks for a new verification email. The backend answers identically for
   * unknown, verified and unverified addresses, so a success only means
   * "the request was accepted".
   */
  resendVerification: async (payload: ResendVerificationRequest): Promise<ResendVerificationResponse> => {
    const response = await api.post<ResendVerificationResponse>("/auth/resend-verification", payload);
    return response.data;
  },

  me: async (): Promise<MeResponse> => {
    const response = await api.get<MeResponse>("/users/me");
    return response.data;
  },

  /**
   * Mints a fresh access token reflecting the caller's CURRENT database
   * role/organization. Used right after an action that can change them
   * (currently: invitation acceptance) so the session's authorization
   * doesn't stay stale until the next full login.
   */
  reissue: async (): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>("/auth/reissue");
    return response.data;
  },
};
