"use client";

import Link from "next/link";
import packageJson from "../../../../package.json";
import { Github, Info, Tag } from "lucide-react";
import type { SystemInfo } from "@/types";

type SettingInfoProps = {
  systemInfo: SystemInfo | null;
};

export function SettingInfo({ systemInfo }: SettingInfoProps) {
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
        <code className="text-sm text-muted-foreground">{systemInfo?.version || "dev"}</code>
      </div>

      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Tag className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">前端版本</span>
        </div>
        <code className="text-sm text-muted-foreground">{packageJson.version}</code>
      </div>
    </div>
  );
}
