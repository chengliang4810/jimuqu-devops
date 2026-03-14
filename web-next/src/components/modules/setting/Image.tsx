"use client";

import { useEffect, useRef, useState } from "react";
import { Globe, Package } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import type { SettingKey } from "@/types";

type SettingsMap = Record<SettingKey, string>;

type SettingImageProps = {
  settings: SettingsMap;
  onSave: (key: SettingKey, value: string) => Promise<void>;
};

export function SettingImage({ settings, onSave }: SettingImageProps) {
  const [mirrorURL, setMirrorURL] = useState(settings.docker_mirror_url);
  const [gitDockerImage, setGitDockerImage] = useState(settings.git_docker_image);
  const initialMirrorURL = useRef(settings.docker_mirror_url);
  const initialGitDockerImage = useRef(settings.git_docker_image);

  useEffect(() => {
    setMirrorURL(settings.docker_mirror_url);
    setGitDockerImage(settings.git_docker_image);
    initialMirrorURL.current = settings.docker_mirror_url;
    initialGitDockerImage.current = settings.git_docker_image;
  }, [settings]);

  return (
    <div className="space-y-5 rounded-3xl border border-border bg-card p-6">
      <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
        <Package className="h-5 w-5" />
        镜像设置
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
          用于构建阶段和 Git 拉取阶段的 Docker 镜像地址前缀，系统会按顺序依次尝试。
        </p>
      </div>

      <div className="space-y-2">
        <div className="flex items-center gap-3">
          <Package className="h-5 w-5 text-muted-foreground" />
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
          拉取代码阶段会通过 `docker run` 使用这个镜像执行 `git clone`。
        </p>
      </div>
    </div>
  );
}
