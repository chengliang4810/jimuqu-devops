import fs from "node:fs";
import path from "node:path";
import { parse as parseYAML } from "yaml";
import { APIClient } from "./client.js";
import { jsonLine } from "./output.js";
import type { ApplyFile, ApplyResult, GlobalOptions } from "./types.js";

const envPattern = /^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$/;

export function readApplyFile(filePath: string): ApplyFile {
  const resolved = path.resolve(filePath);
  const text = fs.readFileSync(resolved, "utf8");
  const ext = path.extname(resolved).toLowerCase();
  const parsed = ext === ".yaml" || ext === ".yml" ? parseYAML(text) : JSON.parse(text);
  return expandEnv(parsed) as ApplyFile;
}

export async function applyConfig(file: ApplyFile, global: GlobalOptions, trigger: boolean, watch: boolean): Promise<ApplyResult> {
  const client = new APIClient(global);
  const result: ApplyResult = {};

  if (file.host) {
    result.host = await upsertByName(client, "/hosts", file.host);
  }

  if (file.notification_channel) {
    result.notification_channel = await upsertByName(client, "/notification-channels", file.notification_channel);
  }

  if (file.project) {
    const projectBody = { ...file.project };
    const existing = await findProject(client, String(projectBody.repo_url || ""), String(projectBody.branch || ""));
    let deployConfig = file.deploy_config ? withLinkedIDs(file.deploy_config, result) : undefined;
    if (existing) {
      if (!deployConfig) {
        const detail = await client.request(`/projects/${existing.id}`);
        deployConfig = deployConfigFromDetail(detail);
      }
      if (!deployConfig) {
        throw new Error("existing project has no deploy_config; include deploy_config in apply file");
      }
      projectBody.deploy_config = deployConfig;
      result.project = await client.request(`/projects/${existing.id}`, { method: "PUT", body: projectBody });
    } else {
      if (!deployConfig) {
        throw new Error("project creation requires deploy_config in apply file");
      }
      projectBody.deploy_config = deployConfig;
      result.project = await client.request("/projects", { method: "POST", body: projectBody });
    }
    result.deploy_config = deployConfigFromDetail(result.project);
  }

  if (file.deploy_config && !file.project) {
    throw new Error("deploy_config requires project in apply file");
  }

  if (trigger) {
    if (!result.project) {
      throw new Error("apply --trigger requires project in apply file");
    }
    const projectID = projectIDFromDetail(result.project);
    const run = await client.request(`/projects/${projectID}/trigger`, { method: "POST" });
    result.run = run;
    if (watch) {
      await followRun(client, runIDFromRun(run), global);
    }
  }

  return result;
}

export async function followRun(client: APIClient, runID: number, options: GlobalOptions = {}): Promise<void> {
  let lastLength = 0;
  let lastStatus = "";
  for (;;) {
    const log = await client.request<{ log_text: string }>(`/runs/${runID}/log`);
    const text = log.log_text || "";
    if (text.length > lastLength) {
      const chunk = text.slice(lastLength);
      lastLength = text.length;
      if (options.json) {
        jsonLine({ event: "log", run_id: runID, text: chunk });
      } else {
        process.stdout.write(chunk);
      }
    }
    const run = await client.request<{ status: string }>(`/runs/${runID}`);
    if (options.json && run.status !== lastStatus) {
      lastStatus = run.status;
      jsonLine({ event: "status", run_id: runID, status: run.status });
    }
    if (run.status === "success" || run.status === "failed") {
      if (options.json) {
        jsonLine({ event: "complete", run_id: runID, status: run.status });
      } else {
        if (!text.endsWith("\n")) {
          process.stdout.write("\n");
        }
        process.stdout.write(`Run ${runID} ${run.status}\n`);
      }
      if (run.status === "failed") {
        process.exitCode = 1;
      }
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }
}

async function upsertByName(client: APIClient, basePath: string, body: Record<string, unknown>): Promise<unknown> {
  const name = String(body.name || "");
  if (!name) {
    throw new Error(`${basePath} item name is required`);
  }
  const items = await client.request<Array<{ id: number; name: string }>>(basePath);
  const existing = items.find((item) => item.name === name);
  if (existing) {
    return client.request(`${basePath}/${existing.id}`, { method: "PUT", body });
  }
  return client.request(basePath, { method: "POST", body });
}

async function findProject(client: APIClient, repoURL: string, branch: string): Promise<{ id: number } | undefined> {
  const projects = await client.request<Array<{ id: number; repo_url: string; branch: string }>>("/projects");
  return projects.find((project) => project.repo_url === repoURL && project.branch === branch);
}

function withLinkedIDs(deployConfig: Record<string, unknown>, result: ApplyResult): Record<string, unknown> {
  const body = { ...deployConfig };
  if (!body.host_id && result.host && typeof result.host === "object" && "id" in result.host) {
    body.host_id = (result.host as { id: number }).id;
  }
  if (!body.notification_channel_id && result.notification_channel && typeof result.notification_channel === "object" && "id" in result.notification_channel) {
    body.notification_channel_id = (result.notification_channel as { id: number }).id;
  }
  return body;
}

function projectIDFromDetail(project: unknown): number {
  if (project && typeof project === "object" && "project" in project) {
    return (project as { project: { id: number } }).project.id;
  }
  if (project && typeof project === "object" && "id" in project) {
    return (project as { id: number }).id;
  }
  throw new Error("unable to read project id from response");
}

function deployConfigFromDetail(project: unknown): Record<string, unknown> | undefined {
  if (project && typeof project === "object" && "deploy_config" in project) {
    const deployConfig = (project as { deploy_config?: Record<string, unknown> | null }).deploy_config;
    return deployConfig || undefined;
  }
  if (project && typeof project === "object" && "project" in project) {
    const deployConfig = (project as { deploy_config?: Record<string, unknown> | null }).deploy_config;
    return deployConfig || undefined;
  }
  return undefined;
}

function runIDFromRun(run: unknown): number {
  if (run && typeof run === "object" && "id" in run) {
    return (run as { id: number }).id;
  }
  throw new Error("unable to read run id from response");
}

function expandEnv(value: unknown): unknown {
  if (typeof value === "string") {
    const match = value.match(envPattern);
    if (!match) {
      return value;
    }
    const envValue = process.env[match[1]];
    if (envValue === undefined) {
      throw new Error(`environment variable ${match[1]} is not set`);
    }
    return envValue;
  }
  if (Array.isArray(value)) {
    return value.map(expandEnv);
  }
  if (value && typeof value === "object") {
    return Object.fromEntries(Object.entries(value).map(([key, item]) => [key, expandEnv(item)]));
  }
  return value;
}
