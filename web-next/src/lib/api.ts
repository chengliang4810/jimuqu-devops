import { buildApiUrl } from "@/lib/api-base";

/**
 * API 请求封装
 */
async function api<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = typeof window !== "undefined" ? localStorage.getItem("jwt_token") : null;

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(token && { Authorization: `Bearer ${token}` }),
    ...options.headers,
  };

  const response = await fetch(buildApiUrl(endpoint), {
    ...options,
    headers,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: "请求失败" }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  return response.json();
}

/**
 * API 方法简化
 */
export const apiClient = {
  get: <T>(endpoint: string) => api<T>(endpoint),
  post: <T>(endpoint: string, data?: unknown) =>
    api<T>(endpoint, { method: "POST", body: JSON.stringify(data) }),
  put: <T>(endpoint: string, data?: unknown) =>
    api<T>(endpoint, { method: "PUT", body: JSON.stringify(data) }),
  delete: <T>(endpoint: string) => api<T>(endpoint, { method: "DELETE" }),
};

/**
 * 登录
 */
export async function login(username: string, password: string) {
  const response = await apiClient.post<{ token: string }>("/admin/login", {
    username,
    password,
  });
  if (response.token) {
    localStorage.setItem("jwt_token", response.token);
  }
  return response;
}

/**
 * 退出登录
 */
export function logout() {
  localStorage.removeItem("jwt_token");
  window.location.href = "/";
}
