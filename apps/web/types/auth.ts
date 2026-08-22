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
