"use client";

import { useEffect, useRef, useState } from "react";
import { Globe, Monitor } from "lucide-react";
import { Input } from "@/components/ui/input";
import type { SettingKey } from "@/types";

type SettingsMap = Record<SettingKey, string>;

type SettingSystemProps = {
  settings: SettingsMap;
  onSave: (key: SettingKey, value: string) => Promise<void>;
};

export function SettingSystem({ settings, onSave }: SettingSystemProps) {
  const [publicBaseURL, setPublicBaseURL] = useState(settings.public_base_url);
  const [proxyURL, setProxyURL] = useState(settings.proxy_url);
  const initialPublicBaseURL = useRef(settings.public_base_url);
  const initialProxyURL = useRef(settings.proxy_url);

  useEffect(() => {
    setPublicBaseURL(settings.public_base_url);
    setProxyURL(settings.proxy_url);
    initialPublicBaseURL.current = settings.public_base_url;
    initialProxyURL.current = settings.proxy_url;
  }, [settings]);

  return (
    <div className="rounded-3xl border border-border bg-card p-6 space-y-5">
      <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
        <Monitor className="h-5 w-5" />
        系统访问
      </h2>

      <div className="space-y-2">
        <div className="flex items-center gap-3">
          <Globe className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">对外访问地址</span>
        </div>
        <Input
          value={publicBaseURL}
          onChange={(event) => setPublicBaseURL(event.target.value)}
          onBlur={() => {
            if (publicBaseURL !== initialPublicBaseURL.current) {
              void onSave("public_base_url", publicBaseURL);
            }
          }}
          placeholder="https://devops.jimuqu.com"
          className="rounded-xl"
        />
        <p className="text-xs text-muted-foreground">
          用于通知中生成运行详情链接，例如 `https://your-domain.com/?view=logs&run_id=123`。
        </p>
      </div>

      <div className="space-y-2">
        <div className="flex items-center gap-3">
          <Globe className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">网络代理地址</span>
        </div>
        <Input
          value={proxyURL}
          onChange={(event) => setProxyURL(event.target.value)}
          onBlur={() => {
            if (proxyURL !== initialProxyURL.current) {
              void onSave("proxy_url", proxyURL);
            }
          }}
          placeholder="http://127.0.0.1:7890"
          className="rounded-xl"
        />
        <p className="text-xs text-muted-foreground">
          会自动写入 `HTTP_PROXY` / `HTTPS_PROXY` / `http_proxy` / `https_proxy`。
        </p>
      </div>
    </div>
  );
}
