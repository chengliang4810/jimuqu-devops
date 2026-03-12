const configuredApiBase = process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || "";

function normalizeApiOrigin(value: string): string {
  return value.replace(/\/+$/, "").replace(/\/api\/v1$/, "");
}

export function getApiOrigin(): string {
  if (configuredApiBase) {
    return normalizeApiOrigin(configuredApiBase);
  }

  if (typeof window !== "undefined") {
    const { protocol, hostname, host, port } = window.location;

    if (port === "3000") {
      return `${protocol}//${hostname}:18080`;
    }

    return `${protocol}//${host}`;
  }

  if (process.env.NODE_ENV === "development") {
    return "http://127.0.0.1:18080";
  }

  return "";
}

export function buildApiUrl(path: string): string {
  return `${getApiOrigin()}/api/v1${path}`;
}
