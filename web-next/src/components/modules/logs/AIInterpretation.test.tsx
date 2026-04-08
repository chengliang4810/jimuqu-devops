// @vitest-environment jsdom

import "@testing-library/jest-dom/vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import React from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { FailedRunAIInterpretation } from "./AIInterpretation";

describe("FailedRunAIInterpretation", () => {
  afterEach(() => {
    cleanup();
  });

  it("renders nothing when AI is disabled", () => {
    const { container } = render(
      <FailedRunAIInterpretation
        runId={1}
        runStatus="failed"
        enabled={false}
        interpret={vi.fn()}
      />
    );

    expect(container).toBeEmptyDOMElement();
  });

  it("renders nothing for non-failed runs", () => {
    const { container } = render(
      <FailedRunAIInterpretation
        runId={1}
        runStatus="success"
        enabled
        interpret={vi.fn()}
      />
    );

    expect(container).toBeEmptyDOMElement();
  });

  it("shows loading and successful interpretation content", async () => {
    const interpret = vi.fn().mockResolvedValue({
      run_id: 1,
      protocol: "openai",
      model: "gpt-test",
      content: "失败摘要\n构建失败\n\n可能原因\n命令不存在\n\n建议操作\n补充脚本后重试",
      log_truncated: true,
    });

    render(
      <FailedRunAIInterpretation
        runId={1}
        runStatus="failed"
        enabled
        interpret={interpret}
      />
    );

    fireEvent.click(screen.getByRole("button", { name: "AI 一键解读" }));
    expect(screen.getByText("AI 正在解读这次失败记录...")).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/失败摘要/)).toBeInTheDocument();
    });
    expect(screen.getByText(/已基于截断日志解读/)).toBeInTheDocument();
  });

  it("shows an error state when interpretation fails", async () => {
    const interpret = vi.fn().mockRejectedValue(new Error("interpret failed"));

    render(
      <FailedRunAIInterpretation
        runId={1}
        runStatus="failed"
        enabled
        interpret={interpret}
      />
    );

    fireEvent.click(screen.getByRole("button", { name: "AI 一键解读" }));

    await waitFor(() => {
      expect(screen.getByText("interpret failed")).toBeInTheDocument();
    });
  });
});
