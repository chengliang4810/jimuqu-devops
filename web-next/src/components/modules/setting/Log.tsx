"use client";

import { useEffect, useRef, useState } from "react";
import { Calendar, ScrollText, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { runApi } from "@/api/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { SettingKey } from "@/types";

type SettingsMap = Record<SettingKey, string>;

type SettingLogProps = {
  settings: SettingsMap;
  onSave: (key: SettingKey, value: string) => Promise<void>;
};

export function SettingLog({ settings, onSave }: SettingLogProps) {
  const [retentionDays, setRetentionDays] = useState(settings.run_retention_days);
  const [clearing, setClearing] = useState(false);
  const initialRetentionDays = useRef(settings.run_retention_days);

  useEffect(() => {
    setRetentionDays(settings.run_retention_days);
    initialRetentionDays.current = settings.run_retention_days;
  }, [settings.run_retention_days]);

  const handleClearLogs = async () => {
    if (!window.confirm("确认清空所有部署记录吗？此操作不可恢复。")) {
      return;
    }

    try {
      setClearing(true);
      const result = await runApi.clear();
      toast.success(`已清空 ${result.cleared} 条部署记录`);
      window.dispatchEvent(new CustomEvent("refresh-logs"));
    } catch (error: any) {
      toast.error(error.message || "清空部署记录失败");
    } finally {
      setClearing(false);
    }
  };

  return (
    <div className="rounded-3xl border border-border bg-card p-6 space-y-5">
      <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
        <ScrollText className="h-5 w-5" />
        部署记录设置
      </h2>

      <div className="space-y-2">
        <div className="flex items-center gap-3">
          <Calendar className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">记录保存天数</span>
        </div>
        <Input
          type="number"
          min={1}
          value={retentionDays}
          onChange={(event) => setRetentionDays(event.target.value)}
          onBlur={() => {
            if (retentionDays !== initialRetentionDays.current) {
              void onSave("run_retention_days", retentionDays);
            }
          }}
          className="rounded-xl"
        />
        <p className="text-xs text-muted-foreground">
          保存后会立即清理超过保留天数的历史部署记录。
        </p>
      </div>

      <div className="space-y-2">
        <div className="flex items-center gap-3">
          <Trash2 className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm font-medium">清空部署记录</span>
        </div>
        <Button
          type="button"
          variant="destructive"
          onClick={() => void handleClearLogs()}
          disabled={clearing}
          className="w-full rounded-xl"
        >
          {clearing ? "清空中..." : "清空记录"}
        </Button>
      </div>
    </div>
  );
}
