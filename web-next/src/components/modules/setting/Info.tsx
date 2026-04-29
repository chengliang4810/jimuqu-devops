"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { runApi, settingApi } from "@/api/client";
import { getApiOrigin } from "@/lib/api-base";
import { toast } from "sonner";
import { Download, Github, Info, Tag } from "lucide-react";
import type { ReleaseInfo, SystemInfo, UpdateStatus } from "@/types";

type SettingInfoProps = {
  systemInfo: SystemInfo | null;
};

function normalizeVersion(value?: string) {
  if (!value) return "-";
  return value.startsWith("v") ? value.slice(1) : value;
}

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

export function SettingInfo({ systemInfo }: SettingInfoProps) {
  const [latestRelease, setLatestRelease] = useState<ReleaseInfo | null>(null);
  const [updateStatus, setUpdateStatus] = useState<UpdateStatus | null>(null);
  const [updating, setUpdating] = useState(false);

  const loadUpdateInfo = async (showToast = false) => {
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
    }
  };

  useEffect(() => {
    void loadUpdateInfo();
  }, []);

  const canApplyUpdate = useMemo(() => {
    return !!updateStatus?.has_update && updateStatus?.can_update !== false && !updating;
  }, [updateStatus?.can_update, updateStatus?.has_update, updating]);

  const waitForServiceRecovery = async () => {
    await sleep(1500);
    const deadline = Date.now() + 120000;

    while (Date.now() < deadline) {
      try {
        const response = await fetch(`${getApiOrigin()}/healthz`, {
          method: "GET",
          cache: "no-store",
        });
        if (response.ok) {
          window.location.reload();
          return true;
        }
      } catch {
        // 服务重启期间请求失败是预期行为，继续轮询即可。
      }

      await sleep(2000);
    }

    return false;
  };

  const handleApplyUpdate = async () => {
    if (!canApplyUpdate) return;
    setUpdating(true);
    try {
      const runs = await runApi.list({ limit: 200 });
      const hasActiveRun = runs.some((run) => run.status === "queued" || run.status === "running");
      if (hasActiveRun) {
        toast.error("存在正在部署中的任务，请等待部署完成后再更新");
        setUpdating(false);
        return;
      }

      const result = await settingApi.applyUpdate();
      toast.success(result.message || "更新已应用，应用即将自动重启");

      const recovered = await waitForServiceRecovery();
      if (!recovered) {
        toast.error("服务正在重启，请稍后手动刷新页面确认版本");
        setUpdating(false);
      }
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
        <Link
          href={latestRelease?.html_url || "https://github.com/chengliang4810/jimuqu-devops/releases"}
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm text-primary hover:underline"
        >
          <code>{normalizeVersion(latestRelease?.tag_name)}</code>
        </Link>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        {updateStatus?.has_update ? (
          <Button
            type="button"
            className="rounded-xl"
            onClick={() => void handleApplyUpdate()}
            disabled={!canApplyUpdate}
          >
            <Download className="mr-2 h-4 w-4" />
            {updating ? "等待重启" : "立即更新"}
          </Button>
        ) : null}
      </div>

      {!updateStatus?.has_update && updateStatus?.current_version === "dev" ? (
        <p className="text-xs text-muted-foreground">
          开发构建不支持在线更新，请使用发布版压缩包或 release 二进制。
        </p>
      ) : null}

      {updateStatus?.has_update && updateStatus.can_update === false ? (
        <p className="text-xs text-muted-foreground">
          Docker 部署不支持在线替换程序，请拉取最新镜像并重建容器。
        </p>
      ) : null}
    </div>
  );
}
