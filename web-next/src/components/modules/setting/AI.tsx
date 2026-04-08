"use client";

import { useEffect, useMemo, useState } from "react";
import { Bot, Eye, EyeOff, KeyRound, Save } from "lucide-react";

import type { AISettings } from "@/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

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
  const [showAPIKey, setShowAPIKey] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setEnabled(settings.enabled);
    setProtocol(settings.protocol);
    setBaseURL(settings.base_url);
    setAPIKey(settings.api_key);
    setModel(settings.model);
  }, [settings]);

  const normalizedSettings = useMemo<AISettings>(
    () => ({
      enabled,
      protocol,
      base_url: baseURL.trim(),
      api_key: apiKey.trim(),
      model: model.trim(),
    }),
    [apiKey, baseURL, enabled, model, protocol]
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
            <SelectItem value="openai">OpenAI 兼容协议</SelectItem>
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
          placeholder="https://api.openai.com/v1"
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
          placeholder="gpt-4.1-mini"
          className="rounded-xl"
        />
      </div>

      <p className="text-xs leading-5 text-muted-foreground">
        关闭时保留已填写配置但不在部署失败详情中展示 AI 解读入口。当前仅支持 OpenAI 兼容的 Chat Completions 接口。
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
