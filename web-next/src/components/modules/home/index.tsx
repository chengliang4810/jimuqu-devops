"use client";

import { useEffect, useState } from "react";
import { Bell, FolderGit2, History, Server } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { homeApi, statsApi } from "@/api/client";
import type { HomeDashboard, Stats } from "@/types";
import { useNavStore } from "@/stores";
import { Activity } from "./activity";
import { StatsChart } from "./chart";
import { Rank } from "./rank";
import { Total } from "./total";
import { PageWrapper } from "@/components/common/PageWrapper";

const statIcons = [FolderGit2, Server, Bell, History];
const statLabels = ["项目数", "主机数", "通知渠道", "部署记录"];
const statSubs = [
  "仓库与分支对应部署对象",
  "通过 SSH 连接目标环境",
  "统一管理部署通知出口",
  "查看每次部署的过程与结果",
];
const statViews = ["projects", "hosts", "notifications", "logs"] as const;
const statTooltips = ["跳转到项目管理", "跳转到主机管理", "跳转到通知渠道", "跳转到部署记录"];

export function Home() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [dashboard, setDashboard] = useState<HomeDashboard | null>(null);
  const { setActiveView } = useNavStore();

  useEffect(() => {
    Promise.all([statsApi.get(), homeApi.getDashboard()])
      .then(([nextStats, nextDashboard]) => {
        setStats(nextStats);
        setDashboard(nextDashboard);
      })
      .catch(console.error);
  }, []);
  const navValues = [
    stats?.project_count ?? dashboard?.total.project_count ?? 0,
    stats?.host_count ?? 0,
    stats?.notify_channel_count ?? 0,
    stats?.run_count ?? dashboard?.total.deploy_count ?? 0,
  ];

  return (
    <PageWrapper className="space-y-6 pb-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {statLabels.map((label, index) => {
          const Icon = statIcons[index];
          return (
            <Card
              key={label}
              className="border-t-4 border-t-primary transition-all hover:-translate-y-0.5 hover:border-primary/60 hover:shadow-md"
            >
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    type="button"
                    className="block w-full cursor-pointer text-left"
                    aria-label={statTooltips[index]}
                    onClick={() => setActiveView(statViews[index])}
                  >
                    <CardContent className="p-6">
                      <div className="mb-2 flex items-center justify-between">
                        <span className="text-sm font-medium text-muted-foreground">{label}</span>
                        <Icon className="h-5 w-5 text-muted-foreground" />
                      </div>
                      <div className="mb-1 text-3xl font-bold text-foreground">
                        {navValues[index]}
                      </div>
                      <p className="text-xs text-muted-foreground">{statSubs[index]}</p>
                    </CardContent>
                  </button>
                </TooltipTrigger>
                <TooltipContent>{statTooltips[index]}</TooltipContent>
              </Tooltip>
            </Card>
          );
        })}
      </div>

      <Total total={dashboard?.total ?? null} />
      <Activity daily={dashboard?.daily ?? []} />

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(0,0.95fr)]">
        <StatsChart daily={dashboard?.daily ?? []} hourly={dashboard?.hourly ?? []} />
        <Rank projects={dashboard?.projects ?? []} />
      </div>
    </PageWrapper>
  );
}
