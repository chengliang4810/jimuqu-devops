"use client";

import { motion } from "motion/react";
import {
  Activity,
  AlertTriangle,
  ChartColumnBig,
  CheckCircle2,
  Clock3,
  LoaderCircle,
  PauseCircle,
  TrendingUp,
} from "lucide-react";
import type { HomeStatsTotal } from "@/types";
import { AnimatedNumber } from "@/components/common/AnimatedNumber";
import { EASING } from "@/lib/animations/fluid-transitions";
import { formatDate } from "@/lib/utils";

type TotalProps = {
  total: HomeStatsTotal | null;
};

type TotalCardItem = {
  label: string;
  value: string | number | undefined;
  icon: typeof Activity;
  color: string;
  bgColor: string;
  unit: string;
  plain?: boolean;
};

type TotalCard = {
  title: string;
  headerIcon: typeof Activity;
  items: TotalCardItem[];
};

function formatAverageDuration(seconds: number | undefined): string {
  if (!seconds || seconds <= 0) {
    return "暂无";
  }

  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainSeconds = seconds % 60;

  if (hours > 0) {
    return `${hours}小时 ${minutes}分`;
  }
  if (minutes > 0) {
    return `${minutes}分 ${remainSeconds}秒`;
  }
  return `${remainSeconds}秒`;
}

export function Total({ total }: TotalProps) {
  const cards: TotalCard[] = [
    {
      title: "部署概览",
      headerIcon: Activity,
      items: [
        {
          label: "部署总数",
          value: total?.deploy_count,
          icon: Activity,
          color: "text-primary",
          bgColor: "bg-primary/10",
          unit: "次",
          plain: false,
        },
        {
          label: "平均部署次数",
          value: total?.average_deploy_per_project.toFixed(1),
          icon: TrendingUp,
          color: "text-primary",
          bgColor: "bg-accent/10",
          unit: "次/项目",
          plain: false,
        },
      ],
    },
    {
      title: "结果统计",
      headerIcon: ChartColumnBig,
      items: [
        {
          label: "成功次数",
          value: total?.success_count,
          icon: CheckCircle2,
          color: "text-primary",
          bgColor: "bg-chart-1/10",
          unit: "次",
          plain: false,
        },
        {
          label: "失败次数",
          value: total?.failed_count,
          icon: AlertTriangle,
          color: "text-primary",
          bgColor: "bg-chart-2/10",
          unit: "次",
          plain: false,
        },
      ],
    },
    {
      title: "执行状态",
      headerIcon: LoaderCircle,
      items: [
        {
          label: "运行中",
          value: total?.running_count,
          icon: LoaderCircle,
          color: "text-primary",
          bgColor: "bg-chart-3/10",
          unit: "次",
          plain: false,
        },
        {
          label: "平均耗时",
          value: formatAverageDuration(total?.average_deploy_duration_seconds),
          icon: Clock3,
          color: "text-primary",
          bgColor: "bg-chart-3/10",
          unit: "",
          plain: true,
        },
      ],
    },
    {
      title: "稳定性",
      headerIcon: Clock3,
      items: [
        {
          label: "成功率",
          value: total?.success_rate.toFixed(1),
          icon: CheckCircle2,
          color: "text-primary",
          bgColor: "bg-chart-4/10",
          unit: "%",
          plain: false,
        },
        {
          label: "最近部署",
          value: total?.last_deploy_at ? formatDate(total.last_deploy_at) : "暂无",
          icon: Clock3,
          color: "text-primary",
          bgColor: "bg-chart-4/10",
          unit: "",
          plain: true,
        },
      ],
    },
  ];

  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
      {cards.map((card, index) => (
        <motion.section
          key={index}
          className="flex flex-row items-center gap-4 rounded-3xl border border-border/60 bg-card p-5 text-card-foreground"
          initial={{ opacity: 0, y: 20, filter: "blur(8px)" }}
          animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
          transition={{
            duration: 0.5,
            ease: EASING.easeOutExpo,
            delay: index * 0.08,
          }}
        >
          <div className="flex self-stretch border-r border-border/50 px-0 py-1 pr-4">
            <div className="flex flex-col items-center justify-center gap-3">
              <card.headerIcon className="h-4 w-4" />
              <h3 className="text-sm font-medium [writing-mode:vertical-lr]">{card.title}</h3>
            </div>
          </div>

          <div className="flex min-w-0 flex-1 flex-col gap-4">
            {card.items.map((item, itemIndex) => (
              <div key={itemIndex} className="flex items-center gap-3">
                <div
                  className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${item.bgColor} ${item.color}`}
                >
                  <item.icon className="h-5 w-5" />
                </div>
                <div className="flex min-w-0 flex-col">
                  <span className="text-xs text-muted-foreground">{item.label}</span>
                  {item.plain ? (
                    <div className="text-sm leading-5">
                      {String(item.value)
                        .split(" ")
                        .map((part, partIndex) => (
                          <div key={partIndex} className="whitespace-nowrap">
                            {part}
                          </div>
                        ))}
                    </div>
                  ) : (
                    <div className="flex items-baseline gap-1 whitespace-nowrap">
                      <span className="text-xl whitespace-nowrap">
                        <AnimatedNumber value={item.value} />
                      </span>
                      {item.unit ? (
                        <span className="whitespace-nowrap text-sm text-muted-foreground">{item.unit}</span>
                      ) : null}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </motion.section>
      ))}
    </div>
  );
}
