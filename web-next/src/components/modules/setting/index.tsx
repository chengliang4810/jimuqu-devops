"use client";

import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { settingApi } from "@/api/client";
import { PageWrapper } from "@/components/common/PageWrapper";
import type { Setting as SettingItem, SettingKey, SystemInfo } from "@/types";
import { SettingInfo } from "./Info";
import { SettingImage } from "./Image";
import { SettingCache } from "./Cache";
import { SettingSystem } from "./System";
import { SettingAccount } from "./Account";
import { SettingLog } from "./Log";
import { SettingBackup } from "./Backup";
import { SettingAI } from "./AI";
import type { AISettings } from "@/types";

const defaultSettings: Record<SettingKey, string> = {
  docker_mirror_url: "",
  git_docker_image: "alpine/git:latest",
  build_cache_dirs: "/root/.m2\n/root/.gradle/caches\n/root/.npm\n/root/.yarn\n/go/pkg/mod\n/root/.cache",
  public_base_url: "",
  proxy_url: "",
  run_retention_days: "30",
};

const defaultAISettings: AISettings = {
  enabled: false,
  protocol: "openai",
  base_url: "",
  api_key: "",
  model: "",
};

function buildSettingsMap(settings: SettingItem[]) {
  return settings.reduce<Record<SettingKey, string>>(
    (accumulator, setting) => {
      accumulator[setting.key] = setting.value;
      return accumulator;
    },
    { ...defaultSettings }
  );
}

export function Setting() {
  const [settings, setSettings] = useState<Record<SettingKey, string>>(defaultSettings);
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [aiSettings, setAISettings] = useState<AISettings>(defaultAISettings);

  const loadData = async () => {
    try {
      const [settingsList, info, ai] = await Promise.all([
        settingApi.list(),
        settingApi.getSystemInfo(),
        settingApi.getAI(),
      ]);
      setSettings(buildSettingsMap(settingsList));
      setSystemInfo(info);
      setAISettings(ai);
    } catch (error) {
      console.error(error);
      toast.error("加载设置失败");
    }
  };

  useEffect(() => {
    void loadData();
  }, []);

  const settingMap = useMemo(() => settings, [settings]);

  const handleSaveSetting = async (key: SettingKey, value: string) => {
    try {
      const saved = await settingApi.update(key, value);
      setSettings((current) => ({ ...current, [saved.key]: saved.value }));
      toast.success("设置已保存");
    } catch (error: any) {
      toast.error(error.message || "设置保存失败");
      throw error;
    }
  };

  const handleSaveAISettings = async (nextSettings: AISettings) => {
    try {
      const saved = await settingApi.updateAI(nextSettings);
      setAISettings(saved);
      toast.success("AI 配置已保存");
    } catch (error: any) {
      toast.error(error.message || "AI 配置保存失败");
      throw error;
    }
  };

  return (
    <div className="h-full min-h-0 overflow-y-auto overscroll-contain rounded-t-3xl">
      <PageWrapper className="columns-1 gap-4 pb-24 md:columns-2 md:pb-4 *:mb-4 *:break-inside-avoid">
        <SettingInfo key="setting-info" systemInfo={systemInfo} />
        <SettingAccount key="setting-account" />
        <SettingBackup key="setting-backup" />
        <SettingAI key="setting-ai" settings={aiSettings} onSave={handleSaveAISettings} />
        <SettingSystem key="setting-system" settings={settingMap} onSave={handleSaveSetting} />
        <SettingImage key="setting-image" settings={settingMap} onSave={handleSaveSetting} />
        <SettingCache key="setting-cache" settings={settingMap} onSave={handleSaveSetting} />
        <SettingLog key="setting-log" settings={settingMap} onSave={handleSaveSetting} />
      </PageWrapper>
    </div>
  );
}
