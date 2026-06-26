// @vitest-environment jsdom

import "@testing-library/jest-dom/vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import React from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { projectApi, hostApi, notifyApi } from "@/api/client";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useNavStore } from "@/stores";
import type { PipelineRun, Project } from "@/types";
import { Projects } from "./index";

vi.mock("@/api/client", () => ({
  projectApi: {
    list: vi.fn(),
    trigger: vi.fn(),
    searchImages: vi.fn(),
  },
  hostApi: {
    list: vi.fn(),
  },
  notifyApi: {
    list: vi.fn(),
  },
}));

vi.mock("sonner", () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

function buildProject(): Project {
  return {
    id: 1,
    sort_order: 1,
    name: "门户服务",
    branch: "main",
    repo_url: "https://example.com/repo.git",
    description: "",
    webhook_token: "webhook-token",
    has_deploy_config: true,
    git_auth_type: "none",
    git_username: "",
    has_git_auth: false,
    has_git_password: false,
    has_git_ssh_key: false,
    created_at: "2026-06-26T00:00:00Z",
    updated_at: "2026-06-26T00:00:00Z",
  };
}

function buildRun(): PipelineRun {
  return {
    id: 88,
    project_id: 1,
    project_name: "门户服务",
    branch: "main",
    status: "queued",
    commit_message: "",
    started_at: "2026-06-26T00:00:00Z",
    finished_at: null,
  };
}

describe("Projects", () => {
  beforeEach(() => {
    useNavStore.setState({ activeView: "projects", pendingRunId: null });
    vi.mocked(projectApi.list).mockResolvedValue([buildProject()]);
    vi.mocked(projectApi.trigger).mockResolvedValue(buildRun());
    vi.mocked(hostApi.list).mockResolvedValue([]);
    vi.mocked(notifyApi.list).mockResolvedValue([]);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("stays on the project page after triggering and opens details only when requested", async () => {
    render(
      <TooltipProvider>
        <Projects />
      </TooltipProvider>
    );

    fireEvent.click(await screen.findByRole("button", { name: "触发部署" }));

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "部署已触发" })).toBeInTheDocument();
    });
    expect(useNavStore.getState().activeView).toBe("projects");
    expect(useNavStore.getState().pendingRunId).toBeNull();

    fireEvent.click(screen.getByRole("button", { name: "查看详情" }));

    expect(useNavStore.getState().activeView).toBe("logs");
    expect(useNavStore.getState().pendingRunId).toBe(88);
  });
});
