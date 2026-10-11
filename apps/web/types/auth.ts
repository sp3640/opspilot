export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
  /**
   * Names the new workspace created for an uninvited registration. Ignored
   * by the backend when the email matches a pending invitation, since that
   * flow joins the invitation's existing organization instead.
   */
  organizationName?: string;
}

export interface RegisterResponse {
  success: boolean;
  message: string;
  data: null;
}

export interface ResendVerificationRequest {
  email: string;
}

export interface ResendVerificationResponse {
  success: boolean;
  message: string;
  data: null;
}

export interface VerifyEmailRequest {
  token: string;
}

/** "verified": this request activated the account. "already_verified": it was already active. */
export type VerifyEmailStatus = "verified" | "already_verified";

export interface VerifyEmailResponse {
  success: boolean;
  message: string;
  data: { status: VerifyEmailStatus };
}

export interface LoginResponse {
  success: boolean;
  message: string;
  data: {
    access_token: string;
    token_type: string;
  };
}

export type UserRole = "Platform Admin" | "DevOps Engineer" | "Developer" | "Viewer";

export interface User {
  id: number;
  name: string;
  email: string;
  role?: UserRole;
}

export interface MeResponse {
  success: boolean;
  message: string;
  data: User;
}
