export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
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

export interface User {
  id: number;
  name: string;
  email: string;
  role?: string;
}

export interface MeResponse {
  success: boolean;
  message: string;
  data: User;
}
