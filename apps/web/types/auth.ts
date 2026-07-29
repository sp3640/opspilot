export interface LoginRequest {
  email: string;
  password: string;
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
}

export interface MeResponse {
  success: boolean;
  message: string;
  data: User;
}
