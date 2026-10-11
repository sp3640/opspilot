import axios from "axios";

/**
 * Stable, machine-readable reasons the backend attaches to auth failures
 * (the `code` field of its error envelope). The UI branches on these rather
 * than on message text.
 */
export const AUTH_ERROR_CODES = {
  emailNotVerified: "email_not_verified",
  emailDeliveryUnavailable: "email_delivery_unavailable",
  verificationTokenInvalid: "verification_token_invalid",
  verificationTokenExpired: "verification_token_expired",
} as const;

interface ApiErrorBody {
  code?: string;
  message?: string;
}

export function getApiErrorCode(error: unknown): string | undefined {
  if (!axios.isAxiosError<ApiErrorBody>(error)) {
    return undefined;
  }
  return error.response?.data?.code;
}

export function getApiErrorStatus(error: unknown): number | undefined {
  return axios.isAxiosError(error) ? error.response?.status : undefined;
}

/**
 * The server's own message when it sent one, else `fallback`. A request that
 * never reached the server (no response) always gets `fallback`, since
 * axios's own text ("Network Error") is not meant for users.
 */
export function getApiErrorMessage(error: unknown, fallback: string): string {
  if (axios.isAxiosError<ApiErrorBody>(error)) {
    return error.response?.data?.message || fallback;
  }
  return fallback;
}
