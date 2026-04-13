"use client";

import { useEffect, useMemo, useState } from "react";
import { Bot, Eye, EyeOff, KeyRound, Save } from "lucide-react";

import type { AISettings } from "@/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

const PROTOCOL_OPTIONS: Array<{
  value: AISettings["protocol"];
  label: string;
  baseURLPlaceholder: string;
  modelPlaceholder: string;
  helperText: string;
}> = [
  {
    value: "openai",
    label: "OpenAI 兼容 / Chat Completions",
    baseURLPlaceholder: "https://api.openai.com/v1",
    modelPlaceholder: "gpt-4.1-mini",
    helperText: "兼容 OpenAI Chat Completions 接口，适合 OpenAI 兼容网关或通用兼容服务。",
  },
  {
    value: "openai_responses",
    label: "OpenAI Responses",
    baseURLPlaceholder: "https://api.openai.com/v1",
    modelPlaceholder: "gpt-5-mini",
    helperText: "使用 OpenAI Responses API，请填写包含 /v1 的根地址。",
  },
  {
    value: "anthropic",
    label: "Claude / Anthropic Messages",
    baseURLPlaceholder: "https://api.anthropic.com/v1",
    modelPlaceholder: "claude-sonnet-4-5",
    helperText: "使用 Anthropic Messages API，后端会自动补充所需版本头。",
  },
  {
    value: "gemini",
    label: "Gemini GenerateContent",
    baseURLPlaceholder: "https://generativelanguage.googleapis.com/v1beta",
    modelPlaceholder: "gemini-2.5-flash",
    helperText: "使用 Gemini GenerateContent 接口，请填写 API 根地址，后端会请求 models/{model}:generateContent。",
  },
];

type SettingAIProps = {
  settings: AISettings;
  onSave: (settings: AISettings) => Promise<void>;
};

export function SettingAI({ settings, onSave }: SettingAIProps) {
  const [enabled, setEnabled] = useState(settings.enabled);
  const [protocol, setProtocol] = useState(settings.protocol);
  const [baseURL, setBaseURL] = useState(settings.base_url);
  const [apiKey, setAPIKey] = useState(settings.api_key);
  const [model, setModel] = useState(settings.model);
  const [userAgent, setUserAgent] = useState(settings.user_agent);
  const [showAPIKey, setShowAPIKey] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setEnabled(settings.enabled);
    setProtocol(settings.protocol);
    setBaseURL(settings.base_url);
    setAPIKey(settings.api_key);
    setModel(settings.model);
    setUserAgent(settings.user_agent);
  }, [settings]);

  const normalizedSettings = useMemo<AISettings>(
    () => ({
      enabled,
      protocol,
      base_url: baseURL.trim(),
      api_key: apiKey.trim(),
      model: model.trim(),
      user_agent: userAgent.trim(),
    }),
    [apiKey, baseURL, enabled, model, protocol, userAgent]
  );
  const protocolMeta = useMemo(
    () => PROTOCOL_OPTIONS.find((option) => option.value === protocol) ?? PROTOCOL_OPTIONS[0],
    [protocol]
  );

  const canSave = !enabled || Boolean(normalizedSettings.base_url && normalizedSettings.api_key && normalizedSettings.model);

  const handleSave = async () => {
    if (!canSave || saving) {
      return;
    }

    try {
      setSaving(true);
      await onSave(normalizedSettings);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="rounded-3xl border border-border bg-card p-6 space-y-5">
      <div className="flex items-center justify-between gap-4">
        <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
          <Bot className="h-5 w-5" />
          AI 解读
        </h2>
        <div className="flex items-center gap-3">
          <Label htmlFor="ai-enabled" className="text-sm font-medium">
            启用 AI 解读
          </Label>
          <Switch
            id="ai-enabled"
            checked={enabled}
            aria-label="启用 AI 解读"
            onCheckedChange={setEnabled}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="ai-protocol">协议</Label>
        <Select value={protocol} onValueChange={(value) => setProtocol(value as AISettings["protocol"])}>
          <SelectTrigger id="ai-protocol" aria-label="协议">
            <SelectValue placeholder="选择协议" />
          </SelectTrigger>
          <SelectContent>
            {PROTOCOL_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="ai-base-url">Base URL</Label>
        <Input
          id="ai-base-url"
          aria-label="Base URL"
          value={baseURL}
          onChange={(event) => setBaseURL(event.target.value)}
          placeholder={protocolMeta.baseURLPlaceholder}
          className="rounded-xl"
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="ai-api-key">API Key</Label>
        <div className="relative">
          <Input
            id="ai-api-key"
            aria-label="API Key"
            type={showAPIKey ? "text" : "password"}
            value={apiKey}
            onChange={(event) => setAPIKey(event.target.value)}
            placeholder="sk-..."
            className="rounded-xl pr-20"
          />
          <KeyRound className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <button
            type="button"
            aria-label="切换 API Key 显示状态"
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground"
            onClick={() => setShowAPIKey((current) => !current)}
          >
            {showAPIKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="ai-model">模型</Label>
        <Input
          id="ai-model"
          aria-label="模型"
          value={model}
          onChange={(event) => setModel(event.target.value)}
          placeholder={protocolMeta.modelPlaceholder}
          className="rounded-xl"
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="ai-user-agent">User-Agent（可选）</Label>
        <Input
          id="ai-user-agent"
          aria-label="User-Agent"
          value={userAgent}
          onChange={(event) => setUserAgent(event.target.value)}
          placeholder="Codex Desktop/0.115.0-alpha.11 (Windows 10.0.22621; x86_64) unknown (Codex Desktop; 26.311.21342)"
          className="rounded-xl"
        />
      </div>

      <p className="text-xs leading-5 text-muted-foreground">
        {protocolMeta.helperText}
      </p>

      <p className="text-xs leading-5 text-muted-foreground">
        关闭时保留已填写配置但不在部署失败详情中展示 AI 解读入口。当前只会启用这一套配置。
      </p>

      <Button
        type="button"
        className="w-full rounded-xl"
        disabled={!canSave || saving}
        aria-label="保存 AI 配置"
        onClick={() => void handleSave()}
      >
        <Save className="mr-2 h-4 w-4" />
        {saving ? "保存中..." : "保存 AI 配置"}
      </Button>
    </div>
  );
}
