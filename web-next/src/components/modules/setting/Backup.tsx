"use client";

import { useRef, useState } from "react";
import { Database, Download, Upload } from "lucide-react";
import { toast } from "sonner";
import { settingApi } from "@/api/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

function downloadJSON(data: unknown) {
  const content = JSON.stringify(data, null, 2);
  const blob = new Blob([content], { type: "application/json;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `jimuqu-backup-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, "-")}.json`;
  link.click();
  URL.revokeObjectURL(url);
}

export function SettingBackup() {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [exporting, setExporting] = useState(false);
  const [importing, setImporting] = useState(false);

  const handleExport = async () => {
    try {
      setExporting(true);
      const backup = await settingApi.exportBackup();
      downloadJSON(backup);
      toast.success("备份文件已下载");
    } catch (error: any) {
      toast.error(error.message || "下载备份失败");
    } finally {
      setExporting(false);
    }
  };

  const handleImport = async () => {
    if (!file) {
      toast.error("请先选择备份文件");
      return;
    }

    try {
      setImporting(true);
      const content = await file.text();
      const backup = JSON.parse(content);
      const result = await settingApi.importBackup(backup);
      toast.success(`导入完成：${Object.values(result.rows_affected).reduce((sum, count) => sum + count, 0)} 条`);
      fileInputRef.current && (fileInputRef.current.value = "");
      setFile(null);
      window.dispatchEvent(new CustomEvent("refresh-logs"));
      window.location.reload();
    } catch (error: any) {
      toast.error(error.message || "导入备份失败");
    } finally {
      setImporting(false);
    }
  };

  return (
    <div className="rounded-3xl border border-border bg-card p-6 space-y-5">
      <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
        <Database className="h-5 w-5" />
        备份 / 恢复
      </h2>

      <div className="space-y-3">
        <div className="text-sm font-medium text-card-foreground">下载备份</div>
        <p className="text-xs text-muted-foreground">
          会导出主机、项目、部署配置、通知渠道和设置数据。
        </p>
        <Button
          type="button"
          variant="outline"
          onClick={() => void handleExport()}
          disabled={exporting}
          className="w-full rounded-xl"
        >
          <Download className="h-4 w-4" />
          {exporting ? "导出中..." : "下载 JSON"}
        </Button>
      </div>

      <div className="h-px bg-border" />

      <div className="space-y-3">
        <div className="text-sm font-medium text-card-foreground">导入备份</div>
        <Input
          ref={fileInputRef}
          type="file"
          accept="application/json,.json"
          onChange={(event) => setFile(event.target.files?.[0] ?? null)}
          className="rounded-xl"
        />
        <Button
          type="button"
          variant="destructive"
          onClick={() => void handleImport()}
          disabled={importing || !file}
          className="w-full rounded-xl"
        >
          <Upload className="h-4 w-4" />
          {importing ? "导入中..." : "导入 JSON"}
        </Button>
      </div>
    </div>
  );
}
