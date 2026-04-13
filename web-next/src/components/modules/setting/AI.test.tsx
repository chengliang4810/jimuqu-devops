// @vitest-environment jsdom

import "@testing-library/jest-dom/vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import React from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { SettingAI } from "./AI";
import type { AISettings } from "@/types";

function buildSettings(overrides: Partial<AISettings> = {}): AISettings {
  return {
    enabled: false,
    protocol: "openai",
    base_url: "",
    api_key: "",
    model: "",
    ...overrides,
  };
}

describe("SettingAI", () => {
  afterEach(() => {
    cleanup();
  });

  it("toggles API key visibility", () => {
    render(<SettingAI settings={buildSettings({ api_key: "plain-key" })} onSave={vi.fn()} />);

    const apiKeyInput = screen.getByLabelText("API Key") as HTMLInputElement;
    expect(apiKeyInput.type).toBe("password");

    fireEvent.click(screen.getByRole("button", { name: "切换 API Key 显示状态" }));
    expect(apiKeyInput.type).toBe("text");
  });

  it("disables save when enabled but required fields are empty", () => {
    render(<SettingAI settings={buildSettings({ enabled: true })} onSave={vi.fn()} />);

    expect(screen.getByRole("button", { name: "保存 AI 配置" })).toBeDisabled();
  });

  it("submits the full settings payload", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<SettingAI settings={buildSettings()} onSave={onSave} />);

    fireEvent.click(screen.getByRole("switch", { name: "启用 AI 解读" }));
    fireEvent.change(screen.getByLabelText("Base URL"), { target: { value: "https://example.com/v1" } });
    fireEvent.change(screen.getByLabelText("API Key"), { target: { value: "plain-key" } });
    fireEvent.change(screen.getByLabelText("模型"), { target: { value: "gpt-4.1-mini" } });
    fireEvent.click(screen.getByRole("button", { name: "保存 AI 配置" }));

    await waitFor(() => {
      expect(onSave).toHaveBeenCalledWith({
        enabled: true,
        protocol: "openai",
        base_url: "https://example.com/v1",
        api_key: "plain-key",
        model: "gpt-4.1-mini",
      });
    });
  });

  it("submits the selected protocol payload", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<SettingAI settings={buildSettings()} onSave={onSave} />);

    fireEvent.click(screen.getByRole("combobox", { name: "协议" }));
    fireEvent.click(screen.getByText("OpenAI Responses"));
    fireEvent.click(screen.getByRole("switch", { name: "启用 AI 解读" }));
    fireEvent.change(screen.getByLabelText("Base URL"), { target: { value: "https://api.openai.com/v1" } });
    fireEvent.change(screen.getByLabelText("API Key"), { target: { value: "plain-key" } });
    fireEvent.change(screen.getByLabelText("模型"), { target: { value: "gpt-5-mini" } });
    fireEvent.click(screen.getByRole("button", { name: "保存 AI 配置" }));

    await waitFor(() => {
      expect(onSave).toHaveBeenCalledWith({
        enabled: true,
        protocol: "openai_responses",
        base_url: "https://api.openai.com/v1",
        api_key: "plain-key",
        model: "gpt-5-mini",
      });
    });
  });

  it("updates protocol-specific guidance when gemini is selected", async () => {
    render(<SettingAI settings={buildSettings()} onSave={vi.fn()} />);

    fireEvent.click(screen.getByRole("combobox", { name: "协议" }));
    fireEvent.click(screen.getByText("Gemini GenerateContent"));

    await waitFor(() => {
      expect(
        screen.getByText("使用 Gemini GenerateContent 接口，请填写 API 根地址，后端会请求 models/{model}:generateContent。")
      ).toBeInTheDocument();
      expect(screen.getByLabelText("Base URL")).toHaveAttribute(
        "placeholder",
        "https://generativelanguage.googleapis.com/v1beta"
      );
    });
  });
});
