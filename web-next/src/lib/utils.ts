import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function parseMultilineInput(value: string): string[] {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean);
}

export function formatMultilineValue(value?: string[] | null): string {
  return Array.isArray(value) ? value.join("\n") : "";
}

export function moveItemById<T extends { id: number }>(
  items: T[],
  activeId: number,
  overId: number
): T[] {
  if (activeId === overId) {
    return items;
  }

  const fromIndex = items.findIndex((item) => item.id === activeId);
  const toIndex = items.findIndex((item) => item.id === overId);
  if (fromIndex === -1 || toIndex === -1) {
    return items;
  }

  const nextItems = [...items];
  const [movedItem] = nextItems.splice(fromIndex, 1);
  nextItems.splice(toIndex, 0, movedItem);
  return nextItems;
}

// 格式化日期
export function formatDate(date: string | null): string {
  if (!date) return "-";
  const d = new Date(date);
  return d.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

// 格式化持续时间（秒）
export function formatDuration(seconds: number | null): string {
  if (seconds === null || seconds === 0) return "-";
  if (seconds < 60) return `${Math.round(seconds)}秒`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}分钟`;
}

// 计算持续时间（从开始到结束）
export function calculateDuration(startedAt: string | null, finishedAt: string | null): number | null {
  if (!startedAt) return null;
  const start = new Date(startedAt).getTime();
  const end = finishedAt ? new Date(finishedAt).getTime() : Date.now();
  return Math.floor((end - start) / 1000);
}

// 格式化简短日期时间
export function formatShortDateTime(date: string | null): string {
  if (!date) return "-";
  const d = new Date(date);
  return d.toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

// 状态 Badge variant 映射
export function getStatusVariant(status: string): "success" | "error" | "warning" | "secondary" | "default" {
  switch (status) {
    case "success":
      return "success";
    case "failed":
      return "error";
    case "running":
      return "warning";
    case "queued":
      return "secondary";
    default:
      return "default";
  }
}

// 状态文本映射
export function getStatusText(status: string): string {
  switch (status) {
    case "success":
      return "成功";
    case "failed":
      return "失败";
    case "running":
      return "运行中";
    case "queued":
      return "排队中";
    default:
      return status;
  }
}
