"use client";

import dayjs from "dayjs";
import { Fragment, useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import type { HomeStatsDaily } from "@/types";

type StatsDailyData = {
  dateStr: string;
  isFuture: boolean;
  stat: HomeStatsDaily | null;
};

const ACTIVITY_LEVELS = [
  { min: 20, level: 4 },
  { min: 10, level: 3 },
  { min: 4, level: 2 },
  { min: 1, level: 1 },
];

function getActivityLevel(value: number): number {
  if (value === 0) return 0;
  return ACTIVITY_LEVELS.find((level) => value >= level.min)?.level || 1;
}

export function Activity({ daily }: { daily: HomeStatsDaily[] }) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [tooltip, setTooltip] = useState<{
    day: StatsDailyData;
    x: number;
    y: number;
    visible: boolean;
  } | null>(null);

  const days = useMemo(() => {
    const formattedMap = new Map(daily.map((stat) => [dayjs(stat.date).format("YYYYMMDD"), stat]));
    const today = dayjs();
    const startDate = today.subtract(today.day() + 53 * 7, "day");
    const result: StatsDailyData[] = [];

    for (let index = 0; index < 54 * 7; index += 1) {
      const currentDate = startDate.add(index, "day");
      const dateStr = currentDate.format("YYYYMMDD");

      result.push({
        dateStr,
        isFuture: currentDate.isAfter(today, "day"),
        stat: formattedMap.get(dateStr) || null,
      });
    }

    return result;
  }, [daily]);

  const [maskImage, setMaskImage] = useState("none");

  const checkScroll = useCallback(() => {
    if (!scrollRef.current) return;
    const { scrollLeft, scrollWidth, clientWidth } = scrollRef.current;
    const isStart = scrollLeft <= 1;
    const isEnd = Math.abs(scrollWidth - clientWidth - scrollLeft) <= 1;

    if (isStart && isEnd) {
      setMaskImage("none");
    } else if (isStart) {
      setMaskImage("linear-gradient(to left, transparent, rgba(0,0,0,0) 10px, black 40px)");
    } else if (isEnd) {
      setMaskImage("linear-gradient(to right, transparent, rgba(0,0,0,0) 10px, black 40px)");
    } else {
      setMaskImage(
        "linear-gradient(to right, transparent, rgba(0,0,0,0) 10px, black 40px, black calc(100% - 40px), rgba(0,0,0,0) calc(100% - 10px), transparent)"
      );
    }
  }, []);

  useLayoutEffect(() => {
    const scrollToRight = () => {
      if (scrollRef.current) {
        scrollRef.current.scrollLeft = scrollRef.current.scrollWidth;
        checkScroll();
      }
    };

    scrollToRight();
    window.addEventListener("resize", scrollToRight);
    return () => window.removeEventListener("resize", scrollToRight);
  }, [days, checkScroll]);

  return (
    <div className="rounded-3xl border border-border/60 bg-card text-card-foreground custom-shadow">
      <div
        ref={scrollRef}
        onScroll={checkScroll}
        className="overflow-x-auto p-4"
        style={{ maskImage, WebkitMaskImage: maskImage }}
      >
        <div className="ml-auto w-fit">
          <div
            className="grid gap-1"
            style={{
              gridTemplateColumns: "repeat(54, 0.875rem)",
              gridTemplateRows: "repeat(7, 0.875rem)",
              gridAutoFlow: "column",
            }}
          >
            {days.map((day) => {
              if (day.isFuture) {
                return <div key={day.dateStr} />;
              }

              const level = getActivityLevel(day.stat?.deploy_count ?? 0);
              return (
                <div
                  key={day.dateStr}
                  className="cursor-pointer rounded-sm transition-all hover:scale-150"
                  onMouseEnter={(event) => {
                    const rect = event.currentTarget.getBoundingClientRect();
                    setTooltip({
                      day,
                      x: rect.left + rect.width / 2,
                      y: rect.top,
                      visible: true,
                    });
                  }}
                  onMouseLeave={() =>
                    setTooltip((prev) => (prev ? { ...prev, visible: false } : null))
                  }
                  style={{
                    backgroundColor:
                      level === 0
                        ? "var(--muted)"
                        : `color-mix(in oklch, var(--primary) ${level * 25}%, var(--muted))`,
                  }}
                />
              );
            })}
          </div>
        </div>
      </div>

      {tooltip && typeof document !== "undefined"
        ? createPortal(
            (() => {
              const isLeft = tooltip.x < 200;
              const isRight = tooltip.x > window.innerWidth - 200;
              const isTop = tooltip.y < window.innerHeight / 2;
              const tooltipDate = dayjs(tooltip.day.dateStr, "YYYYMMDD");
              const tooltipDateLabel = tooltipDate.isValid()
                ? tooltipDate.format("YYYY-MM-DD")
                : tooltip.day.dateStr;

              let transform = "translate(-50%, 15%)";
              if (!isTop && !isLeft && !isRight) {
                transform = "translate(-50%, -105%)";
              } else if (isTop && isLeft) {
                transform = "translate(10%, 15%)";
              } else if (isTop && isRight) {
                transform = "translate(-110%, 15%)";
              } else if (!isTop && isLeft) {
                transform = "translate(10%, -105%)";
              } else if (!isTop && isRight) {
                transform = "translate(-110%, -105%)";
              }

              return (
                <div
                  className={`pointer-events-none fixed z-50 w-fit min-w-max rounded-3xl border bg-background p-3 text-sm text-foreground transition-opacity duration-500 ${
                    tooltip.visible ? "opacity-100" : "opacity-0"
                  }`}
                  style={{
                    left: tooltip.x,
                    top: tooltip.y,
                    transform,
                  }}
                >
                  <div className="space-y-2">
                    <p className="font-semibold text-foreground">{tooltipDateLabel}</p>
                    {tooltip.day.stat ? (
                      <div className="grid grid-cols-[auto_1fr] items-center gap-x-4 gap-y-1 text-muted-foreground">
                        {[
                          ["部署次数", tooltip.day.stat.deploy_count],
                          ["成功次数", tooltip.day.stat.success_count],
                          ["失败次数", tooltip.day.stat.failed_count],
                          ["成功率", `${tooltip.day.stat.success_rate.toFixed(1)}%`],
                        ].map(([label, value], index) => (
                          <Fragment key={index}>
                            <span>{label}</span>
                            <span className="text-right font-medium text-foreground">{value}</span>
                          </Fragment>
                        ))}
                      </div>
                    ) : (
                      <p className="text-muted-foreground">当天暂无部署数据</p>
                    )}
                  </div>
                </div>
              );
            })(),
            document.body
          )
        : null}
    </div>
  );
}
