import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import type { CLIConfig, GlobalOptions } from "./types.js";

export function defaultConfigPath(): string {
  return path.join(os.homedir(), ".jimuqu-devops", "config.json");
}

export function resolveConfigPath(options: GlobalOptions): string {
  return options.config ? path.resolve(options.config) : defaultConfigPath();
}

export function readConfig(options: GlobalOptions): CLIConfig {
  const configPath = resolveConfigPath(options);
  if (!fs.existsSync(configPath)) {
    return {};
  }
  const text = fs.readFileSync(configPath, "utf8");
  if (!text.trim()) {
    return {};
  }
  return JSON.parse(text) as CLIConfig;
}

export function writeConfig(options: GlobalOptions, config: CLIConfig): void {
  const configPath = resolveConfigPath(options);
  fs.mkdirSync(path.dirname(configPath), { recursive: true });
  fs.writeFileSync(configPath, JSON.stringify(config, null, 2), { mode: 0o600 });
}

export function clearConfig(options: GlobalOptions): void {
  const configPath = resolveConfigPath(options);
  if (fs.existsSync(configPath)) {
    fs.rmSync(configPath);
  }
}

export function resolveBaseURL(options: GlobalOptions): string {
  const config = readConfig(options);
  const value = options.url || process.env.JIMUQU_DEVOPS_URL || config.base_url;
  if (!value) {
    throw new Error("server url is required, run login --url <server-url> or set JIMUQU_DEVOPS_URL");
  }
  return value.replace(/\/+$/, "");
}

export function resolveToken(options: GlobalOptions): string | undefined {
  const config = readConfig(options);
  return options.token || process.env.JIMUQU_DEVOPS_TOKEN || config.token;
}
