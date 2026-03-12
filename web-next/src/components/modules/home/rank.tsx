"use client";

import { TrendingUp } from "lucide-react";
import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import type { HomeProjectRank } from "@/types";
import { rankProjects, type RankSortMode } from "./data";

export function Rank({ projects }: { projects: HomeProjectRank[] }) {
  const [rankSortMode, setRankSortMode] = useState<RankSortMode>("deploys");

  const rankedByDeploys = useMemo(() => rankProjects(projects, "deploys"), [projects]);
  const rankedBySuccessRate = useMemo(() => rankProjects(projects, "successRate"), [projects]);
  const rankedByFailed = useMemo(() => rankProjects(projects, "failed"), [projects]);

  const getMedalEmoji = (rank: number): string => {
    switch (rank) {
      case 1:
        return "🥇";
      case 2:
        return "🥈";
      case 3:
        return "🥉";
      default:
        return "";
    }
  };

  const renderList = (items: HomeProjectRank[], mode: RankSortMode) => {
    if (items.length === 0) {
      return (
        <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
          <TrendingUp className="mb-3 h-12 w-12 opacity-30" />
          <p className="text-sm">暂无排行数据</p>
        </div>
      );
    }

    return (
      <div className="max-h-[300px] space-y-3 overflow-y-auto">
        {items.slice(0, 8).map((project, index) => {
          const rank = index + 1;
          const medal = getMedalEmoji(rank);

          return (
            <div
              key={project.project_id}
              className="flex items-center gap-3 rounded-2xl p-3 transition-colors hover:bg-accent/5"
            >
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-lg font-bold">
                {medal || rank}
              </div>

              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">{project.project_name}</p>
                {mode === "deploys" ? (
                  <div className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
                    <span>成功率:</span>
                    <span>{project.success_rate.toFixed(1)}%</span>
                  </div>
                ) : null}
                {mode !== "deploys" ? (
                  <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                    <span>{project.branch}</span>
                    <span>成功 {project.success_count}</span>
                    <span>失败 {project.failed_count}</span>
                  </div>
                ) : null}
              </div>

              <div className="shrink-0 text-right">
                {mode === "deploys" ? (
                  <span className="text-base font-semibold">{project.deploy_count} 次</span>
                ) : mode === "successRate" ? (
                  <span className="text-base font-semibold">
                    {project.success_rate.toFixed(1)}%
                  </span>
                ) : (
                  <div className="flex items-center gap-1 text-sm font-medium tabular-nums">
                    <span className="text-emerald-600">
                      {project.success_count}
                      <span className="text-xs text-muted-foreground"> 成功</span>
                    </span>
                    <span className="font-light text-muted-foreground/40">/</span>
                    <span className="text-destructive">
                      {project.failed_count}
                      <span className="text-xs text-muted-foreground"> 失败</span>
                    </span>
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div className="rounded-3xl border border-border/60 bg-card p-4 text-card-foreground">
      <div className="flex items-center justify-between">
        <h3 className="text-base font-semibold">项目排行</h3>
        <div className="flex gap-2">
          <Button
            size="sm"
            variant={rankSortMode === "deploys" ? "default" : "outline"}
            onClick={() => setRankSortMode("deploys")}
          >
            部署次数
          </Button>
          <Button
            size="sm"
            variant={rankSortMode === "successRate" ? "default" : "outline"}
            onClick={() => setRankSortMode("successRate")}
          >
            成功率
          </Button>
          <Button
            size="sm"
            variant={rankSortMode === "failed" ? "default" : "outline"}
            onClick={() => setRankSortMode("failed")}
          >
            成败比
          </Button>
        </div>
      </div>

      <div className="pt-3">
        {rankSortMode === "deploys" && renderList(rankedByDeploys, "deploys")}
        {rankSortMode === "successRate" && renderList(rankedBySuccessRate, "successRate")}
        {rankSortMode === "failed" && renderList(rankedByFailed, "failed")}
      </div>
    </div>
  );
}
