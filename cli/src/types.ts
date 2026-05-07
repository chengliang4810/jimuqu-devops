export interface GlobalOptions {
  url?: string;
  token?: string;
  config?: string;
  json?: boolean;
  noColor?: boolean;
}

export interface CLIConfig {
  base_url?: string;
  username?: string;
  token?: string;
}

export interface RequestOptions {
  method?: string;
  body?: unknown;
  auth?: boolean;
  rawPath?: boolean;
}

export interface APIToken {
  id: number;
  name: string;
  prefix: string;
  expires_at?: string | null;
  revoked_at?: string | null;
  last_used_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ApplyFile {
  host?: Record<string, unknown>;
  project?: Record<string, unknown>;
  deploy_config?: Record<string, unknown>;
  notification_channel?: Record<string, unknown>;
}

export interface ApplyResult {
  host?: unknown;
  notification_channel?: unknown;
  project?: unknown;
  deploy_config?: unknown;
  run?: unknown;
}
