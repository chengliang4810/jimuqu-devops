#!/usr/bin/env node

import fs from "node:fs";
import { createRequire } from "node:module";
import os from "node:os";
import path from "node:path";
import { Command, CommanderError } from "commander";
import { APIClient } from "./client.js";
import { applyConfig, followRun, readApplyFile } from "./apply.js";
import { agentCommands, applyExample, applySchema } from "./agent.js";
import { clearConfig, readConfig, writeConfig } from "./config.js";
import { output, table } from "./output.js";
import { promptPassword, promptText } from "./prompts.js";
import type { GlobalOptions } from "./types.js";

const program = new Command();
const require = createRequire(import.meta.url);
const packageJSON = require("../package.json") as { version?: string };
const cliVersion = packageJSON.version || "0.0.0";

program
  .name("jimuqu-devops")
  .description("Jimuqu DevOps command line client")
  .version(cliVersion)
  .option("--url <server-url>", "server URL")
  .option("--token <token>", "API token or JWT")
  .option("--config <path>", "config file path")
  .option("--json", "print JSON output")
  .option("--no-color", "disable color output");

function globals(): GlobalOptions {
  return program.opts<GlobalOptions>();
}

function client(): APIClient {
  return new APIClient(globals());
}

program.command("login")
  .description("Login with admin username and password")
  .requiredOption("--url <server-url>", "server URL")
  .option("--username <username>", "admin username")
  .action(async (options: { url: string; username?: string }) => {
    const username = options.username || await promptText("Username: ");
    const password = await promptPassword("Password: ");
    const api = new APIClient({ ...globals(), url: options.url });
    const response = await api.request<{ token: string; username: string }>("/admin/login", {
      method: "POST",
      auth: false,
      body: { username, password },
    });
    writeConfig(globals(), { base_url: options.url.replace(/\/+$/, ""), username: response.username, token: response.token });
    output(response, globals(), `Logged in as ${response.username}`);
  });

program.command("logout")
  .description("Remove local login config")
  .action(() => {
    clearConfig(globals());
    output({ ok: true }, globals(), "Logged out.");
  });

program.command("whoami")
  .description("Show current authenticated account")
  .action(async () => {
    const profile = await client().request("/admin/profile");
    output(profile, globals());
  });

program.command("status")
  .description("Check server and authentication status")
  .action(async () => {
    const api = client();
    const health = await api.health();
    const info = await api.request("/system/info");
    const profile = await api.request("/admin/profile");
    output({ health, info, profile }, globals());
  });

program.command("version")
  .description("Show CLI and server version")
  .action(async () => {
    let server: unknown = null;
    try {
      server = await client().request("/system/info");
    } catch {
      server = null;
    }
    output({ cli: cliVersion, server }, globals());
  });

const token = program.command("token").description("Manage API tokens");
token.command("list").action(async () => {
  const rows = await client().request<Array<Record<string, unknown>>>("/admin/tokens");
  table(rows, ["id", "name", "prefix", "expires_at", "revoked_at", "last_used_at"], globals());
});
token.command("create")
  .requiredOption("--name <name>", "token name")
  .option("--expires <duration>", "expiration, for example 30d, 90d, never", "never")
  .action(async (options: { name: string; expires: string }) => {
    const body: Record<string, unknown> = { name: options.name };
    const expiresAt = parseExpires(options.expires);
    if (expiresAt) {
      body.expires_at = expiresAt.toISOString();
    }
    const created = await client().request("/admin/tokens", { method: "POST", body });
    output(created, globals(), JSON.stringify(created, null, 2));
  });
token.command("revoke")
  .argument("<token-id>")
  .action(async (id: string) => {
    await client().request(`/admin/tokens/${id}`, { method: "DELETE" });
    output({ revoked: Number(id) }, globals(), `Revoked token ${id}`);
  });

resourceCommands("host", "/hosts", ["id", "name", "address", "port", "username", "has_password"]);
projectCommands();
deployConfigCommands();
runLikeCommands("run", "/runs");
deployCommands();
notifyCommands();
settingCommands();
apiCommands();
agentCommandsRegistration();

program.command("apply")
  .requiredOption("--file <path>", "apply file path")
  .option("--trigger", "trigger deployment after applying config")
  .option("--watch", "watch deployment logs after trigger")
  .action(async (options: { file: string; trigger?: boolean; watch?: boolean }) => {
    const data = readApplyFile(options.file);
    const result = await applyConfig(data, globals(), Boolean(options.trigger), Boolean(options.watch));
    output(result, globals());
  });

program.parseAsync(process.argv).catch((error: Error) => {
  if (error instanceof CommanderError && error.code === "commander.helpDisplayed") {
    process.exit(0);
  }
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
});

function resourceCommands(name: string, basePath: string, columns: string[]): void {
  const command = program.command(name).description(`Manage ${name}`);
  command.command("list").action(async () => {
    const rows = await client().request<Array<Record<string, unknown>>>(basePath);
    table(rows, columns, globals());
  });
  command.command("get").argument("<id>").action(async (id: string) => {
    output(await client().request(`${basePath}/${id}`), globals());
  });
  command.command("create")
    .option("--file <path>", "JSON file body")
    .allowUnknownOption(true)
    .action(async (options: { file?: string }, command: Command) => {
      output(await client().request(basePath, { method: "POST", body: bodyFromOptions(options.file, command.args) }), globals());
    });
  command.command("update")
    .argument("<id>")
    .option("--file <path>", "JSON file body")
    .allowUnknownOption(true)
    .action(async (id: string, options: { file?: string }, command: Command) => {
      output(await client().request(`${basePath}/${id}`, { method: "PUT", body: bodyFromOptions(options.file, command.args) }), globals());
    });
  command.command("delete").argument("<id>").action(async (id: string) => {
    await client().request(`${basePath}/${id}`, { method: "DELETE" });
    output({ deleted: Number(id) }, globals(), `Deleted ${name} ${id}`);
  });
}

function projectCommands(): void {
  const command = program.command("project").description("Manage projects");
  command.command("list").action(async () => {
    const rows = await client().request<Array<Record<string, unknown>>>("/projects");
    table(rows, ["id", "name", "repo_url", "branch", "has_deploy_config"], globals());
  });
  command.command("get").argument("<id>").action(async (id: string) => output(await client().request(`/projects/${id}`), globals()));
  command.command("create")
    .option("--file <path>", "JSON file body")
    .allowUnknownOption(true)
    .action(async (options: { file?: string }, command: Command) => {
      output(await client().request("/projects", { method: "POST", body: bodyFromOptions(options.file, command.args) }), globals());
    });
  command.command("update")
    .argument("<id>")
    .option("--file <path>", "JSON file body")
    .allowUnknownOption(true)
    .action(async (id: string, options: { file?: string }, command: Command) => {
      output(await client().request(`/projects/${id}`, { method: "PUT", body: bodyFromOptions(options.file, command.args) }), globals());
    });
  command.command("delete").argument("<id>").action(async (id: string) => {
    await client().request(`/projects/${id}`, { method: "DELETE" });
    output({ deleted: Number(id) }, globals(), `Deleted project ${id}`);
  });
  command.command("clone")
    .argument("<id>")
    .option("--file <path>", "JSON file body")
    .allowUnknownOption(true)
    .action(async (id: string, options: { file?: string }, cmd: Command) => {
      output(await client().request(`/projects/${id}/clone`, { method: "POST", body: bodyFromOptions(options.file, cmd.args) }), globals());
    });
}

function deployConfigCommands(): void {
  const command = program.command("deploy-config").description("Manage deploy config");
  command.command("get").argument("<project-id>").action(async (id: string) => output(await client().request(`/projects/${id}/deploy-config`), globals()));
  command.command("set")
    .argument("<project-id>")
    .requiredOption("--file <path>", "JSON or YAML deploy config")
    .action(async (id: string, options: { file: string }) => {
      output(await client().request(`/projects/${id}/deploy-config`, { method: "PUT", body: readStructuredFile(options.file) }), globals());
    });
  command.command("edit")
    .argument("<project-id>")
    .action(async (id: string) => {
      const current = await client().request(`/projects/${id}/deploy-config`);
      const file = path.join(os.tmpdir(), `jimuqu-deploy-config-${id}.json`);
      fs.writeFileSync(file, JSON.stringify(current, null, 2));
      const editor = process.env.EDITOR || process.env.VISUAL;
      if (!editor) {
        throw new Error(`EDITOR is not set. Config written to ${file}`);
      }
      const child = await import("node:child_process");
      child.execFileSync(editor, [file], { stdio: "inherit" });
      output(await client().request(`/projects/${id}/deploy-config`, { method: "PUT", body: readStructuredFile(file) }), globals());
    });
}

function deployCommands(): void {
  const command = program.command("deploy").description("Run deployments");
  command.command("trigger").argument("<project-id>").option("--watch").action(async (id: string, options: { watch?: boolean }) => {
    const run = await client().request(`/projects/${id}/trigger`, { method: "POST" });
    output(run, globals());
    if (options.watch) {
      await followRun(client(), runID(run), globals());
    }
  });
  command.command("logs").argument("<run-id>").option("--follow").action(async (id: string, options: { follow?: boolean }) => {
    if (options.follow) {
      await followRun(client(), Number(id), globals());
      return;
    }
    output(await client().request(`/runs/${id}/log`), globals());
  });
  command.command("cancel").argument("<run-id>").action(async (id: string) => output(await client().request(`/runs/${id}/cancel`, { method: "POST" }), globals()));
}

function runLikeCommands(name: string, basePath: string): void {
  const command = program.command(name).description("Manage runs");
  command.command("list").option("--limit <n>").option("--offset <n>").action(async (options: { limit?: string; offset?: string }) => {
    const params = new URLSearchParams();
    if (options.limit) params.set("limit", options.limit);
    if (options.offset) params.set("offset", options.offset);
    const query = params.toString();
    const rows = await client().request<Array<Record<string, unknown>>>(`${basePath}${query ? `?${query}` : ""}`);
    table(rows, ["id", "project_id", "project_name", "branch", "status", "created_at"], globals());
  });
  command.command("get").argument("<id>").action(async (id: string) => output(await client().request(`${basePath}/${id}`), globals()));
  command.command("logs").argument("<id>").option("--follow").action(async (id: string, options: { follow?: boolean }) => {
    if (options.follow) {
      await followRun(client(), Number(id), globals());
      return;
    }
    output(await client().request(`${basePath}/${id}/log`), globals());
  });
  command.command("cancel").argument("<id>").action(async (id: string) => output(await client().request(`${basePath}/${id}/cancel`, { method: "POST" }), globals()));
}

function notifyCommands(): void {
  const command = program.command("notify").description("Manage notification channels");
  command.command("list").action(async () => {
    const rows = await client().request<Array<Record<string, unknown>>>("/notification-channels");
    table(rows, ["id", "name", "type", "is_default", "remark"], globals());
  });
  command.command("get").argument("<id>").action(async (id: string) => output(await client().request(`/notification-channels/${id}`), globals()));
  command.command("create")
    .option("--file <path>", "JSON file body")
    .allowUnknownOption(true)
    .action(async (options: { file?: string }, cmd: Command) => {
      output(await client().request("/notification-channels", { method: "POST", body: bodyFromOptions(options.file, cmd.args) }), globals());
    });
  command.command("update")
    .argument("<id>")
    .option("--file <path>", "JSON file body")
    .allowUnknownOption(true)
    .action(async (id: string, options: { file?: string }, cmd: Command) => {
      output(await client().request(`/notification-channels/${id}`, { method: "PUT", body: bodyFromOptions(options.file, cmd.args) }), globals());
    });
  command.command("delete").argument("<id>").action(async (id: string) => {
    await client().request(`/notification-channels/${id}`, { method: "DELETE" });
    output({ deleted: Number(id) }, globals(), `Deleted notification channel ${id}`);
  });
  command.command("default").argument("<id>").action(async (id: string) => {
    await client().request(`/notification-channels/${id}/default`, { method: "PUT" });
    output({ default: Number(id) }, globals(), `Set default notification channel ${id}`);
  });
  command.command("test").argument("<id>").option("--title <title>", "message title", "测试通知").option("--content <content>", "message content", "这是一条来自积木区流水线的测试通知。").action(async (id: string, options: { title: string; content: string }) => {
    output(await client().request(`/notification-channels/${id}/test`, { method: "POST", body: { title: options.title, content: options.content } }), globals());
  });
}

function settingCommands(): void {
  const command = program.command("setting").description("Manage settings");
  command.command("list").action(async () => {
    const rows = await client().request<Array<Record<string, unknown>>>("/settings");
    table(rows, ["key", "value"], globals());
  });
  command.command("get").argument("<key>").action(async (key: string) => {
    const rows = await client().request<Array<{ key: string; value: string }>>("/settings");
    const setting = rows.find((item) => item.key === key);
    if (!setting) throw new Error(`setting not found: ${key}`);
    output(setting, globals(), setting.value);
  });
  command.command("set").argument("<key>").argument("<value>").action(async (key: string, value: string) => {
    output(await client().request(`/settings/${key}`, { method: "PUT", body: { value } }), globals());
  });
}

function apiCommands(): void {
  const command = program.command("api").description("Call raw API");
  for (const method of ["get", "post", "put", "delete"] as const) {
    command.command(method).argument("<path>").option("--data <json>").action(async (apiPath: string, options: { data?: string }) => {
      const body = options.data ? JSON.parse(options.data) : undefined;
      output(await client().request(apiPath, { method: method.toUpperCase(), body }), globals());
    });
  }
}

function agentCommandsRegistration(): void {
  const command = program.command("agent").description("Show Agent-oriented command metadata");
  command.action(() => output({ workflow: ["status", "host list", "project list", "apply --file jimuqu-devops.yml --trigger --watch"], commands: agentCommands.commands }, globals()));
  command.command("commands").action(() => output(agentCommands, globals()));
  command.command("examples").action(() => output(applyExample, globals()));
  command.command("schema").action(() => output(applySchema, globals()));
}

function bodyFromOptions(file: string | undefined, args: string[]): unknown {
  if (file) {
    return readStructuredFile(file);
  }
  return objectFromArgs(args);
}

function objectFromArgs(args: string[]): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  const globalOptions = new Set(["url", "token", "config", "json", "no-color", "color"]);
  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];
    if (!arg.startsWith("--")) {
      continue;
    }
    const rawKey = arg.slice(2);
    const value = args[index + 1];
    if (globalOptions.has(rawKey)) {
      if (value !== undefined && !value.startsWith("--") && rawKey !== "json" && rawKey !== "no-color" && rawKey !== "color") {
        index += 1;
      }
      continue;
    }
    const key = rawKey.replace(/-/g, "_");
    if (value === undefined || value.startsWith("--")) {
      result[key] = true;
    } else {
      result[key] = parseValue(value);
      index += 1;
    }
  }
  return result;
}

function readStructuredFile(file: string): unknown {
  const text = fs.readFileSync(path.resolve(file), "utf8");
  if (file.endsWith(".yml") || file.endsWith(".yaml")) {
    return readApplyFile(file);
  }
  return JSON.parse(text);
}

function parseValue(value: string): unknown {
  if (value === "true") return true;
  if (value === "false") return false;
  if (/^-?\d+$/.test(value)) return Number(value);
  if ((value.startsWith("[") && value.endsWith("]")) || (value.startsWith("{") && value.endsWith("}"))) {
    return JSON.parse(value);
  }
  return value;
}

function parseExpires(value: string): Date | undefined {
  if (!value || value === "never") return undefined;
  const match = value.match(/^(\d+)(d|h|m)$/);
  if (!match) {
    throw new Error("--expires must be never or a duration like 30d, 12h, 60m");
  }
  const amount = Number(match[1]);
  const unit = match[2];
  const milliseconds = unit === "d" ? amount * 24 * 60 * 60 * 1000 : unit === "h" ? amount * 60 * 60 * 1000 : amount * 60 * 1000;
  return new Date(Date.now() + milliseconds);
}

function runID(run: unknown): number {
  if (run && typeof run === "object" && "id" in run) {
    return (run as { id: number }).id;
  }
  throw new Error("unable to read run id from response");
}
