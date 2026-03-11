const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || "";

// 获取 Token
function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("jwt_token");
}

// 设置 Token
export function setToken(token: string): void {
  if (typeof window === "undefined") return;
  localStorage.setItem("jwt_token", token);
}

// 清除 Token
export function clearToken(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem("jwt_token");
}

// 检查是否已认证
export function isAuthenticated(): boolean {
  return !!getToken();
}

// 请求封装
async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}/api/v1${path}`, {
    ...options,
    headers,
    redirect: "follow",
  });

  if (!response.ok) {
    if (response.status === 401) {
      clearToken();
      if (typeof window !== "undefined") {
        window.location.href = "/";
      }
    }
    const error = await response.json().catch(() => ({ error: "请求失败" }));
    throw new Error(error.error || "请求失败");
  }

  return response.json();
}

// ==================== 认证 API ====================
export const authApi = {
  login: (username: string, password: string) =>
    request<{ token: string; username: string }>("/admin/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }),
};

// ==================== 主机 API ====================
export const hostApi = {
  list: () => request<import("@/types").Host[]>("/hosts"),
  get: (id: number) => request<import("@/types").Host>(`/hosts/${id}`),
  create: (data: Partial<import("@/types").Host>) =>
    request<import("@/types").Host>("/hosts", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (id: number, data: Partial<import("@/types").Host>) =>
    request<import("@/types").Host>(`/hosts/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  delete: (id: number) =>
    request(`/hosts/${id}`, { method: "DELETE" }),
};

// ==================== 项目 API ====================
export const projectApi = {
  list: () => request<import("@/types").Project[]>("/projects"),
  get: (id: number) => request<import("@/types").Project>(`/projects/${id}`),
  create: (data: any) =>
    request<import("@/types").Project>("/projects", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (id: number, data: any) =>
    request<import("@/types").Project>(`/projects/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  delete: (id: number) =>
    request(`/projects/${id}`, { method: "DELETE" }),
  trigger: (id: number) =>
    request<{ message: string }>(`/projects/${id}/trigger`, { method: "POST" }),
};

// ==================== 部署记录 API ====================
export const runApi = {
  list: () => request<import("@/types").PipelineRun[]>("/runs"),
  get: (id: number) => request<import("@/types").PipelineRun>(`/runs/${id}`),
};

// ==================== 通知渠道 API ====================
export const notifyApi = {
  list: () => request<import("@/types").NotifyChannel[]>("/notification-channels"),
  get: (id: number) => request<import("@/types").NotifyChannel>(`/notification-channels/${id}`),
  create: (data: Partial<import("@/types").NotifyChannel>) =>
    request<import("@/types").NotifyChannel>("/notification-channels", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (id: number, data: Partial<import("@/types").NotifyChannel>) =>
    request<import("@/types").NotifyChannel>(`/notification-channels/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  delete: (id: number) =>
    request(`/notification-channels/${id}`, { method: "DELETE" }),
  test: (id: number, message?: string) =>
    request<{ message: string }>(`/notification-channels/${id}/test`, {
      method: "POST",
      body: JSON.stringify({ message: message || "测试通知" }),
    }),
};

// ==================== 统计 API ====================
export const statsApi = {
  get: () => request<import("@/types").Stats>("/stats"),
};
