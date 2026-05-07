import { resolveBaseURL, resolveToken } from "./config.js";
import type { GlobalOptions, RequestOptions } from "./types.js";

export class APIError extends Error {
  status: number;
  body: string;

  constructor(status: number, body: string) {
    super(body || `HTTP ${status}`);
    this.status = status;
    this.body = body;
  }
}

export class APIClient {
  private global: GlobalOptions;

  constructor(global: GlobalOptions) {
    this.global = global;
  }

  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const baseURL = resolveBaseURL(this.global);
    const url = options.rawPath ? `${baseURL}${path}` : `${baseURL}/api/v1${path.startsWith("/") ? path : `/${path}`}`;
    const headers: Record<string, string> = {};
    if (options.body !== undefined) {
      headers["Content-Type"] = "application/json";
    }
    if (options.auth !== false) {
      const token = resolveToken(this.global);
      if (!token) {
        throw new Error("token is required, run login or set JIMUQU_DEVOPS_TOKEN");
      }
      headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(url, {
      method: options.method || "GET",
      headers,
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
    });
    const text = await response.text();
    if (!response.ok) {
      let message = text;
      try {
        const parsed = JSON.parse(text) as { error?: string };
        message = parsed.error || text;
      } catch {
        // keep raw body
      }
      throw new APIError(response.status, message);
    }
    if (!text) {
      return undefined as T;
    }
    return JSON.parse(text) as T;
  }

  async health(): Promise<unknown> {
    return this.request("/healthz", { auth: false, rawPath: true });
  }
}
