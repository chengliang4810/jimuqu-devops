"use client";

import { useEffect, useRef, useState } from "react";
import { Globe, Monitor } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import type { SettingKey } from "@/types";

type SettingsMap = Record<SettingKey, string>;

type SettingSystemProps = {
  settings: SettingsMap;
  onSave: (key: SettingKey, value: string) => Promise<void>;
};

export function SettingSystem({ settings, onSave }: SettingSystemProps) {
  const [mirrorURL, setMirrorURL] = useState(settings.docker_mirror_url);
  const [gitDockerImage, setGitDockerImage] = useState(settings.git_docker_image);
  const [publicBaseURL, setPublicBaseURL] = useState(settings.public_base_url);
  const [proxyURL, setProxyURL] = useState(settings.proxy_url);
  const initialMirrorURL = useRef(settings.docker_mirror_url);
  const initialGitDockerImage = useRef(settings.git_docker_image);
  const initialPublicBaseURL = useRef(settings.public_base_url);
  const initialProxyURL = useRef(settings.proxy_url);

  useEffect(() => {
    setMirrorURL(settings.docker_mirror_url);
    setGitDockerImage(settings.git_docker_image);
    setPublicBaseURL(settings.public_base_url);
    setProxyURL(settings.proxy_url);
    initialMirrorURL.current = settings.docker_mirror_url;
    initialGitDockerImage.current = settings.git_docker_image;
    initialPublicBaseURL.current = settings.public_base_url;
    initialProxyURL.current = settings.proxy_url;
  }, [settings]);

  return (
    <div className="rounded-3xl border border-border bg-card p-6 space-y-5">
      <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
        <Monitor className="h-5 w-5" />
        系统设置
      </h2>

      <div className="space-y-2">
        <div className="flex items-center gap-3">
          <Globe className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">镜像加速地址</span>
        </div>
        <Textarea
          value={mirrorURL}
          onChange={(event) => setMirrorURL(event.target.value)}
          onBlur={() => {
            if (mirrorURL !== initialMirrorURL.current) {
              void onSave("docker_mirror_url", mirrorURL);
            }
          }}
          placeholder={"每行一个镜像加速地址，例如\nmirror.ccs.tencentyun.com\ndocker.1ms.run"}
          className="rounded-xl"
          rows={4}
        />
        <p className="text-xs text-muted-foreground">
          用于构建阶段的 Docker 镜像地址前缀，每行一个，系统会按顺序依次尝试。
        </p>
      </div>

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

      <div className="space-y-2">
        <div className="flex items-center gap-3">
          <Monitor className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">Git Docker 镜像</span>
        </div>
        <Input
          value={gitDockerImage}
          onChange={(event) => setGitDockerImage(event.target.value)}
          onBlur={() => {
            if (gitDockerImage !== initialGitDockerImage.current) {
              void onSave("git_docker_image", gitDockerImage);
            }
          }}
          placeholder="alpine/git:latest"
          className="rounded-xl"
        />
        <p className="text-xs text-muted-foreground">
          拉取代码阶段会通过 `docker run` 使用这个镜像执行 `git clone`，宿主机无需安装 Git。
        </p>
      </div>
    </div>
  );
}
