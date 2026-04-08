import { buildApiUrl } from "@/lib/api-base";

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

  const response = await fetch(buildApiUrl(path), {
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
    const errorText = await response.text();
    const error = errorText ? JSON.parse(errorText) : { error: "请求失败" };
    throw new Error(error.error || "请求失败");
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();
  if (!text) {
    return undefined as T;
  }

  return JSON.parse(text) as T;
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
  reorder: (ids: number[]) =>
    request<void>("/hosts/reorder", {
      method: "PUT",
      body: JSON.stringify({ ids }),
    }),
  delete: (id: number) =>
    request(`/hosts/${id}`, { method: "DELETE" }),
};

// ==================== 项目 API ====================
export const projectApi = {
  list: () => request<import("@/types").Project[]>("/projects"),
  get: (id: number) => request<import("@/types").ProjectDetail>(`/projects/${id}`),
  create: (data: any) =>
    request<import("@/types").ProjectDetail>("/projects", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (id: number, data: any) =>
    request<import("@/types").ProjectDetail>(`/projects/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  clone: (id: number, data: { name: string; branch: string; description?: string }) =>
    request<import("@/types").ProjectDetail>(`/projects/${id}/clone`, {
      method: "POST",
      body: JSON.stringify(data),
    }),
  reorder: (ids: number[]) =>
    request<void>("/projects/reorder", {
      method: "PUT",
      body: JSON.stringify({ ids }),
    }),
  delete: (id: number) =>
    request(`/projects/${id}`, { method: "DELETE" }),
  getDeployConfig: (id: number) =>
    request<import("@/types").DeployConfig>(`/projects/${id}/deploy-config`),
  upsertDeployConfig: (id: number, data: any) =>
    request<import("@/types").DeployConfig>(`/projects/${id}/deploy-config`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  searchImages: (query: string, limit = 8) =>
    request<import("@/types").ImageSearchResponse>(
      `/images/search?q=${encodeURIComponent(query)}&limit=${limit}`
    ),
  trigger: (id: number) =>
    request<import("@/types").PipelineRun>(`/projects/${id}/trigger`, { method: "POST" }),
};

// ==================== 部署记录 API ====================
function buildRunListPath(limit?: number, offset?: number): string {
  const params = new URLSearchParams();
  if (typeof limit === "number" && limit > 0) {
    params.set("limit", String(limit));
  }
  if (typeof offset === "number" && offset >= 0) {
    params.set("offset", String(offset));
  }

  const query = params.toString();
  return query ? `/runs?${query}` : "/runs";
}

export const runApi = {
  list: (params?: { limit?: number; offset?: number }) =>
    request<import("@/types").PipelineRun[]>(buildRunListPath(params?.limit, params?.offset)),
  get: (id: number) => request<import("@/types").PipelineRun>(`/runs/${id}`),
  getLog: (id: number) => request<import("@/types").PipelineRunLog>(`/runs/${id}/log`),
  interpret: (id: number) =>
    request<import("@/types").AIInterpretationResponse>(`/runs/${id}/interpret`, { method: "POST" }),
  cancel: (id: number) => request<import("@/types").PipelineRun>(`/runs/${id}/cancel`, { method: "POST" }),
  clear: () => request<{ cleared: number }>("/runs", { method: "DELETE" }),
};

function buildImageSearchPath(query: string, limit?: number): string {
  const trimmedQuery = query.trim();

  if (!trimmedQuery) {
    throw new Error("Image search query must not be empty");
  }

  const params = new URLSearchParams();
  params.set("q", trimmedQuery);

  if (typeof limit === "number" && limit > 0) {
    params.set("limit", String(limit));
  }

  return `/images/search?${params.toString()}`;
}

export const imageApi = {
  search: (query: string, limit?: number) =>
    request<import("@/types").ImageSearchResponse>(buildImageSearchPath(query, limit)),
};

// ==================== 通知渠道 API ====================
export const notifyApi = {
  list: () => request<import("@/types").NotifyChannel[]>("/notification-channels"),
  get: (id: number) => request<import("@/types").NotifyChannelDetail>(`/notification-channels/${id}`),
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
  reorder: (ids: number[]) =>
    request<void>("/notification-channels/reorder", {
      method: "PUT",
      body: JSON.stringify({ ids }),
    }),
  delete: (id: number) =>
    request(`/notification-channels/${id}`, { method: "DELETE" }),
  setDefault: (id: number) =>
    request(`/notification-channels/${id}/default`, { method: "PUT" }),
  test: (id: number, title = "测试通知", content = "这是一条来自积木区流水线的测试通知。") =>
    request<{ status: string }>(`/notification-channels/${id}/test`, {
      method: "POST",
      body: JSON.stringify({ title, content }),
    }),
};

// ==================== 统计 API ====================
export const statsApi = {
  get: () => request<import("@/types").Stats>("/stats"),
};

export const homeApi = {
  getDashboard: () => request<import("@/types").HomeDashboard>("/dashboard/home"),
};

export const settingApi = {
  list: () => request<import("@/types").Setting[]>("/settings"),
  update: (key: import("@/types").SettingKey, value: string) =>
    request<import("@/types").Setting>(`/settings/${key}`, {
      method: "PUT",
      body: JSON.stringify({ value }),
    }),
  getAI: () => request<import("@/types").AISettings>("/settings/ai"),
  getAIStatus: () => request<import("@/types").AISettingsStatus>("/settings/ai/status"),
  updateAI: (settings: import("@/types").AISettings) =>
    request<import("@/types").AISettings>("/settings/ai", {
      method: "PUT",
      body: JSON.stringify(settings),
    }),
  exportBackup: () => request<any>("/settings/backup"),
  importBackup: (data: any) =>
    request<import("@/types").BackupRestoreResult>("/settings/restore", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  getProfile: () => request<import("@/types").AccountProfile>("/admin/profile"),
  changeUsername: (newUsername: string) =>
    request<void>("/admin/username", {
      method: "PUT",
      body: JSON.stringify({ new_username: newUsername }),
    }),
  changePassword: (oldPassword: string, newPassword: string) =>
    request<void>("/admin/password", {
      method: "PUT",
      body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
    }),
  getSystemInfo: () => request<import("@/types").SystemInfo>("/system/info"),
  getLatestRelease: () => request<import("@/types").ReleaseInfo>("/update"),
  getUpdateStatus: () => request<import("@/types").UpdateStatus>("/update/now-version"),
  applyUpdate: () => request<import("@/types").UpdateResult>("/update", { method: "POST" }),
};
