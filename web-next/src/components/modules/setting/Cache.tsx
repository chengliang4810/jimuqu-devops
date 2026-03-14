"use client";

import { useEffect, useRef, useState } from "react";
import { DatabaseZap } from "lucide-react";
import { Textarea } from "@/components/ui/textarea";
import type { SettingKey } from "@/types";

type SettingsMap = Record<SettingKey, string>;

type SettingCacheProps = {
  settings: SettingsMap;
  onSave: (key: SettingKey, value: string) => Promise<void>;
};

const DEFAULT_CACHE_DIRS = [
  "/root/.m2",
  "/root/.gradle/caches",
  "/root/.npm",
  "/root/.yarn",
  "/go/pkg/mod",
  "/root/.cache",
].join("\n");

export function SettingCache({ settings, onSave }: SettingCacheProps) {
  const [cacheDirs, setCacheDirs] = useState(settings.build_cache_dirs);
  const initialCacheDirs = useRef(settings.build_cache_dirs);

  useEffect(() => {
    setCacheDirs(settings.build_cache_dirs);
    initialCacheDirs.current = settings.build_cache_dirs;
  }, [settings]);

  return (
    <div className="space-y-5 rounded-3xl border border-border bg-card p-6">
      <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
        <DatabaseZap className="h-5 w-5" />
        缓存设置
      </h2>

      <div className="space-y-2">
        <Textarea
          value={cacheDirs}
          onChange={(event) => setCacheDirs(event.target.value)}
          onBlur={() => {
            if (cacheDirs !== initialCacheDirs.current) {
              void onSave("build_cache_dirs", cacheDirs);
            }
          }}
          placeholder={DEFAULT_CACHE_DIRS}
          className="rounded-xl"
          rows={7}
        />
        <p className="text-xs text-muted-foreground">
          每行一个容器内绝对路径。所有项目构建都会统一挂载到服务端 `APP_DATA_DIR/cache` 下对应目录。
        </p>
      </div>
    </div>
  );
}
