// @vitest-environment jsdom

import "@testing-library/jest-dom/vitest";
import { cleanup, render } from "@testing-library/react";
import React from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { runApi } from "@/api/client";
import { useNavStore } from "@/stores";
import type { PipelineRun } from "@/types";
import { Logs } from "./index";

vi.mock("@/api/client", () => ({
  runApi: {
    list: vi.fn(),
    get: vi.fn(),
    getLog: vi.fn(),
    interpret: vi.fn(),
    cancel: vi.fn(),
  },
  settingApi: {
    getAIStatus: vi.fn(),
  },
}));

vi.mock("sonner", () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

function buildRun(id: number): PipelineRun {
  return {
    id,
    project_id: id,
    project_name: `项目 ${id}`,
    branch: "main",
    status: "queued",
    commit_message: "",
    started_at: "2026-06-26T00:00:00Z",
    finished_at: null,
  };
}

describe("Logs", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    useNavStore.setState({ activeView: "logs", pendingRunId: null });
    vi.mocked(runApi.list).mockResolvedValue([buildRun(1)]);
  });

  afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it("refreshes deployment records on a timer", async () => {
    render(<Logs />);

    expect(runApi.list).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(10_000);

    expect(runApi.list).toHaveBeenCalledTimes(2);
  });
});
