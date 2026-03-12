"use client";

import dayjs from "dayjs";
import { useMemo, useState } from "react";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import type { HomeStatsDaily, HomeStatsHourly } from "@/types";
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart";
import { AnimatedNumber } from "@/components/common/AnimatedNumber";
import { Button } from "@/components/ui/button";

type ChartMetricType = "deploys" | "successRate" | "failed";
type ChartPeriod = "1" | "7" | "30";

const PERIODS: readonly ChartPeriod[] = ["1", "7", "30"];

export function StatsChart({
  daily,
  hourly,
}: {
  daily: HomeStatsDaily[];
  hourly: HomeStatsHourly[];
}) {
  const [chartMetricType, setChartMetricType] = useState<ChartMetricType>("deploys");
  const [period, setChartPeriod] = useState<ChartPeriod>("7");

  const sortedDaily = useMemo(
    () => [...daily].sort((left, right) => left.date.localeCompare(right.date)),
    [daily]
  );

  const getChartDataKey = (type: ChartMetricType) => {
    return type === "successRate" ? "success_rate" : type === "failed" ? "failed_count" : "deploy_count";
  };

  const chartData = useMemo(() => {
    const dataKey = getChartDataKey(chartMetricType);
    if (period === "1") {
      return hourly.map((stat) => ({
        date: `${stat.hour}:00`,
        [dataKey]:
          chartMetricType === "successRate"
            ? stat.success_rate
            : chartMetricType === "failed"
              ? stat.failed_count
              : stat.deploy_count,
      }));
    }

    const days = Number(period);
    return sortedDaily.slice(-days).map((stat) => ({
      date: dayjs(stat.date).format("MM/DD"),
      [dataKey]:
        chartMetricType === "successRate"
          ? stat.success_rate
          : chartMetricType === "failed"
            ? stat.failed_count
            : stat.deploy_count,
    }));
  }, [chartMetricType, hourly, period, sortedDaily]);

  const totals = useMemo(() => {
    const source =
      period === "1"
        ? hourly.map((item) => ({
            deploy_count: item.deploy_count,
            success_count: item.success_count,
            failed_count: item.failed_count,
            success_rate: item.success_rate,
          }))
        : sortedDaily.slice(-Number(period));

    const deploys = source.reduce((sum, item) => sum + item.deploy_count, 0);
    const success = source.reduce((sum, item) => sum + item.success_count, 0);
    const failed = source.reduce((sum, item) => sum + item.failed_count, 0);
    const successRate = success + failed > 0 ? (success * 100) / (success + failed) : 0;
    return { deploys, success, failed, successRate };
  }, [hourly, period, sortedDaily]);

  const chartConfig = useMemo(() => {
    const dataKey = getChartDataKey(chartMetricType);
    const labels = {
      deploy_count: "部署次数",
      success_rate: "成功率",
      failed_count: "失败次数",
    };
    return {
      [dataKey]: { label: labels[dataKey], color: "var(--primary)" },
    };
  }, [chartMetricType]);

  const getPeriodLabel = (value: ChartPeriod) => {
    if (value === "1") return "今天";
    if (value === "7") return "最近 7 天";
    return "最近 30 天";
  };

  const handlePeriodClick = () => {
    const currentIndex = PERIODS.indexOf(period);
    const nextIndex = (currentIndex + 1) % PERIODS.length;
    setChartPeriod(PERIODS[nextIndex]);
  };

  const getChartStroke = (type: ChartMetricType) => {
    if (type === "successRate") return "var(--chart-2)";
    if (type === "failed") return "var(--chart-3)";
    return "var(--chart-1)";
  };

  const getChartFill = (type: ChartMetricType) => {
    if (type === "successRate") return "url(#fillMetric2)";
    if (type === "failed") return "url(#fillMetric3)";
    return "url(#fillMetric1)";
  };

  return (
    <div className="rounded-3xl border border-border/60 bg-card pb-0 pt-4 text-card-foreground custom-shadow">
      <div className="space-y-2 px-4 pb-2">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-base font-semibold">部署趋势</h3>
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant={chartMetricType === "deploys" ? "default" : "outline"}
              onClick={() => setChartMetricType("deploys")}
            >
              部署次数
            </Button>
            <Button
              size="sm"
              variant={chartMetricType === "successRate" ? "default" : "outline"}
              onClick={() => setChartMetricType("successRate")}
            >
              成功率
            </Button>
            <Button
              size="sm"
              variant={chartMetricType === "failed" ? "default" : "outline"}
              onClick={() => setChartMetricType("failed")}
            >
              失败次数
            </Button>
          </div>
        </div>

        <div className="flex items-start justify-between">
          <div className="flex gap-2 text-sm">
            <div>
              <div className="text-xs text-muted-foreground">部署总数</div>
              <div className="text-xl font-semibold">
                <AnimatedNumber value={totals.deploys} />
                <span className="ml-0.5 text-sm text-muted-foreground">次</span>
              </div>
            </div>
            <div className="w-px self-stretch bg-border" />
            <div>
              <div className="text-xs text-muted-foreground">成功次数</div>
              <div className="text-xl font-semibold">
                <AnimatedNumber value={totals.success} />
                <span className="ml-0.5 text-sm text-muted-foreground">次</span>
              </div>
            </div>
            <div className="w-px self-stretch bg-border" />
            <div>
              <div className="text-xs text-muted-foreground">失败次数</div>
              <div className="text-xl font-semibold">
                <AnimatedNumber value={totals.failed} />
                <span className="ml-0.5 text-sm text-muted-foreground">次</span>
              </div>
            </div>
          </div>
          <div
            className="cursor-pointer text-sm transition-opacity hover:opacity-80"
            onClick={handlePeriodClick}
          >
            <div className="text-xs text-muted-foreground">时间范围</div>
            <div className="text-base font-semibold">{getPeriodLabel(period)}</div>
          </div>
        </div>
      </div>

      <ChartContainer config={chartConfig} className="h-40 w-full">
        <AreaChart accessibilityLayer data={chartData}>
          <defs>
            <linearGradient id="fillMetric1" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="var(--chart-1)" stopOpacity={1} />
              <stop offset="95%" stopColor="var(--chart-1)" stopOpacity={0.1} />
            </linearGradient>
            <linearGradient id="fillMetric2" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="var(--chart-2)" stopOpacity={1} />
              <stop offset="95%" stopColor="var(--chart-2)" stopOpacity={0.1} />
            </linearGradient>
            <linearGradient id="fillMetric3" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="var(--chart-3)" stopOpacity={1} />
              <stop offset="95%" stopColor="var(--chart-3)" stopOpacity={0.1} />
            </linearGradient>
          </defs>
          <CartesianGrid vertical={false} strokeDasharray="3 3" />
          <XAxis dataKey="date" tickLine={false} axisLine={false} />
          <YAxis
            tickLine={false}
            axisLine={false}
            tickFormatter={(value) =>
              chartMetricType === "successRate" ? `${Number(value).toFixed(0)}%` : `${value}`
            }
          />
          <ChartTooltip
            cursor={false}
            content={<ChartTooltipContent indicator="line" formatter={(value: number | string, name: string) => (
              <div className="flex flex-1 items-center justify-between gap-4 leading-none">
                <span className="text-muted-foreground">{name}</span>
                <span className="font-mono font-medium tabular-nums text-foreground">
                  {chartMetricType === "successRate"
                    ? `${Number(value).toFixed(1)}%`
                    : Number(value).toLocaleString()}
                </span>
              </div>
            )} />}
          />
          <Area
            type="monotone"
            dataKey={getChartDataKey(chartMetricType)}
            stroke={getChartStroke(chartMetricType)}
            fill={getChartFill(chartMetricType)}
          />
        </AreaChart>
      </ChartContainer>
    </div>
  );
}
