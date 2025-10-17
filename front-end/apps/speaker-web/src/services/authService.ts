import { httpClient } from "./httpClient";

export interface LoginPayload {
  username: string;
  password: string;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

/**
 * 调用后台登录接口，获取管理员访问令牌。
 */
export async function login(payload: LoginPayload) {
  const data = await httpClient.post<AuthResponse>("/api/v1/auth/login", payload);
  return data;
}

/**
 * 调用刷新接口，返回新的访问令牌。
 */
export async function refresh(refreshToken: string) {
  const data = await httpClient.post<AuthResponse>("/api/v1/auth/refresh", {
    refreshToken
  });
  return data;
}
