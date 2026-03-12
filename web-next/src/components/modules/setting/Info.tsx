"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import packageJson from "../../../../package.json";
import { Button } from "@/components/ui/button";
import { settingApi } from "@/api/client";
import { toast } from "sonner";
import { Download, ExternalLink, Github, Info, RefreshCw, Tag } from "lucide-react";
import type { ReleaseInfo, SystemInfo, UpdateStatus } from "@/types";

type SettingInfoProps = {
  systemInfo: SystemInfo | null;
};

function formatReleaseTime(value?: string) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("zh-CN", { hour12: false });
}

function normalizeVersion(value?: string) {
  if (!value) return "-";
  return value.startsWith("v") ? value.slice(1) : value;
}

export function SettingInfo({ systemInfo }: SettingInfoProps) {
  const [latestRelease, setLatestRelease] = useState<ReleaseInfo | null>(null);
  const [updateStatus, setUpdateStatus] = useState<UpdateStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [updating, setUpdating] = useState(false);

  const loadUpdateInfo = async (showToast = false) => {
    setLoading(true);
    try {
      const [release, status] = await Promise.all([
        settingApi.getLatestRelease(),
        settingApi.getUpdateStatus(),
      ]);
      setLatestRelease(release);
      setUpdateStatus(status);
      if (showToast) {
        toast.success("已检查更新");
      }
    } catch (error: any) {
      if (showToast) {
        toast.error(error.message || "检查更新失败");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadUpdateInfo();
  }, []);

  const canApplyUpdate = useMemo(() => {
    return !!updateStatus?.has_update && !updating;
  }, [updateStatus?.has_update, updating]);

  const handleApplyUpdate = async () => {
    if (!canApplyUpdate) return;
    setUpdating(true);
    try {
      const result = await settingApi.applyUpdate();
      toast.success(result.message || "更新已开始，服务即将重启");
    } catch (error: any) {
      toast.error(error.message || "在线更新失败");
      setUpdating(false);
    }
  };

  return (
    <div className="rounded-3xl border border-border bg-card p-6 space-y-5">
      <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
        <Info className="h-5 w-5" />
        版本信息
      </h2>

      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Github className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">仓库地址</span>
        </div>
        <Link
          href={systemInfo?.repo_url || "https://github.com/chengliang4810/jimuqu-devops.git"}
          target="_blank"
          rel="noopener noreferrer"
          className="text-right text-sm text-primary hover:underline"
        >
          chengliang4810/jimuqu-devops
        </Link>
      </div>

      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Tag className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">当前版本</span>
        </div>
        <code className="text-sm text-muted-foreground">{updateStatus?.current_version || systemInfo?.version || "dev"}</code>
      </div>

      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Tag className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">最新版本</span>
        </div>
        <code className="text-sm text-muted-foreground">{normalizeVersion(latestRelease?.tag_name)}</code>
      </div>

      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Tag className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">发布时间</span>
        </div>
        <span className="text-sm text-muted-foreground">{formatReleaseTime(latestRelease?.published_at)}</span>
      </div>

      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Tag className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">前端版本</span>
        </div>
        <code className="text-sm text-muted-foreground">{packageJson.version}</code>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <Button
          type="button"
          variant="outline"
          className="rounded-xl"
          onClick={() => void loadUpdateInfo(true)}
          disabled={loading || updating}
        >
          <RefreshCw className="mr-2 h-4 w-4" />
          检查更新
        </Button>
        <Button
          type="button"
          className="rounded-xl"
          onClick={() => void handleApplyUpdate()}
          disabled={!canApplyUpdate}
        >
          <Download className="mr-2 h-4 w-4" />
          {updating ? "更新中" : "立即更新"}
        </Button>
        <Button type="button" variant="ghost" className="rounded-xl" asChild>
          <Link
            href={latestRelease?.html_url || "https://github.com/chengliang4810/jimuqu-devops/releases"}
            target="_blank"
            rel="noopener noreferrer"
          >
            <ExternalLink className="mr-2 h-4 w-4" />
            查看 Release
          </Link>
        </Button>
      </div>

      {!updateStatus?.has_update ? (
        <p className="text-xs text-muted-foreground">
          {updateStatus?.current_version === "dev"
            ? "开发构建不支持在线更新，请使用发布版压缩包或 release 二进制。"
            : "当前已经是最新版本。"}
        </p>
      ) : null}
    </div>
  );
}
