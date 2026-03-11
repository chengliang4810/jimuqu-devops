"use client";

import { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { statsApi } from "@/api/client";
import type { Stats } from "@/types";
import { FolderGit2, Server, History, Bell } from "lucide-react";

const statIcons = [FolderGit2, Server, History, Bell];
const statLabels = ["项目数", "主机数", "部署记录", "通知渠道"];
const statSubs = [
  "仓库分支即项目唯一键",
  "通过 SSH 账号密码管理",
  "包含构建与部署日志",
  "部署成功失败通知",
];

export function Home() {
  const [stats, setStats] = useState<Stats | null>(null);

  useEffect(() => {
    statsApi.get().then(setStats).catch(console.error);
  }, []);

  const statValues = stats
    ? [stats.project_count, stats.host_count, stats.run_count, stats.notify_channel_count]
    : [0, 0, 0, 0];

  const steps = [
    { step: 1, title: "添加主机", desc: "在「主机」页面添加目标部署服务器，配置SSH连接信息" },
    { step: 2, title: "创建项目", desc: "在「项目」页面添加Git仓库，配置编译和部署参数" },
    { step: 3, title: "配置Webhook", desc: "复制项目Webhook Token到Git仓库设置中实现自动触发" },
    { step: 4, title: "查看日志", desc: "在「部署记录」页面查看构建部署过程和结果" },
  ];

  return (
    <div className="space-y-6">
      {/* 统计卡片 */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {statLabels.map((label, index) => {
          const Icon = statIcons[index];
          return (
            <Card key={label} className="border-t-4 border-t-primary">
              <CardContent className="p-6">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-muted-foreground">{label}</span>
                  <Icon className="h-5 w-5 text-muted-foreground" />
                </div>
                <div className="text-3xl font-bold text-foreground mb-1">
                  {statValues[index]}
                </div>
                <p className="text-xs text-muted-foreground">{statSubs[index]}</p>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* 使用步骤 */}
      <Card>
        <CardContent className="p-6">
          <div className="space-y-4">
            {steps.map((item) => (
              <div
                key={item.step}
                className="flex items-start gap-4 p-4 rounded-lg border border-border hover:border-primary transition-colors"
              >
                <div className="flex-shrink-0 w-10 h-10 rounded-lg bg-primary text-primary-foreground flex items-center justify-center font-bold text-lg">
                  {item.step}
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-1">{item.title}</h4>
                  <p className="text-sm text-muted-foreground">{item.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
