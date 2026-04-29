import { describe, expect, it } from "vitest";

import { mergeProjectIntoList } from "./project-list";
import type { Project } from "@/types";

function buildProject(id: number, name: string): Project {
  return {
    id,
    sort_order: id,
    name,
    branch: "main",
    repo_url: "https://example.com/repo.git",
    description: "",
    webhook_token: `token-${id}`,
    has_deploy_config: false,
    git_auth_type: "none",
    git_username: "",
    has_git_auth: false,
    has_git_password: false,
    has_git_ssh_key: false,
    created_at: "2026-04-29T00:00:00Z",
    updated_at: "2026-04-29T00:00:00Z",
  };
}

describe("mergeProjectIntoList", () => {
  it("puts a cloned project into the current list immediately", () => {
    const source = buildProject(1, "源项目");
    const cloned = buildProject(2, "源项目-副本");

    expect(mergeProjectIntoList([source], cloned)).toEqual([cloned, source]);
  });

  it("replaces an existing project instead of duplicating it", () => {
    const source = buildProject(1, "源项目");
    const staleClone = buildProject(2, "旧副本");
    const cloned = buildProject(2, "新副本");

    expect(mergeProjectIntoList([source, staleClone], cloned)).toEqual([cloned, source]);
  });
});
